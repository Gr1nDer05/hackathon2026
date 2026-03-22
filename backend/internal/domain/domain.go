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
	ID                 int64          `json:"id"`
	Login              string         `json:"login,omitempty"`
	Email              string         `json:"email"`
	FullName           string         `json:"full_name"`
	Role               UserRole       `json:"role"`
	IsActive           bool           `json:"is_active"`
	PortalAccessUntil  NullableString `json:"portal_access_until"`
	BlockedUntil       NullableString `json:"blocked_until"`
	SubscriptionPlan   string         `json:"subscription_plan,omitempty"`
	AccountStatus      string         `json:"account_status,omitempty"`
	SubscriptionStatus string         `json:"subscription_status,omitempty"`
	CreatedAt          string         `json:"created_at,omitempty"`
	UpdatedAt          string         `json:"updated_at,omitempty"`
}

type Test struct {
	ID                          int64          `json:"id"`
	Title                       string         `json:"title"`
	Description                 string         `json:"description"`
	CreatedByUserID             int64          `json:"created_by_user_id"`
	ReportTemplateID            int64          `json:"report_template_id"`
	RecommendedDuration         int            `json:"recommended_duration,omitempty"`
	MaxParticipants             int            `json:"max_participants,omitempty"`
	HasParticipantLimit         bool           `json:"has_participant_limit"`
	CollectRespondentAge        bool           `json:"collect_respondent_age"`
	CollectRespondentGender     bool           `json:"collect_respondent_gender"`
	CollectRespondentEducation  bool           `json:"collect_respondent_education"`
	ShowClientReportImmediately bool           `json:"show_client_report_immediately"`
	Status                      string         `json:"status"`
	PublicSlug                  string         `json:"public_slug,omitempty"`
	IsPublic                    bool           `json:"is_public"`
	PublicURL                   string         `json:"public_url,omitempty"`
	StartedSessionsCount        int            `json:"started_sessions_count"`
	InProgressSessionsCount     int            `json:"in_progress_sessions_count"`
	CompletedSessionsCount      int            `json:"completed_sessions_count"`
	LastStartedAt               NullableString `json:"last_started_at"`
	LastCompletedAt             NullableString `json:"last_completed_at"`
	LastActivityAt              NullableString `json:"last_activity_at"`
	CreatedAt                   string         `json:"created_at,omitempty"`
	UpdatedAt                   string         `json:"updated_at,omitempty"`
}

const (
	TestStatusDraft     = "draft"
	TestStatusPublished = "published"
)

const (
	SubscriptionPlanBasic = "basic"
	SubscriptionPlanPro   = "pro"
)

type CreateTestInput struct {
	Title                       string `json:"title" binding:"required"`
	Description                 string `json:"description"`
	ReportTemplateID            int64  `json:"report_template_id"`
	RecommendedDuration         int    `json:"recommended_duration"`
	MaxParticipants             int    `json:"max_participants"`
	HasParticipantLimit         *bool  `json:"has_participant_limit,omitempty"`
	CollectRespondentAge        bool   `json:"collect_respondent_age"`
	CollectRespondentGender     bool   `json:"collect_respondent_gender"`
	CollectRespondentEducation  bool   `json:"collect_respondent_education"`
	ShowClientReportImmediately bool   `json:"show_client_report_immediately"`
	Status                      string `json:"status"`
}

type UpdateTestInput struct {
	Title                       string `json:"title" binding:"required"`
	Description                 string `json:"description"`
	ReportTemplateID            int64  `json:"report_template_id"`
	RecommendedDuration         int    `json:"recommended_duration"`
	MaxParticipants             int    `json:"max_participants"`
	HasParticipantLimit         *bool  `json:"has_participant_limit,omitempty"`
	CollectRespondentAge        bool   `json:"collect_respondent_age"`
	CollectRespondentGender     bool   `json:"collect_respondent_gender"`
	CollectRespondentEducation  bool   `json:"collect_respondent_education"`
	ShowClientReportImmediately bool   `json:"show_client_report_immediately"`
	Status                      string `json:"status"`
}

type Question struct {
	ID           int64              `json:"id"`
	TestID       int64              `json:"test_id"`
	Text         string             `json:"text"`
	QuestionType string             `json:"question_type"`
	OrderNumber  int                `json:"order_number"`
	IsRequired   bool               `json:"is_required"`
	ScaleWeights map[string]float64 `json:"scale_weights,omitempty"`
	Options      []QuestionOption   `json:"options,omitempty"`
	CreatedAt    string             `json:"created_at,omitempty"`
	UpdatedAt    string             `json:"updated_at,omitempty"`
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
	ScaleWeights map[string]float64    `json:"scale_weights"`
	Options      []QuestionOptionInput `json:"options"`
}

