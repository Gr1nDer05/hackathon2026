package service

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
)

const SubscriptionEndingReminderLookahead = 24 * time.Hour
const SubscriptionEndingReminderCheckInterval = 15 * time.Minute

func (s *AppService) StartBackgroundJobs(ctx context.Context) {
	if s == nil || s.mailer == nil {
		return
	}

	go s.runSubscriptionReminderLoop(ctx)
}

func (s *AppService) runSubscriptionReminderLoop(ctx context.Context) {
	if err := s.ProcessSubscriptionEndingReminders(ctx); err != nil {
		log.Printf("failed to process subscription reminders: %v", err)
	}

	ticker := time.NewTicker(SubscriptionEndingReminderCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.ProcessSubscriptionEndingReminders(ctx); err != nil {
				log.Printf("failed to process subscription reminders: %v", err)
			}
		}
	}
}

func (s *AppService) ProcessSubscriptionEndingReminders(ctx context.Context) error {
	if s == nil || s.mailer == nil {
		return nil
	}

	now := time.Now()
	candidates, err := s.repo.ListSubscriptionReminderCandidates(ctx, now, now.Add(SubscriptionEndingReminderLookahead))
	if err != nil {
		return err
	}

	admins, err := s.repo.ListVerifiedAdmins(ctx)
	if err != nil {
		return err
	}

	for _, candidate := range candidates {
		if candidate.PsychologistReminderSentAt.IsZero() {
			if err := s.sendPsychologistSubscriptionEndingReminder(ctx, candidate); err != nil {
				log.Printf("failed to send psychologist subscription reminder to %s: %v", candidate.PsychologistEmail, err)
			} else if err := s.repo.MarkSubscriptionPsychologistReminderSent(ctx, candidate.PsychologistID, time.Now()); err != nil {
				return err
			}
		}

		if candidate.AdminReminderSentAt.IsZero() && len(admins) > 0 {
			notification := adminNotificationFromReminderCandidate(candidate)
			if err := s.sendAdminSubscriptionEndingReminder(ctx, admins, notification); err != nil {
				log.Printf("failed to send admin subscription reminder for psychologist %d: %v", candidate.PsychologistID, err)
			} else if err := s.repo.MarkSubscriptionAdminReminderSent(ctx, candidate.PsychologistID, time.Now()); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *AppService) sendAdminSubscriptionEndingReminder(ctx context.Context, admins []domain.User, notification domain.AdminNotification) error {
	var firstErr error

	for _, admin := range admins {
		if strings.TrimSpace(admin.Email) == "" {
			continue
		}

		if err := s.mailer.SendAdminSubscriptionEndingReminder(ctx, admin.Email, admin.FullName, notification); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

func (s *AppService) sendPsychologistSubscriptionEndingReminder(ctx context.Context, candidate domain.SubscriptionReminderCandidate) error {
	if strings.TrimSpace(candidate.PsychologistEmail) == "" {
		return nil
	}

	return s.mailer.SendPsychologistSubscriptionEndingReminder(
		ctx,
		candidate.PsychologistEmail,
		candidate.PsychologistName,
		candidate.PortalAccessUntil,
	)
}

func adminNotificationFromReminderCandidate(candidate domain.SubscriptionReminderCandidate) domain.AdminNotification {
	return domain.AdminNotification{
		Type:              domain.AdminNotificationTypeSubscriptionExpiring,
		PsychologistID:    candidate.PsychologistID,
		PsychologistEmail: candidate.PsychologistEmail,
		PsychologistName:  candidate.PsychologistName,
		PortalAccessUntil: candidate.PortalAccessUntil,
		Message:           "Psychologist portal access expires within the next 24 hours",
		Severity:          "warning",
	}
}
