package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"os"
	"strconv"
	"strings"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
	"github.com/joho/godotenv"
)

type Mailer interface {
	SendPsychologistCredentials(ctx context.Context, email string, fullName string, password string) error
	SendAdminEmailVerificationCode(ctx context.Context, email string, fullName string, code string) error
	SendAdminSubscriptionEndingReminder(ctx context.Context, adminEmail string, adminName string, notification domain.AdminNotification) error
	SendPsychologistSubscriptionEndingReminder(ctx context.Context, email string, fullName string, portalAccessUntil string) error
	SendPsychologistSubscriptionExtended(ctx context.Context, email string, fullName string, portalAccessUntil string, days int) error
}

type SMTPConfig struct {
	Enabled   bool
	Host      string
	Port      int
	Username  string
	Password  string
	FromEmail string
	FromName  string
}

type SMTPMailer struct {
	config SMTPConfig
}

func LoadSMTPConfig() SMTPConfig {
	_ = godotenv.Load()

	port, _ := strconv.Atoi(strings.TrimSpace(os.Getenv("SMTP_PORT")))
	return SMTPConfig{
		Enabled:   parseEnvBool("SMTP_ENABLED"),
		Host:      strings.TrimSpace(os.Getenv("SMTP_HOST")),
		Port:      port,
		Username:  strings.TrimSpace(os.Getenv("SMTP_USERNAME")),
		Password:  os.Getenv("SMTP_PASSWORD"),
		FromEmail: strings.TrimSpace(os.Getenv("SMTP_FROM_EMAIL")),
		FromName:  strings.TrimSpace(os.Getenv("SMTP_FROM_NAME")),
	}
}

func NewSMTPMailer(config SMTPConfig) (*SMTPMailer, error) {
	if !config.Enabled {
		return nil, nil
	}
	if config.Host == "" || config.Port <= 0 || config.Username == "" || config.Password == "" || config.FromEmail == "" {
		return nil, fmt.Errorf("smtp is enabled but config is incomplete")
	}

	return &SMTPMailer{config: config}, nil
}

func (m *SMTPMailer) SendPsychologistCredentials(ctx context.Context, email string, fullName string, password string) error {
	_ = ctx

	subject := "Доступ в ProfDNK"
	body := buildPsychologistCredentialsEmail(fullName, email, password)
	message := buildSMTPMessage(m.fromHeader(), email, subject, body)

	if m.config.Port == 465 {
		return m.sendWithTLS(email, message)
	}

	return m.sendWithSTARTTLS(email, message)
}

func (m *SMTPMailer) SendAdminEmailVerificationCode(ctx context.Context, email string, fullName string, code string) error {
	_ = ctx

	subject := "Подтверждение почты в ProfDNK"
	body := buildAdminEmailVerificationCodeEmail(fullName, code)
	message := buildSMTPMessage(m.fromHeader(), email, subject, body)

	if m.config.Port == 465 {
		return m.sendWithTLS(email, message)
	}

	return m.sendWithSTARTTLS(email, message)
}

func (m *SMTPMailer) SendAdminSubscriptionEndingReminder(ctx context.Context, adminEmail string, adminName string, notification domain.AdminNotification) error {
	_ = ctx

	subject := "У психолога скоро закончится подписка"
	body := buildAdminSubscriptionEndingReminderEmail(adminName, notification)
	message := buildSMTPMessage(m.fromHeader(), adminEmail, subject, body)

	if m.config.Port == 465 {
		return m.sendWithTLS(adminEmail, message)
	}

	return m.sendWithSTARTTLS(adminEmail, message)
}

func (m *SMTPMailer) SendPsychologistSubscriptionEndingReminder(ctx context.Context, email string, fullName string, portalAccessUntil string) error {
	_ = ctx

	subject := "Подписка ProfDNK скоро закончится"
	body := buildPsychologistSubscriptionEndingReminderEmail(fullName, portalAccessUntil)
	message := buildSMTPMessage(m.fromHeader(), email, subject, body)

	if m.config.Port == 465 {
		return m.sendWithTLS(email, message)
	}

	return m.sendWithSTARTTLS(email, message)
}

func (m *SMTPMailer) SendPsychologistSubscriptionExtended(ctx context.Context, email string, fullName string, portalAccessUntil string, days int) error {
	_ = ctx

	subject := "Подписка ProfDNK продлена"
	body := buildPsychologistSubscriptionExtendedEmail(fullName, portalAccessUntil, days)
	message := buildSMTPMessage(m.fromHeader(), email, subject, body)

	if m.config.Port == 465 {
		return m.sendWithTLS(email, message)
	}

	return m.sendWithSTARTTLS(email, message)
}