type UpdateQuestionInput struct {
	Text         string                `json:"text" binding:"required"`
	QuestionType string                `json:"question_type" binding:"required"`
	OrderNumber  int                   `json:"order_number"`
	IsRequired   bool                  `json:"is_required"`
	ScaleWeights *map[string]float64   `json:"scale_weights"`
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
	ID          int64   `json:"id"`
	Label       string  `json:"label"`
	Value       string  `json:"value"`
	OrderNumber int     `json:"order_number"`
	Score       float64 `json:"-"`
}

type PublicQuestion struct {
	ID           int64                  `json:"id"`
	Text         string                 `json:"text"`
	QuestionType string                 `json:"question_type"`
	OrderNumber  int                    `json:"order_number"`
	IsRequired   bool                   `json:"is_required"`
	ScaleWeights map[string]float64     `json:"-"`
	Options      []PublicQuestionOption `json:"options,omitempty"`
}

type PublicPsychologistUser struct {
	ID       int64  `json:"id"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
}

type PublicPsychologistProfile struct {
	Specialization  string `json:"specialization"`
	City            string `json:"city"`
	About           string `json:"about"`
	Education       string `json:"education"`
	Methods         string `json:"methods"`
	ExperienceYears int    `json:"experience_years"`
	Timezone        string `json:"timezone"`
	IsPublic        bool   `json:"is_public"`
}

type PublicPsychologistCard struct {
	Headline         string `json:"headline"`
	ShortBio         string `json:"short_bio"`
	ContactEmail     string `json:"contact_email"`
	ContactPhone     string `json:"contact_phone"`
	Telegram         string `json:"telegram"`
	OnlineAvailable  bool   `json:"online_available"`
	OfflineAvailable bool   `json:"offline_available"`
}

type PublicPsychologist struct {
	User    PublicPsychologistUser    `json:"user"`
	Profile PublicPsychologistProfile `json:"profile"`
	Card    PublicPsychologistCard    `json:"card"`
}

type PublicTest struct {
	ID                          int64              `json:"id"`
	Slug                        string             `json:"slug"`
	Title                       string             `json:"title"`
	Description                 string             `json:"description"`
	RecommendedDuration         int                `json:"recommended_duration,omitempty"`
	MaxParticipants             int                `json:"max_participants,omitempty"`
	HasParticipantLimit         bool               `json:"has_participant_limit"`
	CollectRespondentAge        bool               `json:"collect_respondent_age"`
	CollectRespondentGender     bool               `json:"collect_respondent_gender"`
	CollectRespondentEducation  bool               `json:"collect_respondent_education"`
	ShowClientReportImmediately bool               `json:"show_client_report_immediately"`
	Psychologist                PublicPsychologist `json:"psychologist"`
	Questions                   []PublicQuestion   `json:"questions"`
}

type PublicTestSession struct {
	ID                  int64  `json:"id"`
	TestID              int64  `json:"test_id"`
	AccessToken         string `json:"access_token"`
	RespondentName      string `json:"respondent_name,omitempty"`
	RespondentPhone     string `json:"respondent_phone,omitempty"`
	RespondentEmail     string `json:"respondent_email,omitempty"`
	RespondentAge       int    `json:"respondent_age,omitempty"`
	RespondentGender    string `json:"respondent_gender,omitempty"`
	RespondentEducation string `json:"respondent_education,omitempty"`
	Status              string `json:"status"`
	StartedAt           string `json:"started_at"`
	ExpiresAt           string `json:"expires_at,omitempty"`
	CompletedAt         string `json:"completed_at,omitempty"`
}

type StartPublicTestResponse struct {
	Session PublicTestSession  `json:"session"`
	Test    PublicTest         `json:"test"`
	Answers []PublicTestAnswer `json:"answers"`
	Resumed bool               `json:"resumed"`
}

type StartPublicTestInput struct {
	RespondentName      string `json:"respondent_name" binding:"required"`
	RespondentPhone     string `json:"respondent_phone" binding:"required"`
	RespondentEmail     string `json:"respondent_email"`
	RespondentAge       int    `json:"respondent_age"`
	RespondentGender    string `json:"respondent_gender"`
	RespondentEducation string `json:"respondent_education"`
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
	SessionID             int64              `json:"session_id"`
	Status                string             `json:"status"`
	Answers               []PublicTestAnswer `json:"answers"`
	ClientReportAvailable bool               `json:"client_report_available"`
	ClientReportURL       string             `json:"client_report_url,omitempty"`
	Metrics               map[string]float64 `json:"metrics,omitempty"`
	TopMetrics            []ResultMetric     `json:"top_metrics,omitempty"`
	CareerResult          *CareerResult      `json:"career_result,omitempty"`
}

type PsychologistTestSubmission struct {
	SessionID           int64              `json:"session_id"`
	TestID              int64              `json:"test_id"`
	RespondentName      string             `json:"respondent_name,omitempty"`
	RespondentPhone     string             `json:"respondent_phone,omitempty"`
	RespondentEmail     string             `json:"respondent_email,omitempty"`
	RespondentAge       int                `json:"respondent_age,omitempty"`
	RespondentGender    string             `json:"respondent_gender,omitempty"`
	RespondentEducation string             `json:"respondent_education,omitempty"`
	Status              string             `json:"status"`
	StartedAt           string             `json:"started_at"`
	CompletedAt         string             `json:"completed_at,omitempty"`
	AnswersCount        int                `json:"answers_count"`
	Answers             []PublicTestAnswer `json:"answers,omitempty"`
	Metrics             map[string]float64 `json:"metrics,omitempty"`
	TopMetrics          []ResultMetric     `json:"top_metrics,omitempty"`
	CareerResult        *CareerResult      `json:"career_result,omitempty"`
}

type PublicTestAccessInfo struct {
	TestID          int64
	Status          string
	IsPublic        bool
	MaxParticipants int
	CurrentSessions int
}

const (
	CareerScaleAnalytic  = "analytic"
	CareerScaleCreative  = "creative"
	CareerScaleSocial    = "social"
	CareerScaleOrganizer = "organizer"
	CareerScalePractical = "practical"
)

var CareerScales = []string{
	CareerScaleAnalytic,
	CareerScaleCreative,
	CareerScaleSocial,
	CareerScaleOrganizer,
	CareerScalePractical,
}

type CareerScaleResult struct {
	Scale      string  `json:"scale"`
	RawScore   float64 `json:"raw_score"`
	MaxScore   float64 `json:"max_score"`
	Percentage float64 `json:"percentage"`
}

type CareerProfessionResult struct {
	Profession string  `json:"profession"`
	Score      float64 `json:"score"`
}

type ResultMetric struct {
	Key   string  `json:"key"`
	Value float64 `json:"value"`
}

type CareerResult struct {
	Scales         []CareerScaleResult      `json:"scales"`
	TopScales      []CareerScaleResult      `json:"top_scales"`
	TopProfessions []CareerProfessionResult `json:"top_professions"`
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

type ReportTemplate struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	TemplateBody string `json:"template_body"`
	CreatedBy    int64  `json:"created_by"`
	CreatedAt    string `json:"created_at,omitempty"`
	UpdatedAt    string `json:"updated_at,omitempty"`
}

type CreateReportTemplateInput struct {
	Name         string `json:"name" binding:"required"`
	Description  string `json:"description"`
	TemplateBody string `json:"template_body" binding:"required"`
}

type UpdateReportTemplateInput struct {
	Name         string `json:"name" binding:"required"`
	Description  string `json:"description"`
	TemplateBody string `json:"template_body" binding:"required"`
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
	Tests   []Test              `json:"tests"`
}

type PsychologistAuthResponse struct {
	ExpiresAt time.Time             `json:"expires_at"`
	Workspace PsychologistWorkspace `json:"workspace"`
}

type AuthenticatedUser struct {
	ID                int64          `json:"id"`
	Email             string         `json:"email"`
	FullName          string         `json:"full_name"`
	Role              UserRole       `json:"role"`
	IsActive          bool           `json:"is_active"`
	PortalAccessUntil NullableString `json:"portal_access_until"`
	BlockedUntil      NullableString `json:"blocked_until"`
	SubscriptionPlan  string         `json:"subscription_plan,omitempty"`
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

type UpdateAdminMeInput struct {
	Email string `json:"email" binding:"required,email"`
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
	IsActive              *bool               `json:"is_active"`
	PortalAccessUntil     OptionalStringInput `json:"portal_access_until"`
	BlockedUntil          OptionalStringInput `json:"blocked_until"`
	SubscriptionPlan      string              `json:"subscription_plan,omitempty"`
	SubscriptionDays      *int                `json:"subscription_days,omitempty"`
	SubscriptionDaysAlias *int                `json:"subscriptionDays,omitempty"`
}

type PsychologistAccessUpdate struct {
	IsActiveSet          bool
	IsActive             bool
	PortalAccessUntilSet bool
	PortalAccessUntil    *time.Time
	SubscriptionPlanSet  bool
	SubscriptionPlan     string
	SubscriptionDaysSet  bool
	SubscriptionDays     int
	BlockedUntilSet      bool
	BlockedUntil         *time.Time
}

type GenerateReportTemplateDraftInput struct {
	Prompt string `json:"prompt" binding:"required"`
	TestID int64  `json:"test_id,omitempty"`
}

type GenerateReportTemplateDraftResponse struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	TemplateBody string `json:"template_body"`
	Model        string `json:"model,omitempty"`
}

type CreateSubscriptionPurchaseRequestInput struct {
	SubscriptionPlan string `json:"subscription_plan" binding:"required"`
}

type SubscriptionPurchaseRequest struct {
	ID                int64  `json:"id"`
	PsychologistID    int64  `json:"psychologist_id"`
	PsychologistName  string `json:"psychologist_name"`
	PsychologistEmail string `json:"psychologist_email"`
	SubscriptionPlan  string `json:"subscription_plan"`
	DurationDays      int    `json:"duration_days"`
	Status            string `json:"status"`
	CreatedAt         string `json:"created_at,omitempty"`
	UpdatedAt         string `json:"updated_at,omitempty"`
}
