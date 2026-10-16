package infrai

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPresenceDecodesBusinessRejectionBeforeStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"ok":false,"data":null,"error":{"code":"INVALID_ARGUMENT","message":"channel is required"},"metadata":null}`))
	}))
	defer server.Close()

	client := newClient(server.URL, "test-key", server.Client())
	_, err := client.Presence(context.Background(), "dispatch/east")
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("Presence() error = %T, want *APIError", err)
	}
	if apiErr.Code != "INVALID_ARGUMENT" || apiErr.HTTPStatus != http.StatusBadRequest {
		t.Fatalf("Presence() error = %#v", apiErr)
	}
}
