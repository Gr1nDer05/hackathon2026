package domain

import (
	"bytes"
	"encoding/json"
)

// OptionalStringInput keeps the difference between an omitted field and an
// explicit null so admin access updates can clear deadlines via JSON null.
type OptionalStringInput struct {
	Set   bool
	Value *string
}

func (o *OptionalStringInput) UnmarshalJSON(data []byte) error {
	o.Set = true

	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		o.Value = nil
		return nil
	}

	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	o.Value = &value
	return nil
}
