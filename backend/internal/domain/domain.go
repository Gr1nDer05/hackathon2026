package domain

import "time"

type AppStatus struct {
	Status string `json:"status"`
	Name   string `json:"name"`
}

type UserRole string

const (
	RoleAdmin        UserRole = "admin"
	RolePsychologist UserRole = "psychologist"
	RoleClient       UserRole = "client"
)

type User struct {
	ID                int64    `json:"id"`
	Login             string   `json:"login,omitempty"`
	Email             string   `json:"email"`
	FullName          string   `json:"full_name"`
	Role              UserRole `json:"role"`
	IsActive          bool     `json:"is_active"`
	PortalAccessUntil string   `json:"portal_access_until,omitempty"`
	BlockedUntil      string   `json:"blocked_until,omitempty"`
	CreatedAt         string   `json:"created_at,omitempty"`
	UpdatedAt         string   `json:"updated_at,omitempty"`
}

type Test struct {
	ID                  int64  `json:"id"`
	Title               string `json:"title"`
	Description         string `json:"description"`
	CreatedByUserID     int64  `json:"created_by_user_id"`
	ReportTemplateID    int64  `json:"report_template_id,omitempty"`
	RecommendedDuration int    `json:"recommended_duration,omitempty"`
	MaxParticipants     int    `json:"max_participants,omitempty"`
	Status              string `json:"status"`
	PublicSlug          string `json:"public_slug,omitempty"`
	IsPublic            bool   `json:"is_public"`
	PublicURL           string `json:"public_url,omitempty"`
	CreatedAt           string `json:"created_at,omitempty"`
	UpdatedAt           string `json:"updated_at,omitempty"`
}

const (
	TestStatusDraft     = "draft"
	TestStatusPublished = "published"
)

type CreateTestInput struct {
	Title               string `json:"title" binding:"required"`
	Description         string `json:"description"`
	ReportTemplateID    int64  `json:"report_template_id"`
	RecommendedDuration int    `json:"recommended_duration"`
	MaxParticipants     int    `json:"max_participants"`
	Status              string `json:"status"`
}

type UpdateTestInput struct {
	Title               string `json:"title" binding:"required"`
	Description         string `json:"description"`
	ReportTemplateID    int64  `json:"report_template_id"`
	RecommendedDuration int    `json:"recommended_duration"`
	MaxParticipants     int    `json:"max_participants"`
	Status              string `json:"status"`
}

type Question struct {
	ID           int64            `json:"id"`
	TestID       int64            `json:"test_id"`
	Text         string           `json:"text"`
	QuestionType string           `json:"question_type"`
	OrderNumber  int              `json:"order_number"`
	IsRequired   bool             `json:"is_required"`
	Options      []QuestionOption `json:"options,omitempty"`
	CreatedAt    string           `json:"created_at,omitempty"`
	UpdatedAt    string           `json:"updated_at,omitempty"`
}

const (
	QuestionTypeSingleChoice = "single_choice"
	QuestionTypeMultiple     = "multiple_choice"
	QuestionTypeScale        = "scale"
	QuestionTypeText         = "text"
	QuestionTypeNumber       = "number"
)

type QuestionOption struct {
	ID          int64   `json:"id"`
	QuestionID  int64   `json:"question_id"`
	Label       string  `json:"label"`
	Value       string  `json:"value"`
	OrderNumber int     `json:"order_number"`
	Score       float64 `json:"score"`
}

type QuestionOptionInput struct {
	Label       string  `json:"label" binding:"required"`
	Value       string  `json:"value" binding:"required"`
	OrderNumber int     `json:"order_number"`
	Score       float64 `json:"score"`
}

type CreateQuestionInput struct {
	Text         string                `json:"text" binding:"required"`
	QuestionType string                `json:"question_type" binding:"required"`
	OrderNumber  int                   `json:"order_number"`
	IsRequired   bool                  `json:"is_required"`
	Options      []QuestionOptionInput `json:"options"`
}

type UpdateQuestionInput struct {
	Text         string                `json:"text" binding:"required"`
	QuestionType string                `json:"question_type" binding:"required"`
	OrderNumber  int                   `json:"order_number"`
	IsRequired   bool                  `json:"is_required"`
	Options      []QuestionOptionInput `json:"options"`
}