func (m *SMTPMailer) sendWithTLS(to string, message []byte) error {
	address := fmt.Sprintf("%s:%d", m.config.Host, m.config.Port)
	conn, err := tls.Dial("tcp", address, &tls.Config{
		ServerName: m.config.Host,
		MinVersion: tls.VersionTLS12,
	})
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, m.config.Host)
	if err != nil {
		return err
	}
	defer client.Close()

	return m.sendWithClient(client, to, message)
}

func (m *SMTPMailer) sendWithSTARTTLS(to string, message []byte) error {
	address := fmt.Sprintf("%s:%d", m.config.Host, m.config.Port)
	client, err := smtp.Dial(address)
	if err != nil {
		return err
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{
			ServerName: m.config.Host,
			MinVersion: tls.VersionTLS12,
		}); err != nil {
			return err
		}
	}

	return m.sendWithClient(client, to, message)
}

func (m *SMTPMailer) sendWithClient(client *smtp.Client, to string, message []byte) error {
	auth := smtp.PlainAuth("", m.config.Username, m.config.Password, m.config.Host)
	if ok, _ := client.Extension("AUTH"); ok {
		if err := client.Auth(auth); err != nil {
			return err
		}
	}

	if err := client.Mail(m.config.FromEmail); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}

	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := writer.Write(message); err != nil {
		_ = writer.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	return client.Quit()
}

func (m *SMTPMailer) fromHeader() string {
	if m.config.FromName == "" {
		return m.config.FromEmail
	}

	return fmt.Sprintf("%s <%s>", m.config.FromName, m.config.FromEmail)
}

func buildPsychologistCredentialsEmail(fullName string, email string, password string) string {
	name := strings.TrimSpace(fullName)
	if name == "" {
		name = "коллега"
	}

	return fmt.Sprintf(
		"Здравствуйте, %s!\r\n\r\nВаш аккаунт психолога в ProfDNK создан.\r\n\r\nЛогин: %s\r\nПароль: %s\r\n\r\n",
		name,
		email,
		password,
	)
}

func buildAdminEmailVerificationCodeEmail(fullName string, code string) string {
	name := strings.TrimSpace(fullName)
	if name == "" {
		name = "администратор"
	}

	return fmt.Sprintf(
		"Здравствуйте, %s!\r\n\r\nВаш код подтверждения почты в ProfDNK: %s\r\n\r\nВведите этот код в админ-панели, чтобы подтвердить почту.\r\n\r\n",
		name,
		code,
	)
}

func buildAdminSubscriptionEndingReminderEmail(adminName string, notification domain.AdminNotification) string {
	name := strings.TrimSpace(adminName)
	if name == "" {
		name = "администратор"
	}

	return fmt.Sprintf(
		"Здравствуйте, %s!\r\n\r\nУ психолога скоро закончится подписка.\r\n\r\nID: %d\r\nФИО: %s\r\nEmail: %s\r\nПодписка активна до: %s\r\n\r\nПожалуйста, продлите подписку в админ-панели.\r\n\r\n",
		name,
		notification.PsychologistID,
		notification.PsychologistName,
		notification.PsychologistEmail,
		notification.PortalAccessUntil,
	)
}

func buildPsychologistSubscriptionEndingReminderEmail(fullName string, portalAccessUntil string) string {
	name := strings.TrimSpace(fullName)
	if name == "" {
		name = "коллега"
	}

	body := fmt.Sprintf("Здравствуйте, %s!\r\n\r\nСрок действия вашей подписки ProfDNK скоро закончится.", name)
	if strings.TrimSpace(portalAccessUntil) != "" {
		body += fmt.Sprintf("\r\nПодписка активна до: %s.", portalAccessUntil)
	}

	return body + "\r\nПожалуйста, обратитесь к администратору для продления подписки.\r\n\r\n"
}

func buildPsychologistSubscriptionExtendedEmail(fullName string, portalAccessUntil string, days int) string {
	name := strings.TrimSpace(fullName)
	if name == "" {
		name = "коллега"
	}

	body := fmt.Sprintf("Здравствуйте, %s!\r\n\r\nВаша подписка ProfDNK была продлена", name)
	if days > 0 {
		body += fmt.Sprintf(" на %d дн.", days)
	}
	if strings.TrimSpace(portalAccessUntil) != "" {
		body += fmt.Sprintf("\r\nНовый срок действия подписки: %s.", portalAccessUntil)
	}

	return body + "\r\n\r\n"
}

func buildSMTPMessage(from string, to string, subject string, body string) []byte {
	headers := []string{
		"From: " + from,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
	}

	return []byte(strings.Join(headers, "\r\n") + "\r\n\r\n" + body)
}

func parseEnvBool(key string) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	return value == "1" || value == "true" || value == "yes" || value == "on"
}
