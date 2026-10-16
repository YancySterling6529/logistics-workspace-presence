package logistics

import "testing"

func TestDecideException(t *testing.T) {
	tests := []struct {
		name  string
		event ShipmentEvent
		want  ExceptionDecision
	}{
		{name: "carrier exception stays open", event: ShipmentEvent{Kind: "in_transit", Exception: "customs hold"}, want: ExceptionDecision{State: "open", Reason: "customs hold"}},
		{name: "delivery needs proof", event: ShipmentEvent{Kind: "delivered"}, want: ExceptionDecision{State: "open", Reason: "proof of delivery required"}},
		{name: "delivery with proof is clear", event: ShipmentEvent{Kind: "delivered", Proof: &ProofOfDelivery{Name: "receipt.jpg", ContentType: "image/jpeg", Reference: "pod-4581"}}, want: ExceptionDecision{State: "clear"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DecideException(tt.event); got != tt.want {
				t.Fatalf("DecideException() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
