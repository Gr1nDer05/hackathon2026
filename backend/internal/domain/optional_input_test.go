package domain

import (
	"encoding/json"
	"testing"
)

func TestOptionalStringInputDistinguishesNullFromOmitted(t *testing.T) {
	type payload struct {
		Value OptionalStringInput `json:"value"`
	}

	var withNull payload
	if err := json.Unmarshal([]byte(`{"value":null}`), &withNull); err != nil {
		t.Fatalf("failed to unmarshal null payload: %v", err)
	}
	if !withNull.Value.Set {
		t.Fatalf("expected null field to be marked as set")
	}
	if withNull.Value.Value != nil {
		t.Fatalf("expected null field to keep nil value")
	}

	var omitted payload
	if err := json.Unmarshal([]byte(`{}`), &omitted); err != nil {
		t.Fatalf("failed to unmarshal omitted payload: %v", err)
	}
	if omitted.Value.Set {
		t.Fatalf("expected omitted field to stay unset")
	}
}
