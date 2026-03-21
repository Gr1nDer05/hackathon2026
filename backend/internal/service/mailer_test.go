package service

import (
	"strings"
	"testing"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
)

func TestBuildPsychologistCredentialsEmailContainsLoginAndPassword(t *testing.T) {
	body := buildPsychologistCredentialsEmail("Иванов Иван Иванович", "psych@example.com", "secret123")

	if !strings.Contains(body, "Логин: psych@example.com") {
		t.Fatalf("expected login in email body, got %q", body)
	}
	if !strings.Contains(body, "Пароль: secret123") {
		t.Fatalf("expected password in email body, got %q", body)
	}
}

func TestBuildSMTPMessageContainsHeaders(t *testing.T) {
	message := string(buildSMTPMessage("ProfDNK <noreply@mail.profdnk.ru>", "user@example.com", "Subject", "Body"))

	if !strings.Contains(message, "From: ProfDNK <noreply@mail.profdnk.ru>") {
		t.Fatalf("expected From header, got %q", message)
	}
	if !strings.Contains(message, "To: user@example.com") {
		t.Fatalf("expected To header, got %q", message)
	}
	if !strings.Contains(message, "Subject: Subject") {
		t.Fatalf("expected Subject header, got %q", message)
	}
}

func TestBuildAdminEmailVerificationCodeEmailContainsCode(t *testing.T) {
	body := buildAdminEmailVerificationCodeEmail("Админ", "123456")

	if !strings.Contains(body, "123456") {
		t.Fatalf("expected verification code in email body, got %q", body)
	}
}

func TestBuildAdminSubscriptionEndingReminderEmailContainsPsychologistData(t *testing.T) {
	body := buildAdminSubscriptionEndingReminderEmail("Админ", domain.AdminNotification{
		PsychologistID:    42,
		PsychologistName:  "Иванов Иван",
		PsychologistEmail: "psych@example.com",
		PortalAccessUntil: "2026-03-22T12:00:00Z",
	})

	if !strings.Contains(body, "Иванов Иван") || !strings.Contains(body, "psych@example.com") {
		t.Fatalf("expected psychologist details in admin reminder, got %q", body)
	}
}

func TestBuildPsychologistSubscriptionExtendedEmailContainsDuration(t *testing.T) {
	body := buildPsychologistSubscriptionExtendedEmail("Иванов Иван", "2026-04-20T12:00:00Z", 30)

	if !strings.Contains(body, "30") || !strings.Contains(body, "2026-04-20T12:00:00Z") {
		t.Fatalf("expected extension details in psychologist email, got %q", body)
	}
}
