package domain

import (
	"encoding/json"
	"strings"
	"time"
)

type NullableString string

func NewNullableString(value string) NullableString {
	return NullableString(strings.TrimSpace(value))
}

func (s NullableString) String() string {
	return string(s)
}

func (s NullableString) IsZero() bool {
	return strings.TrimSpace(string(s)) == ""
}

func (s NullableString) MarshalJSON() ([]byte, error) {
	if s.IsZero() {
		return []byte("null"), nil
	}

	return json.Marshal(string(s))
}

const (
	AccountStatusActive  = "active"
	AccountStatusBlocked = "blocked"

	SubscriptionStatusActive  = "active"
	SubscriptionStatusBlocked = "blocked"
	SubscriptionStatusExpired = "expired"
)

func ResolvePsychologistStatuses(isActive bool, portalAccessUntil NullableString, blockedUntil NullableString, now time.Time) (string, string) {
	accountStatus := AccountStatusActive
	if !isActive {
		accountStatus = AccountStatusBlocked
	}

	if until, ok := parseNullableRFC3339(blockedUntil); ok && until.After(now) {
		accountStatus = AccountStatusBlocked
	}

	if accountStatus == AccountStatusBlocked {
		return accountStatus, SubscriptionStatusBlocked
	}

	subscriptionStatus := SubscriptionStatusActive
	if until, ok := parseNullableRFC3339(portalAccessUntil); ok && !until.After(now) {
		subscriptionStatus = SubscriptionStatusExpired
	}

	return accountStatus, subscriptionStatus
}

func parseNullableRFC3339(value NullableString) (time.Time, bool) {
	if value.IsZero() {
		return time.Time{}, false
	}

	parsed, err := time.Parse(time.RFC3339, value.String())
	if err != nil {
		return time.Time{}, false
	}

	return parsed, true
}

func NormalizeSubscriptionPlan(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "", SubscriptionPlanBasic:
		return SubscriptionPlanBasic
	case SubscriptionPlanPro:
		return SubscriptionPlanPro
	default:
		return ""
	}
}

func IsProSubscriptionPlan(value string) bool {
	return NormalizeSubscriptionPlan(value) == SubscriptionPlanPro
}