type PublishTestResponse struct {
	TestID     int64  `json:"test_id"`
	PublicSlug string `json:"public_slug"`
	PublicURL  string `json:"public_url"`
	IsPublic   bool   `json:"is_public"`
	Status     string `json:"status"`
}

type PublicQuestionOption struct {
	ID          int64  `json:"id"`
	Label       string `json:"label"`
	Value       string `json:"value"`
	OrderNumber int    `json:"order_number"`
}

type PublicQuestion struct {
	ID           int64                  `json:"id"`
	Text         string                 `json:"text"`
	QuestionType string                 `json:"question_type"`
	OrderNumber  int                    `json:"order_number"`
	IsRequired   bool                   `json:"is_required"`
	Options      []PublicQuestionOption `json:"options,omitempty"`
}

type PublicTest struct {
	ID                  int64            `json:"id"`
	Slug                string           `json:"slug"`
	Title               string           `json:"title"`
	Description         string           `json:"description"`
	RecommendedDuration int              `json:"recommended_duration,omitempty"`
	MaxParticipants     int              `json:"max_participants,omitempty"`
	Questions           []PublicQuestion `json:"questions"`
}

type PublicTestSession struct {
	ID              int64  `json:"id"`
	TestID          int64  `json:"test_id"`
	AccessToken     string `json:"access_token"`
	RespondentName  string `json:"respondent_name,omitempty"`
	RespondentEmail string `json:"respondent_email,omitempty"`
	Status          string `json:"status"`
	StartedAt       string `json:"started_at"`
	CompletedAt     string `json:"completed_at,omitempty"`
}

type StartPublicTestResponse struct {
	Session PublicTestSession `json:"session"`
	Test    PublicTest        `json:"test"`
}

type StartPublicTestInput struct {
	RespondentName  string `json:"respondent_name"`
	RespondentEmail string `json:"respondent_email"`
}

type PublicAnswerInput struct {
	QuestionID   int64    `json:"question_id" binding:"required"`
	AnswerText   string   `json:"answer_text"`
	AnswerValue  string   `json:"answer_value"`
	AnswerValues []string `json:"answer_values"`
}

type SubmitPublicTestInput struct {
	AccessToken string              `json:"access_token" binding:"required"`
	Answers     []PublicAnswerInput `json:"answers" binding:"required"`
}

type PublicTestAnswer struct {
	ID           int64    `json:"id"`
	SessionID    int64    `json:"session_id"`
	QuestionID   int64    `json:"question_id"`
	AnswerText   string   `json:"answer_text,omitempty"`
	AnswerValue  string   `json:"answer_value,omitempty"`
	AnswerValues []string `json:"answer_values,omitempty"`
	CreatedAt    string   `json:"created_at,omitempty"`
	UpdatedAt    string   `json:"updated_at,omitempty"`
}

type SubmitPublicTestResponse struct {
	SessionID int64              `json:"session_id"`
	Status    string             `json:"status"`
	Answers   []PublicTestAnswer `json:"answers"`
}

type PsychologistTestSubmission struct {
	SessionID       int64              `json:"session_id"`
	TestID          int64              `json:"test_id"`
	RespondentName  string             `json:"respondent_name,omitempty"`
	RespondentEmail string             `json:"respondent_email,omitempty"`
	Status          string             `json:"status"`
	StartedAt       string             `json:"started_at"`
	CompletedAt     string             `json:"completed_at,omitempty"`
	AnswersCount    int                `json:"answers_count"`
	Answers         []PublicTestAnswer `json:"answers,omitempty"`
}

type PublicTestAccessInfo struct {
	TestID          int64
	Status          string
	IsPublic        bool
	MaxParticipants int
	CurrentSessions int
}

const (
	FormulaConditionAlways         = "always"
	FormulaConditionAnswerEquals   = "answer_equals"
	FormulaConditionAnswerIn       = "answer_in"
	FormulaConditionAnswerNumericG = "answer_numeric_gte"
	FormulaConditionAnswerNumericL = "answer_numeric_lte"
)

