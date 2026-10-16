package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/infrai-examples/logistics-presence/internal/infrai"
	"github.com/infrai-examples/logistics-presence/internal/logistics"
)

type server struct {
	client *infrai.Client
}

func main() {
	apiKey := os.Getenv("INFRAI_API_KEY")
	if apiKey == "" {
		log.Fatal("INFRAI_API_KEY is required")
	}
	s := &server{client: infrai.NewClient(apiKey)}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /workspace/{channel}", s.workspace)
	mux.HandleFunc("POST /workspace/{channel}/shipments/{shipment}/events", s.shipmentEvent)
	addr := envOr("LISTEN_ADDR", ":8080")
	log.Printf("logistics presence listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func (s *server) workspace(w http.ResponseWriter, r *http.Request) {
	workspace, err := logistics.LoadWorkspace(r.Context(), s.client, r.PathValue("channel"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, workspace)
}

type eventInput struct {
	EventID   string                     `json:"event_id"`
	Kind      string                     `json:"kind"`
	Occurred  time.Time                  `json:"occurred_at"`
	Proof     *logistics.ProofOfDelivery `json:"proof,omitempty"`
	Exception string                     `json:"exception,omitempty"`
	AccountID string                     `json:"account_id,omitempty"`
}

func (s *server) shipmentEvent(w http.ResponseWriter, r *http.Request) {
	var input eventInput
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(input.EventID) == "" || strings.TrimSpace(input.Kind) == "" {
		http.Error(w, "event_id and kind are required", http.StatusBadRequest)
		return
	}
	event := logistics.ShipmentEvent{
		EventID: input.EventID, Shipment: r.PathValue("shipment"), Kind: input.Kind,
		Occurred: input.Occurred, Proof: input.Proof, Exception: input.Exception,
	}
	decision, err := logistics.PublishShipmentEvent(r.Context(), s.client, r.PathValue("channel"), input.AccountID, event)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, decision)
}

func writeError(w http.ResponseWriter, err error) {
	var apiErr *infrai.APIError
	if errors.As(err, &apiErr) && apiErr.HTTPStatus >= 400 && apiErr.HTTPStatus < 500 {
		writeJSON(w, apiErr.HTTPStatus, map[string]string{"error": apiErr.Message, "code": apiErr.Code})
		return
	}
	writeJSON(w, http.StatusBadGateway, map[string]string{"error": "upstream request failed"})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func envOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
