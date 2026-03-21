package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestUserMarshalsNullableDatesAsNull(t *testing.T) {
	t.Helper()

	user := User{
		ID:                 12,
		Email:              "anna@example.com",
		FullName:           "Иванова Анна Сергеевна",
		Role:               RolePsychologist,
		IsActive:           true,
		PortalAccessUntil:  NewNullableString(""),
		BlockedUntil:       NewNullableString(""),
		AccountStatus:      AccountStatusActive,
		SubscriptionStatus: SubscriptionStatusActive,
	}

	body, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("failed to marshal user: %v", err)
	}

	payload := string(body)
	if !strings.Contains(payload, `"portal_access_until":null`) {
		t.Fatalf("expected portal_access_until null, got %s", payload)
	}
	if !strings.Contains(payload, `"blocked_until":null`) {
		t.Fatalf("expected blocked_until null, got %s", payload)
	}
	if !strings.Contains(payload, `"account_status":"active"`) {
		t.Fatalf("expected account_status in payload, got %s", payload)
	}
}