type FormulaRule struct {
	ID            int64   `json:"id"`
	TestID        int64   `json:"test_id"`
	Name          string  `json:"name"`
	QuestionID    int64   `json:"question_id,omitempty"`
	ConditionType string  `json:"condition_type"`
	ExpectedValue string  `json:"expected_value"`
	ScoreDelta    float64 `json:"score_delta"`
	ResultKey     string  `json:"result_key"`
	Priority      int     `json:"priority"`
	CreatedAt     string  `json:"created_at,omitempty"`
	UpdatedAt     string  `json:"updated_at,omitempty"`
}

type CreateFormulaRuleInput struct {
	Name          string  `json:"name" binding:"required"`
	QuestionID    int64   `json:"question_id"`
	ConditionType string  `json:"condition_type" binding:"required"`
	ExpectedValue string  `json:"expected_value"`
	ScoreDelta    float64 `json:"score_delta"`
	ResultKey     string  `json:"result_key"`
	Priority      int     `json:"priority"`
}

type UpdateFormulaRuleInput struct {
	Name          string  `json:"name" binding:"required"`
	QuestionID    int64   `json:"question_id"`
	ConditionType string  `json:"condition_type" binding:"required"`
	ExpectedValue string  `json:"expected_value"`
	ScoreDelta    float64 `json:"score_delta"`
	ResultKey     string  `json:"result_key"`
	Priority      int     `json:"priority"`
}

type FormulaAnswerInput struct {
	QuestionID   int64    `json:"question_id" binding:"required"`
	AnswerValue  string   `json:"answer_value"`
	AnswerValues []string `json:"answer_values"`
}

type CalculateFormulaInput struct {
	Answers []FormulaAnswerInput `json:"answers"`
}

type CalculateFormulaResponse struct {
	TotalScore       float64            `json:"total_score"`
	Metrics          map[string]float64 `json:"metrics"`
	TriggeredRuleIDs []int64            `json:"triggered_rule_ids"`
}

