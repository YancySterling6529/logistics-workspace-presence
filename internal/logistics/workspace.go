package logistics

import (
	"context"
	"fmt"
	"time"

	"github.com/infrai-examples/logistics-presence/internal/infrai"
)

type PresenceReader interface {
	Presence(context.Context, string) ([]infrai.PresenceMember, error)
}

type Publisher interface {
	Publish(context.Context, string, string, any, string, string) error
}

type Workspace struct {
	Channel string
	Online  []string
}

func LoadWorkspace(ctx context.Context, reader PresenceReader, channel string) (Workspace, error) {
	members, err := reader.Presence(ctx, channel)
	if err != nil {
		return Workspace{}, err
	}
	workspace := Workspace{Channel: channel, Online: make([]string, 0, len(members))}
	for _, member := range members {
		workspace.Online = append(workspace.Online, member.ClientID)
	}
	return workspace, nil
}

type ProofOfDelivery struct {
	Name        string `json:"name"`
	ContentType string `json:"content_type"`
	Reference   string `json:"reference"`
}

type ShipmentEvent struct {
	EventID   string           `json:"event_id"`
	Shipment  string           `json:"shipment"`
	Kind      string           `json:"kind"`
	Occurred  time.Time        `json:"occurred_at"`
	Proof     *ProofOfDelivery `json:"proof,omitempty"`
	Exception string           `json:"exception,omitempty"`
}

type ExceptionDecision struct {
	State  string `json:"state"`
	Reason string `json:"reason"`
}

func DecideException(event ShipmentEvent) ExceptionDecision {
	if event.Exception != "" {
		return ExceptionDecision{State: "open", Reason: event.Exception}
	}
	if event.Kind == "delivered" && event.Proof == nil {
		return ExceptionDecision{State: "open", Reason: "proof of delivery required"}
	}
	return ExceptionDecision{State: "clear"}
}

func PublishShipmentEvent(ctx context.Context, publisher Publisher, channel, accountID string, event ShipmentEvent) (ExceptionDecision, error) {
	decision := DecideException(event)
	payload := struct {
		ShipmentEvent ShipmentEvent     `json:"shipment_event"`
		Decision      ExceptionDecision `json:"exception_decision"`
	}{ShipmentEvent: event, Decision: decision}
	if err := publisher.Publish(ctx, channel, "shipment.updated", payload, accountID, event.EventID); err != nil {
		return ExceptionDecision{}, fmt.Errorf("publish shipment event: %w", err)
	}
	return decision, nil
}