type ClientAnswer struct {
	ID          int64  `json:"id"`
	AttemptID   int64  `json:"attempt_id"`
	QuestionID  int64  `json:"question_id"`
	ClientID    int64  `json:"client_id"`
	AnswerText  string `json:"answer_text"`
	AnswerValue string `json:"answer_value,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
}

type TestAttemptStatus string

const (
	TestAttemptInProgress TestAttemptStatus = "in_progress"
	TestAttemptCompleted  TestAttemptStatus = "completed"
	TestAttemptReviewed   TestAttemptStatus = "reviewed"
)

type TestAttempt struct {
	ID          int64             `json:"id"`
	TestID      int64             `json:"test_id"`
	ClientID    int64             `json:"client_id"`
	AssignedBy  int64             `json:"assigned_by"`
	Status      TestAttemptStatus `json:"status"`
	StartedAt   string            `json:"started_at,omitempty"`
	CompletedAt string            `json:"completed_at,omitempty"`
	CreatedAt   string            `json:"created_at,omitempty"`
	UpdatedAt   string            `json:"updated_at,omitempty"`
}

type Result struct {
	ID              int64   `json:"id"`
	AttemptID       int64   `json:"attempt_id"`
	PsychologistID  int64   `json:"psychologist_id"`
	Summary         string  `json:"summary"`
	Score           float64 `json:"score,omitempty"`
	Interpretation  string  `json:"interpretation,omitempty"`
	Recommendations string  `json:"recommendations,omitempty"`
	GeneratedReport string  `json:"generated_report,omitempty"`
	CreatedAt       string  `json:"created_at,omitempty"`
	UpdatedAt       string  `json:"updated_at,omitempty"`
}

type ReportTemplate struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	TemplateBody string `json:"template_body"`
	CreatedBy    int64  `json:"created_by"`
	CreatedAt    string `json:"created_at,omitempty"`
	UpdatedAt    string `json:"updated_at,omitempty"`
}

type PsychologistRegistrationInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	FullName string `json:"full_name" binding:"required"`
	IsActive bool   `json:"is_active"`
}

type PsychologistLoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type PsychologistProfile struct {
	UserID          int64  `json:"user_id"`
	About           string `json:"about"`
	Specialization  string `json:"specialization"`
	ExperienceYears int    `json:"experience_years"`
	Education       string `json:"education"`
	Methods         string `json:"methods"`
	City            string `json:"city"`
	Timezone        string `json:"timezone"`
	IsPublic        bool   `json:"is_public"`
	CreatedAt       string `json:"created_at,omitempty"`
	UpdatedAt       string `json:"updated_at,omitempty"`
}

type PsychologistCard struct {
	UserID           int64  `json:"user_id"`
	Headline         string `json:"headline"`
	ShortBio         string `json:"short_bio"`
	ContactEmail     string `json:"contact_email"`
	ContactPhone     string `json:"contact_phone"`
	Website          string `json:"website"`
	Telegram         string `json:"telegram"`
	PriceFrom        int    `json:"price_from"`
	OnlineAvailable  bool   `json:"online_available"`
	OfflineAvailable bool   `json:"offline_available"`
	CreatedAt        string `json:"created_at,omitempty"`
	UpdatedAt        string `json:"updated_at,omitempty"`
}

type UpdatePsychologistProfileInput struct {
	About           string `json:"about"`
	Specialization  string `json:"specialization"`
	ExperienceYears int    `json:"experience_years"`
	Education       string `json:"education"`
	Methods         string `json:"methods"`
	City            string `json:"city"`
	Timezone        string `json:"timezone"`
	IsPublic        bool   `json:"is_public"`
}

type UpdatePsychologistCardInput struct {
	Headline         string `json:"headline"`
	ShortBio         string `json:"short_bio"`
	ContactEmail     string `json:"contact_email"`
	ContactPhone     string `json:"contact_phone"`
	Website          string `json:"website"`
	Telegram         string `json:"telegram"`
	PriceFrom        int    `json:"price_from"`
	OnlineAvailable  bool   `json:"online_available"`
	OfflineAvailable bool   `json:"offline_available"`
}

type PsychologistWorkspace struct {
	User    User                `json:"user"`
	Profile PsychologistProfile `json:"profile"`
	Card    PsychologistCard    `json:"card"`
}

type PsychologistAuthResponse struct {
	ExpiresAt time.Time             `json:"expires_at"`
	Workspace PsychologistWorkspace `json:"workspace"`
}

type AuthenticatedUser struct {
	ID                int64    `json:"id"`
	Email             string   `json:"email"`
	FullName          string   `json:"full_name"`
	Role              UserRole `json:"role"`
	IsActive          bool     `json:"is_active"`
	PortalAccessUntil string   `json:"portal_access_until,omitempty"`
	BlockedUntil      string   `json:"blocked_until,omitempty"`
}

type UserCredentials struct {
	User         User
	PasswordHash string
}

type AdminLoginInput struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AdminAuthResponse struct {
	ExpiresAt time.Time `json:"expires_at"`
	User      User      `json:"user"`
}

type AdminSeedInput struct {
	Login    string
	Password string
	FullName string
}

type CreatePsychologistInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	FullName string `json:"full_name" binding:"required"`
	IsActive bool   `json:"is_active"`
}

type UpdatePsychologistAccountInput struct {
	Email    string `json:"email" binding:"required,email"`
	FullName string `json:"full_name" binding:"required"`
	IsActive bool   `json:"is_active"`
}

type UpdatePsychologistAccessInput struct {
	IsActive          *bool   `json:"is_active"`
	PortalAccessUntil *string `json:"portal_access_until"`
	BlockedUntil      *string `json:"blocked_until"`
}

type PsychologistAccessUpdate struct {
	IsActiveSet          bool
	IsActive             bool
	PortalAccessUntilSet bool
	PortalAccessUntil    *time.Time
	BlockedUntilSet      bool
	BlockedUntil         *time.Time
}

const AdminNotificationTypeSubscriptionExpiring = "psychologist_subscription_expiring"

type AdminNotification struct {
	Type              string `json:"type"`
	PsychologistID    int64  `json:"psychologist_id"`
	PsychologistEmail string `json:"psychologist_email"`
	PsychologistName  string `json:"psychologist_name"`
	PortalAccessUntil string `json:"portal_access_until"`
	Message           string `json:"message"`
	Severity          string `json:"severity"`
}
