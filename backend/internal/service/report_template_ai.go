package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
)

var (
	ErrReportTemplateDraftUnavailable = errors.New("report template draft generation unavailable")
	ErrReportTemplateDraftForbidden   = errors.New("report template draft generation requires pro plan")
	ErrReportTemplateDraftFailed      = errors.New("report template draft generation failed")

	singleQuotedJSONTokenPattern = regexp.MustCompile(`'([^'\\]*(?:\\.[^'\\]*)*)'`)
	unquotedJSONKeyPattern       = regexp.MustCompile(`([{\[,]\s*|^\s*)([A-Za-z_][A-Za-z0-9_]*)(\s*:)`)
	trailingCommaPattern         = regexp.MustCompile(`,(\s*[}\]])`)
	pythonBooleanPattern         = regexp.MustCompile(`\bTrue\b|\bFalse\b|\bNone\b`)
)

type reportTemplateDraftGenerator interface {
	Enabled() bool
	Model() string
	GenerateReportTemplateDraft(ctx context.Context, request reportTemplateDraftRequest) (generatedReportTemplateDraft, error)
}

type reportTemplateDraftRequest struct {
	Prompt      string
	TestContext string
}

type generatedReportTemplateDraft struct {
	Name         string                       `json:"name"`
	Description  string                       `json:"description"`
	Client       reportAudienceTemplateConfig `json:"client"`
	Psychologist reportAudienceTemplateConfig `json:"psychologist"`
}

type openAIReportTemplateDraftGenerator struct {
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
}

type openAIResponsesRequest struct {
	Model           string               `json:"model"`
	Input           []openAIInputMessage `json:"input"`
	Text            openAITextConfig     `json:"text"`
	MaxOutputTokens int                  `json:"max_output_tokens,omitempty"`
}

type openAIChatCompletionsRequest struct {
	Model             string                         `json:"model"`
	Messages          []openAIChatMessage            `json:"messages"`
	ResponseFormat    *openAIChatCompletionFormat    `json:"response_format,omitempty"`
	StructuredOutputs bool                           `json:"structured_outputs,omitempty"`
	Provider          *openRouterProviderPreferences `json:"provider,omitempty"`
	MaxTokens         int                            `json:"max_tokens,omitempty"`
}

type openRouterProviderPreferences struct {
	AllowFallbacks    bool `json:"allow_fallbacks,omitempty"`
	RequireParameters bool `json:"require_parameters,omitempty"`
}

type openAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIChatCompletionFormat struct {
	Type       string                     `json:"type"`
	JSONSchema openAIChatCompletionSchema `json:"json_schema,omitempty"`
}

type openAIChatCompletionSchema struct {
	Name   string `json:"name"`
	Strict bool   `json:"strict,omitempty"`
	Schema any    `json:"schema"`
}

type openAIInputMessage struct {
	Role    string               `json:"role"`
	Content []openAIInputContent `json:"content"`
}

type openAIInputContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type openAITextConfig struct {
	Format openAITextFormat `json:"format"`
}

type openAITextFormat struct {
	Type   string `json:"type"`
	Name   string `json:"name,omitempty"`
	Strict bool   `json:"strict,omitempty"`
	Schema any    `json:"schema,omitempty"`
}

type openAIResponsesResponse struct {
	OutputText string             `json:"output_text"`
	Output     []openAIOutputItem `json:"output"`
	Error      *openAIAPIError    `json:"error,omitempty"`
}

type openAIChatCompletionsResponse struct {
	Choices []openAIChatChoice `json:"choices"`
	Error   *openAIAPIError    `json:"error,omitempty"`
}

type openAIChatChoice struct {
	Message openAIChatChoiceMessage `json:"message"`
}

type openAIChatChoiceMessage struct {
	Content any `json:"content"`
}

type openAIOutputItem struct {
	Type    string                `json:"type"`
	Content []openAIOutputContent `json:"content"`
}

type openAIOutputContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type openAIAPIError struct {
	Message  string         `json:"message"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

func newReportTemplateDraftGeneratorFromEnv() reportTemplateDraftGenerator {
	apiKey := strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	if apiKey == "" {
		return nil
	}

	baseURL := strings.TrimSpace(os.Getenv("OPENAI_BASE_URL"))
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	model := strings.TrimSpace(os.Getenv("OPENAI_REPORT_TEMPLATE_MODEL"))
	if model == "" {
		model = "gpt-5-mini"
	}

	return &openAIReportTemplateDraftGenerator{
		apiKey:  apiKey,
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
		httpClient: &http.Client{
			Timeout: 45 * time.Second,
		},
	}
}

func (g *openAIReportTemplateDraftGenerator) Enabled() bool {
	return g != nil && strings.TrimSpace(g.apiKey) != ""
}

func (g *openAIReportTemplateDraftGenerator) Model() string {
	if g == nil {
		return ""
	}
	return g.model
}

func (g *openAIReportTemplateDraftGenerator) GenerateReportTemplateDraft(ctx context.Context, request reportTemplateDraftRequest) (generatedReportTemplateDraft, error) {
	if !g.Enabled() {
		return generatedReportTemplateDraft{}, ErrReportTemplateDraftUnavailable
	}

	if g.usesOpenRouterCompletionsAPI() {
		return g.generateReportTemplateDraftWithChatCompletions(ctx, request)
	}

	return g.generateReportTemplateDraftWithResponses(ctx, request)
}

func (g *openAIReportTemplateDraftGenerator) generateReportTemplateDraftWithResponses(ctx context.Context, request reportTemplateDraftRequest) (generatedReportTemplateDraft, error) {
	body, err := json.Marshal(openAIResponsesRequest{
		Model: g.model,
		Input: []openAIInputMessage{
			{
				Role: "system",
				Content: []openAIInputContent{{
					Type: "input_text",
					Text: buildReportTemplateDraftSystemPrompt(),
				}},
			},
			{
				Role: "user",
				Content: []openAIInputContent{{
					Type: "input_text",
					Text: buildReportTemplateDraftUserPrompt(request),
				}},
			},
		},
		Text: openAITextConfig{
			Format: openAITextFormat{
				Type:   "json_schema",
				Name:   "report_template_draft",
				Strict: true,
				Schema: reportTemplateDraftJSONSchema(),
			},
		},
		MaxOutputTokens: 1800,
	})
	if err != nil {
		return generatedReportTemplateDraft{}, err
	}

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, g.baseURL+"/responses", bytes.NewReader(body))
	if err != nil {
		return generatedReportTemplateDraft{}, err
	}
	httpRequest.Header.Set("Authorization", "Bearer "+g.apiKey)
	httpRequest.Header.Set("Content-Type", "application/json")

	response, err := g.httpClient.Do(httpRequest)
	if err != nil {
		return generatedReportTemplateDraft{}, providerClientFailed(err, "AI provider request failed")
	}
	defer response.Body.Close()

	payload, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return generatedReportTemplateDraft{}, providerClientFailed(err, "AI provider response read failed")
	}

	var decoded openAIResponsesResponse
	if err := json.Unmarshal(payload, &decoded); err != nil {
		if response.StatusCode >= http.StatusBadRequest {
			return generatedReportTemplateDraft{}, providerRequestFailed(payload, "AI provider request failed")
		}
		return generatedReportTemplateDraft{}, fmt.Errorf("%w: %v", ErrReportTemplateDraftFailed, err)
	}

	if response.StatusCode >= http.StatusBadRequest {
		return generatedReportTemplateDraft{}, providerRequestFailed(payload, "AI provider request failed")
	}

	content := strings.TrimSpace(decoded.OutputText)
	if content == "" {
		content = extractOpenAIOutputText(decoded.Output)
	}
	if content == "" {
		return generatedReportTemplateDraft{}, fmt.Errorf("%w: empty model output", ErrReportTemplateDraftFailed)
	}

	return decodeGeneratedReportTemplateDraft(content)
}

func (g *openAIReportTemplateDraftGenerator) generateReportTemplateDraftWithChatCompletions(ctx context.Context, request reportTemplateDraftRequest) (generatedReportTemplateDraft, error) {
	draft, payload, err := g.generateReportTemplateDraftWithChatCompletionsRequest(ctx, request, true, "")
	if err == nil {
		return draft, nil
	}

	if !shouldRetryWithoutStructuredOutput(payload) {
		if !shouldRetryWithOpenRouterFreeModel(payload, g.model) {
			return generatedReportTemplateDraft{}, err
		}
		draft, _, err = g.generateReportTemplateDraftWithChatCompletionsRequest(ctx, request, false, "openrouter/free")
		return draft, err
	}

	draft, payload, err = g.generateReportTemplateDraftWithChatCompletionsRequest(ctx, request, false, "")
	if err == nil {
		return draft, nil
	}

	if !shouldRetryWithOpenRouterFreeModel(payload, g.model) {
		return generatedReportTemplateDraft{}, err
	}

	draft, _, err = g.generateReportTemplateDraftWithChatCompletionsRequest(ctx, request, false, "openrouter/free")
	return draft, err
}

func (g *openAIReportTemplateDraftGenerator) generateReportTemplateDraftWithChatCompletionsRequest(ctx context.Context, request reportTemplateDraftRequest, useStructuredOutput bool, modelOverride string) (generatedReportTemplateDraft, []byte, error) {
	var responseFormat *openAIChatCompletionFormat
	if useStructuredOutput {
		responseFormat = &openAIChatCompletionFormat{
			Type: "json_schema",
			JSONSchema: openAIChatCompletionSchema{
				Name:   "report_template_draft",
				Strict: true,
				Schema: reportTemplateDraftJSONSchema(),
			},
		}
	}
	model := strings.TrimSpace(modelOverride)
	if model == "" {
		model = g.model
	}

	var provider *openRouterProviderPreferences
	if g.usesOpenRouterCompletionsAPI() {
		provider = &openRouterProviderPreferences{
			AllowFallbacks:    true,
			RequireParameters: useStructuredOutput,
		}
	}

	body, err := json.Marshal(openAIChatCompletionsRequest{
		Model: model,
		Messages: []openAIChatMessage{
			{
				Role:    "system",
				Content: buildReportTemplateDraftSystemPrompt(),
			},
			{
				Role:    "user",
				Content: buildReportTemplateDraftUserPrompt(request),
			},
		},
		ResponseFormat:    responseFormat,
		StructuredOutputs: useStructuredOutput,
		Provider:          provider,
		MaxTokens:         1800,
	})
	if err != nil {
		return generatedReportTemplateDraft{}, nil, err
	}

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, g.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return generatedReportTemplateDraft{}, nil, err
	}
	httpRequest.Header.Set("Authorization", "Bearer "+g.apiKey)
	httpRequest.Header.Set("Content-Type", "application/json")

	response, err := g.doChatCompletionsWithRetry(httpRequest)
	if err != nil {
		return generatedReportTemplateDraft{}, nil, providerClientFailed(err, "AI provider request failed")
	}
	defer response.Body.Close()

	payload, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return generatedReportTemplateDraft{}, nil, providerClientFailed(err, "AI provider response read failed")
	}

	var decoded openAIChatCompletionsResponse
	if err := json.Unmarshal(payload, &decoded); err != nil {
		if response.StatusCode >= http.StatusBadRequest {
			return generatedReportTemplateDraft{}, payload, providerRequestFailed(payload, "AI provider request failed")
		}
		return generatedReportTemplateDraft{}, payload, fmt.Errorf("%w: %v", ErrReportTemplateDraftFailed, err)
	}

	if response.StatusCode >= http.StatusBadRequest {
		return generatedReportTemplateDraft{}, payload, providerRequestFailed(payload, "AI provider request failed")
	}

	if len(decoded.Choices) == 0 {
		return generatedReportTemplateDraft{}, payload, fmt.Errorf("%w: empty model output", ErrReportTemplateDraftFailed)
	}

	content := extractChatCompletionText(decoded.Choices[0].Message.Content)
	if content == "" {
		return generatedReportTemplateDraft{}, payload, fmt.Errorf("%w: empty model output", ErrReportTemplateDraftFailed)
	}
	draft, err := decodeGeneratedReportTemplateDraft(content)
	if err != nil {
		return generatedReportTemplateDraft{}, payload, err
	}

	return draft, payload, nil
}

func (g *openAIReportTemplateDraftGenerator) usesOpenRouterCompletionsAPI() bool {
	if g == nil {
		return false
	}

	return strings.Contains(strings.ToLower(strings.TrimSpace(g.baseURL)), "openrouter.ai")
}

func providerRequestFailed(payload []byte, fallback string) error {
	message := extractProviderErrorMessage(payload)
	if message == "" {
		message = fallback
	}
	return fmt.Errorf("%w: %s", ErrReportTemplateDraftFailed, message)
}

func providerClientFailed(err error, fallback string) error {
	if err == nil {
		return fmt.Errorf("%w: %s", ErrReportTemplateDraftFailed, fallback)
	}
	return fmt.Errorf("%w: %s", ErrReportTemplateDraftFailed, strings.TrimSpace(err.Error()))
}

func (g *openAIReportTemplateDraftGenerator) doChatCompletionsWithRetry(request *http.Request) (*http.Response, error) {
	if g == nil || g.httpClient == nil {
		return nil, errors.New("http client is not configured")
	}

	const maxAttempts = 2
	var lastErr error

	for attempt := 0; attempt < maxAttempts; attempt++ {
		cloned := request.Clone(request.Context())
		if request.GetBody != nil {
			body, err := request.GetBody()
			if err != nil {
				return nil, err
			}
			cloned.Body = body
		}

		response, err := g.httpClient.Do(cloned)
		if err != nil {
			lastErr = err
			if attempt == maxAttempts-1 {
				return nil, err
			}
			if !sleepWithContext(request.Context(), 350*time.Millisecond) {
				return nil, request.Context().Err()
			}
			continue
		}

		if !shouldRetryTransientStatus(response.StatusCode) {
			return response, nil
		}

		payload, readErr := io.ReadAll(io.LimitReader(response.Body, 1<<20))
		response.Body.Close()
		if readErr != nil {
			lastErr = readErr
			if attempt == maxAttempts-1 {
				return nil, readErr
			}
			if !sleepWithContext(request.Context(), 350*time.Millisecond) {
				return nil, request.Context().Err()
			}
			continue
		}

		if !shouldRetryTransientProviderFailure(payload) || attempt == maxAttempts-1 {
			response.Body = io.NopCloser(bytes.NewReader(payload))
			return response, nil
		}

		lastErr = providerRequestFailed(payload, "AI provider request failed")
		if !sleepWithContext(request.Context(), 450*time.Millisecond) {
			return nil, request.Context().Err()
		}
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, errors.New("ai provider request failed")
}

func sleepWithContext(ctx context.Context, duration time.Duration) bool {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func shouldRetryTransientStatus(statusCode int) bool {
	return statusCode == http.StatusTooManyRequests ||
		statusCode == http.StatusBadGateway ||
		statusCode == http.StatusServiceUnavailable ||
		statusCode == http.StatusGatewayTimeout
}

func shouldRetryWithoutStructuredOutput(payload []byte) bool {
	message := strings.ToLower(strings.TrimSpace(extractProviderErrorMessage(payload)))
	if message == "" {
		return false
	}

	needles := []string{
		"response_format",
		"json_schema",
		"structured",
		"schema",
		"unsupported",
		"not supported",
		"does not support",
		"invalid format",
	}
	for _, needle := range needles {
		if strings.Contains(message, needle) {
			return true
		}
	}

	return false
}

func shouldRetryWithOpenRouterFreeModel(payload []byte, currentModel string) bool {
	currentModel = strings.ToLower(strings.TrimSpace(currentModel))
	if currentModel == "" || currentModel == "openrouter/free" {
		return false
	}

	message := strings.ToLower(strings.TrimSpace(extractProviderErrorMessage(payload)))
	if message == "" {
		return false
	}

	needles := []string{
		"provider returned error",
		"no endpoints found",
		"no provider available",
		"temporarily unavailable",
		"rate limit",
		"overloaded",
		"upstream",
		"timed out",
		"timeout",
	}
	for _, needle := range needles {
		if strings.Contains(message, needle) {
			return true
		}
	}

	return false
}

func shouldRetryTransientProviderFailure(payload []byte) bool {
	message := strings.ToLower(strings.TrimSpace(extractProviderErrorMessage(payload)))
	if message == "" {
		return false
	}

	needles := []string{
		"rate-limit",
		"rate limit",
		"temporarily rate-limited",
		"retry shortly",
		"temporarily unavailable",
		"provider returned error",
		"upstream",
		"timeout",
		"timed out",
		"overloaded",
	}
	for _, needle := range needles {
		if strings.Contains(message, needle) {
			return true
		}
	}

	return false
}

func extractProviderErrorMessage(payload []byte) string {
	var decoded struct {
		Error *openAIAPIError `json:"error"`
	}
	if err := json.Unmarshal(payload, &decoded); err == nil && decoded.Error != nil {
		message := strings.TrimSpace(decoded.Error.Message)
		raw := extractProviderMetadataText(decoded.Error.Metadata)
		if raw != "" {
			if message == "" || strings.EqualFold(message, "provider returned error") {
				return raw
			}
			return message + ": " + raw
		}
		if message != "" {
			return message
		}
	}

	message := strings.TrimSpace(string(payload))
	if message == "" {
		return ""
	}
	if len(message) > 400 {
		message = message[:400] + "..."
	}
	return message
}

func extractProviderMetadataText(metadata map[string]any) string {
	if len(metadata) == 0 {
		return ""
	}

	raw := strings.TrimSpace(anyToCompactString(metadata["raw"]))
	providerName := strings.TrimSpace(anyToCompactString(metadata["provider_name"]))
	if raw != "" && providerName != "" {
		return providerName + ": " + raw
	}
	if raw != "" {
		return raw
	}
	if providerName != "" {
		return providerName
	}

	return ""
}

func anyToCompactString(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(typed)
	default:
		payload, err := json.Marshal(typed)
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(payload))
	}
}

func extractChatCompletionText(content any) string {
	switch value := content.(type) {
	case string:
		return strings.TrimSpace(value)
	case []any:
		var builder strings.Builder
		for _, item := range value {
			part, ok := item.(map[string]any)
			if !ok {
				continue
			}
			text, _ := part["text"].(string)
			text = strings.TrimSpace(text)
			if text == "" {
				continue
			}
			if builder.Len() > 0 {
				builder.WriteByte('\n')
			}
			builder.WriteString(text)
		}
		return strings.TrimSpace(builder.String())
	default:
		return ""
	}
}

func extractJSONObject(content string) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return ""
	}

	if strings.HasPrefix(content, "```") {
		content = strings.TrimSpace(strings.TrimPrefix(content, "```json"))
		content = strings.TrimSpace(strings.TrimPrefix(content, "```"))
		content = strings.TrimSpace(strings.TrimSuffix(content, "```"))
	}

	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start >= 0 && end >= start {
		return strings.TrimSpace(content[start : end+1])
	}

	return content
}

func decodeGeneratedReportTemplateDraft(content string) (generatedReportTemplateDraft, error) {
	candidates := uniqueNonEmptyStrings(
		strings.TrimSpace(content),
		extractJSONObject(content),
		repairPotentialJSON(content),
		repairPotentialJSON(extractJSONObject(content)),
	)

	for _, candidate := range candidates {
		if draft, ok := parseGeneratedReportTemplateDraftFromJSON(candidate); ok {
			sanitizeGeneratedReportTemplateDraft(&draft)
			return draft, nil
		}
	}

	if draft, ok := parseGeneratedReportTemplateDraftFromText(content); ok {
		sanitizeGeneratedReportTemplateDraft(&draft)
		return draft, nil
	}

	return generatedReportTemplateDraft{}, fmt.Errorf("%w: unable to coerce model output into template draft json", ErrReportTemplateDraftFailed)
}

func parseGeneratedReportTemplateDraftFromJSON(candidate string) (generatedReportTemplateDraft, bool) {
	candidate = strings.TrimSpace(candidate)
	if candidate == "" {
		return generatedReportTemplateDraft{}, false
	}

	var direct generatedReportTemplateDraft
	if err := json.Unmarshal([]byte(candidate), &direct); err == nil && generatedReportTemplateDraftHasContent(direct) {
		return direct, true
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(candidate), &payload); err != nil {
		return generatedReportTemplateDraft{}, false
	}

	draft := generatedReportTemplateDraft{}
	for key, value := range payload {
		applyLooseDraftField(&draft, normalizeLooseKey(key), value)
	}

	if !generatedReportTemplateDraftHasContent(draft) {
		return generatedReportTemplateDraft{}, false
	}

	return draft, true
}

func parseGeneratedReportTemplateDraftFromText(content string) (generatedReportTemplateDraft, bool) {
	content = sanitizeLooseText(content)
	if content == "" {
		return generatedReportTemplateDraft{}, false
	}

	paragraphs := splitIntoParagraphs(content)
	lines := strings.Split(content, "\n")

	draft := generatedReportTemplateDraft{}
	currentAudience := reportAudience("")

	for _, rawLine := range lines {
		line := normalizeLooseLine(rawLine)
		if line == "" {
			continue
		}

		if audience, ok := detectAudienceHeading(line); ok {
			currentAudience = audience
			continue
		}

		key, value, ok := splitLooseKeyValue(line)
		if !ok {
			if currentAudience == reportAudienceClient {
				appendIntroParagraph(&draft.Client, line)
			} else if currentAudience == reportAudiencePsychologist {
				appendIntroParagraph(&draft.Psychologist, line)
			}
			continue
		}

		normalizedKey := normalizeLooseKey(key)
		if currentAudience != "" {
			target := &draft.Client
			if currentAudience == reportAudiencePsychologist {
				target = &draft.Psychologist
			}
			applyLooseAudienceField(target, normalizedKey, value, currentAudience)
			continue
		}

		switch normalizedKey {
		case "name", "title", "template_name", "report_name", "название":
			if draft.Name == "" {
				draft.Name = strings.TrimSpace(value)
			}
		case "description", "template_description", "описание":
			if draft.Description == "" {
				draft.Description = strings.TrimSpace(value)
			}
		default:
			applyLooseDraftField(&draft, normalizedKey, value)
		}
	}

	if draft.Name == "" && len(paragraphs) > 0 {
		draft.Name = deriveDraftNameFromText(paragraphs[0])
	}
	if draft.Description == "" {
		draft.Description = deriveDraftDescriptionFromParagraphs(paragraphs)
	}
	if len(draft.Client.IntroParagraphs) == 0 && len(paragraphs) > 0 {
		draft.Client.IntroParagraphs = takeParagraphs(paragraphs, 2)
	}
	if len(draft.Psychologist.IntroParagraphs) == 0 && len(paragraphs) > 0 {
		start := 0
		if len(paragraphs) > 1 {
			start = 1
		}
		draft.Psychologist.IntroParagraphs = takeParagraphs(paragraphs[start:], 2)
	}

	if !generatedReportTemplateDraftHasContent(draft) {
		return generatedReportTemplateDraft{}, false
	}

	return draft, true
}

func generatedReportTemplateDraftHasContent(draft generatedReportTemplateDraft) bool {
	return strings.TrimSpace(draft.Name) != "" ||
		strings.TrimSpace(draft.Description) != "" ||
		reportAudienceTemplateConfigHasContent(draft.Client) ||
		reportAudienceTemplateConfigHasContent(draft.Psychologist)
}

func reportAudienceTemplateConfigHasContent(config reportAudienceTemplateConfig) bool {
	return strings.TrimSpace(config.Title) != "" ||
		strings.TrimSpace(config.ChartCaption) != "" ||
		len(config.MetaLabels) > 0 ||
		len(config.SectionTitles) > 0 ||
		len(config.IntroParagraphs) > 0 ||
		len(config.ClosingParagraphs) > 0
}

func repairPotentialJSON(content string) string {
	content = sanitizeLooseText(content)
	if content == "" {
		return ""
	}

	content = replaceSingleQuotedJSONTokens(content)
	content = quoteLooseJSONKeys(content)
	content = trailingCommaPattern.ReplaceAllString(content, "$1")
	content = pythonBooleanPattern.ReplaceAllStringFunc(content, func(value string) string {
		switch value {
		case "True":
			return "true"
		case "False":
			return "false"
		case "None":
			return "null"
		default:
			return value
		}
	})

	return strings.TrimSpace(content)
}

func replaceSingleQuotedJSONTokens(content string) string {
	matches := singleQuotedJSONTokenPattern.FindAllStringSubmatchIndex(content, -1)
	if len(matches) == 0 {
		return content
	}

	var builder strings.Builder
	last := 0
	for _, match := range matches {
		builder.WriteString(content[last:match[0]])
		value := content[match[2]:match[3]]
		encoded, err := json.Marshal(value)
		if err != nil {
			builder.WriteString(content[match[0]:match[1]])
		} else {
			builder.WriteString(string(encoded))
		}
		last = match[1]
	}
	builder.WriteString(content[last:])
	return builder.String()
}

func quoteLooseJSONKeys(content string) string {
	lines := strings.Split(content, "\n")
	for idx, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		lines[idx] = quoteLooseJSONKeysInLine(line)
	}
	return strings.Join(lines, "\n")
}

func quoteLooseJSONKeysInLine(line string) string {
	var builder strings.Builder
	last := 0
	matches := unquotedJSONKeyPattern.FindAllStringSubmatchIndex(line, -1)
	if len(matches) == 0 {
		return line
	}

	for _, match := range matches {
		builder.WriteString(line[last:match[0]])
		prefix := line[match[2]:match[3]]
		key := line[match[4]:match[5]]
		suffix := line[match[6]:match[7]]
		builder.WriteString(prefix)
		builder.WriteString(`"`)
		builder.WriteString(key)
		builder.WriteString(`"`)
		builder.WriteString(suffix)
		last = match[1]
	}
	builder.WriteString(line[last:])
	return builder.String()
}

func sanitizeLooseText(content string) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return ""
	}

	replacements := map[string]string{
		"```json": "",
		"```JSON": "",
		"```":     "",
		"“":       `"`,
		"”":       `"`,
		"„":       `"`,
		"’":       `'`,
		"‘":       `'`,
	}
	for oldValue, newValue := range replacements {
		content = strings.ReplaceAll(content, oldValue, newValue)
	}

	return strings.TrimSpace(content)
}

func uniqueNonEmptyStrings(values ...string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

func normalizeLooseKey(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	replacer := strings.NewReplacer(" ", "_", "-", "_", ".", "_", "/", "_")
	value = replacer.Replace(value)
	value = strings.Trim(value, "_:;")
	return value
}

func normalizeLooseLine(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimLeft(value, "-*# ")
	return strings.TrimSpace(value)
}

func detectAudienceHeading(line string) (reportAudience, bool) {
	normalized := normalizeLooseKey(line)
	switch normalized {
	case "client", "for_client", "client_block", "клиент", "для_клиента":
		return reportAudienceClient, true
	case "psychologist", "for_psychologist", "psychologist_block", "психолог", "для_психолога":
		return reportAudiencePsychologist, true
	default:
		return "", false
	}
}

func splitLooseKeyValue(line string) (string, string, bool) {
	for _, separator := range []string{":", " - ", " — ", " – "} {
		parts := strings.SplitN(line, separator, 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if key != "" && value != "" {
				return key, value, true
			}
		}
	}
	return "", "", false
}

func applyLooseDraftField(draft *generatedReportTemplateDraft, key string, value any) {
	if draft == nil {
		return
	}

	switch key {
	case "name", "title", "template_name", "report_name":
		if draft.Name == "" {
			draft.Name = stringifyLooseValue(value)
		}
	case "description", "template_description", "overview":
		if draft.Description == "" {
			draft.Description = stringifyLooseValue(value)
		}
	case "client", "for_client", "client_template", "client_report", "respondent":
		if payload, ok := value.(map[string]any); ok {
			applyLooseAudienceMap(&draft.Client, payload, reportAudienceClient)
		}
	case "psychologist", "for_psychologist", "psychologist_template", "psychologist_report", "specialist":
		if payload, ok := value.(map[string]any); ok {
			applyLooseAudienceMap(&draft.Psychologist, payload, reportAudiencePsychologist)
		}
	default:
		if strings.HasPrefix(key, "client_") {
			applyLooseAudienceField(&draft.Client, strings.TrimPrefix(key, "client_"), value, reportAudienceClient)
		}
		if strings.HasPrefix(key, "psychologist_") {
			applyLooseAudienceField(&draft.Psychologist, strings.TrimPrefix(key, "psychologist_"), value, reportAudiencePsychologist)
		}
	}
}

func applyLooseAudienceMap(config *reportAudienceTemplateConfig, payload map[string]any, audience reportAudience) {
	if config == nil {
		return
	}

	for key, value := range payload {
		applyLooseAudienceField(config, normalizeLooseKey(key), value, audience)
	}
}

func applyLooseAudienceField(config *reportAudienceTemplateConfig, key string, value any, audience reportAudience) {
	if config == nil {
		return
	}

	switch key {
	case "title", "header", "report_title", "заголовок", "название_блока":
		if config.Title == "" {
			config.Title = stringifyLooseValue(value)
		}
	case "chart_caption", "diagram_caption", "chart_description", "подпись_к_диаграмме", "подпись_диаграммы":
		if config.ChartCaption == "" {
			config.ChartCaption = stringifyLooseValue(value)
		}
	case "intro", "introduction", "intro_text", "intro_paragraphs", "opening", "opening_paragraphs", "вступление", "введение":
		for _, paragraph := range looseParagraphs(value) {
			appendIntroParagraph(config, paragraph)
		}
	case "closing", "closing_paragraphs", "conclusion", "заключение":
		for _, paragraph := range looseParagraphs(value) {
			appendClosingParagraph(config, paragraph)
		}
	case "meta_labels", "meta", "labels", "метки":
		for metaKey, metaValue := range looseStringMap(value) {
			if config.MetaLabels == nil {
				config.MetaLabels = map[string]string{}
			}
			config.MetaLabels[normalizeLooseKey(metaKey)] = metaValue
		}
	case "section_titles", "sections", "titles", "заголовки_секций":
		for sectionKey, sectionValue := range looseStringMap(value) {
			setLooseSectionTitle(config, normalizeLooseKey(sectionKey), sectionValue, audience)
		}
	default:
		if stringValue := stringifyLooseValue(value); stringValue != "" {
			setLooseSectionTitle(config, key, stringValue, audience)
		}
	}
}

func setLooseSectionTitle(config *reportAudienceTemplateConfig, key string, value string, audience reportAudience) {
	if config == nil || strings.TrimSpace(value) == "" {
		return
	}
	if config.SectionTitles == nil {
		config.SectionTitles = map[string]string{}
	}

	switch key {
	case "summary", "summary_title", "short_summary", "краткий_вывод", "вывод":
		if audience == reportAudienceClient {
			config.SectionTitles[reportSectionSummary] = value
		}
	case "chart_data", "chart_title", "diagram_title", "diagram", "диаграмма", "график":
		if audience == reportAudienceClient {
			config.SectionTitles[reportSectionChartData] = value
		}
	case "results", "results_title", "results_block", "result_block", "metrics", "metrics_title", "scales_list", "scales", "результаты", "блок_результатов", "метрики":
		config.SectionTitles[reportSectionScalesList] = value
	case "interpretation", "interpretation_title", "интерпретация":
		if audience == reportAudienceClient {
			config.SectionTitles[reportSectionInterpretation] = value
		}
	case "recommendations", "recommendations_title", "next_steps", "рекомендации":
		if audience == reportAudienceClient {
			config.SectionTitles[reportSectionRecommendations] = value
		}
	case "raw_scores", "raw_scores_title", "raw", "raw_results", "сырые_показатели", "сырые_баллы":
		if audience == reportAudiencePsychologist {
			config.SectionTitles[reportSectionRawScores] = value
		}
	case "answers", "answers_table", "answers_table_title", "response_table", "таблица_ответов", "ответы_респондента":
		if audience == reportAudiencePsychologist {
			config.SectionTitles[reportSectionAnswersTable] = value
		}
	case "intro", "вступление":
		config.SectionTitles[reportSectionIntro] = value
	case "closing", "заключение":
		config.SectionTitles[reportSectionClosing] = value
	}
}

func stringifyLooseValue(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case float64:
		return strings.TrimSpace(fmt.Sprintf("%.0f", typed))
	case int:
		return strings.TrimSpace(fmt.Sprintf("%d", typed))
	case bool:
		if typed {
			return "true"
		}
		return "false"
	default:
		return ""
	}
}

func looseParagraphs(value any) []string {
	switch typed := value.(type) {
	case string:
		return splitIntoParagraphs(typed)
	case []any:
		result := make([]string, 0, len(typed))
		for _, item := range typed {
			if paragraph := stringifyLooseValue(item); paragraph != "" {
				result = append(result, paragraph)
			}
		}
		return result
	default:
		return nil
	}
}

func looseStringMap(value any) map[string]string {
	payload, ok := value.(map[string]any)
	if !ok {
		return nil
	}

	result := make(map[string]string, len(payload))
	for key, rawValue := range payload {
		if stringValue := stringifyLooseValue(rawValue); stringValue != "" {
			result[key] = stringValue
		}
	}
	return result
}

func appendIntroParagraph(config *reportAudienceTemplateConfig, value string) {
	value = strings.TrimSpace(value)
	if config == nil || value == "" {
		return
	}
	config.IntroParagraphs = append(config.IntroParagraphs, value)
}

func appendClosingParagraph(config *reportAudienceTemplateConfig, value string) {
	value = strings.TrimSpace(value)
	if config == nil || value == "" {
		return
	}
	config.ClosingParagraphs = append(config.ClosingParagraphs, value)
}

func splitIntoParagraphs(content string) []string {
	content = sanitizeLooseText(content)
	if content == "" {
		return nil
	}

	chunks := strings.Split(content, "\n\n")
	result := make([]string, 0, len(chunks))
	for _, chunk := range chunks {
		paragraph := strings.TrimSpace(strings.ReplaceAll(chunk, "\n", " "))
		if paragraph != "" {
			result = append(result, paragraph)
		}
	}
	return result
}

func deriveDraftNameFromText(content string) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return ""
	}

	if len(content) > 80 {
		content = content[:80]
	}
	content = strings.TrimSpace(content)
	if content == "" {
		return ""
	}
	return content
}

func deriveDraftDescriptionFromParagraphs(paragraphs []string) string {
	if len(paragraphs) == 0 {
		return ""
	}

	description := strings.TrimSpace(paragraphs[0])
	if len(paragraphs) > 1 {
		description += " " + strings.TrimSpace(paragraphs[1])
	}
	if len(description) > 280 {
		description = description[:280]
	}
	return strings.TrimSpace(description)
}

func takeParagraphs(paragraphs []string, limit int) []string {
	if limit <= 0 || len(paragraphs) == 0 {
		return nil
	}
	if len(paragraphs) < limit {
		limit = len(paragraphs)
	}
	result := make([]string, 0, limit)
	for idx := 0; idx < limit; idx++ {
		value := strings.TrimSpace(paragraphs[idx])
		if value != "" {
			result = append(result, value)
		}
	}
	return result
}

func extractOpenAIOutputText(items []openAIOutputItem) string {
	for _, item := range items {
		for _, content := range item.Content {
			if content.Type == "output_text" || content.Type == "text" {
				text := strings.TrimSpace(content.Text)
				if text != "" {
					return text
				}
			}
		}
	}

	return ""
}

func buildReportTemplateDraftSystemPrompt() string {
	return strings.TrimSpace(`Ты помогаешь психологу подготовить черновик шаблона отчета для сервиса онлайн-тестирования.

Сгенерируй только JSON по заданной схеме.

Требования:
- Пиши весь контент на русском языке.
- Не используй markdown, комментарии и объяснения вне JSON.
- Не оставляй пустые строки, пустые массивы и пустые объекты, если поле можно осмысленно заполнить.
- Сформируй черновик, который психолог потом сможет отредактировать вручную.
- Поле "name" должно быть коротким и понятным.
- Поле "description" должно кратко объяснять, для какого теста или сценария подойдет шаблон.
- Для блока "client" используй более мягкий и понятный тон.
- Для блока "psychologist" используй более технический и рабочий тон.
- Не придумывай новых полей кроме тех, что разрешены схемой.
- section_titles и meta_labels делай короткими и прикладными.
- intro_paragraphs и closing_paragraphs должны быть лаконичными и практически полезными.
- Для блока "client" обязательно заполни:
  title, meta_labels.respondent, meta_labels.session, meta_labels.status,
  intro_paragraphs, closing_paragraphs, chart_caption,
  section_titles.summary, section_titles.chart_data, section_titles.scales_list,
  section_titles.interpretation, section_titles.recommendations.
- Для блока "psychologist" обязательно заполни:
  title, meta_labels.respondent, meta_labels.session, meta_labels.status,
  meta_labels.top_scales, meta_labels.top_professions,
  intro_paragraphs, closing_paragraphs,
  section_titles.scales_list, section_titles.raw_scores, section_titles.answers_table.
- Заголовок блока результатов не оставляй пустым.
- Подпись к диаграмме не оставляй пустой.
- В каждом блоке сделай минимум два осмысленных вступительных абзаца.
- Для client и psychologist верни полный, законченный черновик, а не минимальный набросок.
- Если контекст теста неполный, все равно создай универсальный качественный черновик.`)
}

func buildReportTemplateDraftUserPrompt(request reportTemplateDraftRequest) string {
	var builder strings.Builder
	builder.WriteString("Промпт психолога:\n")
	builder.WriteString(strings.TrimSpace(request.Prompt))

	if strings.TrimSpace(request.TestContext) != "" {
		builder.WriteString("\n\nКонтекст теста:\n")
		builder.WriteString(strings.TrimSpace(request.TestContext))
	}

	builder.WriteString("\n\nСобери такой черновик шаблона, который можно будет сразу показать психологу в интерфейсе и при необходимости сохранить почти без правок.")
	builder.WriteString("\n\nЗаполни именно этот JSON-каркас, сохраняя те же ключи:")
	builder.WriteString("\n{\n")
	builder.WriteString(`  "name": "Короткое название шаблона",` + "\n")
	builder.WriteString(`  "description": "Короткое описание сценария применения шаблона",` + "\n")
	builder.WriteString(`  "client": {` + "\n")
	builder.WriteString(`    "title": "Заголовок клиентского отчета",` + "\n")
	builder.WriteString(`    "meta_labels": {` + "\n")
	builder.WriteString(`      "respondent": "Клиент",` + "\n")
	builder.WriteString(`      "session": "Сессия",` + "\n")
	builder.WriteString(`      "status": "Статус"` + "\n")
	builder.WriteString("    },\n")
	builder.WriteString(`    "section_titles": {` + "\n")
	builder.WriteString(`      "summary": "Краткий вывод",` + "\n")
	builder.WriteString(`      "chart_data": "Профиль результата",` + "\n")
	builder.WriteString(`      "scales_list": "Результаты",` + "\n")
	builder.WriteString(`      "interpretation": "Интерпретация",` + "\n")
	builder.WriteString(`      "recommendations": "Рекомендации"` + "\n")
	builder.WriteString("    },\n")
	builder.WriteString(`    "intro_paragraphs": ["Первый абзац", "Второй абзац"],` + "\n")
	builder.WriteString(`    "closing_paragraphs": ["Итоговый абзац"],` + "\n")
	builder.WriteString(`    "chart_caption": "Подпись к диаграмме"` + "\n")
	builder.WriteString("  },\n")
	builder.WriteString(`  "psychologist": {` + "\n")
	builder.WriteString(`    "title": "Заголовок рабочего отчета психолога",` + "\n")
	builder.WriteString(`    "meta_labels": {` + "\n")
	builder.WriteString(`      "respondent": "Респондент",` + "\n")
	builder.WriteString(`      "session": "Сессия",` + "\n")
	builder.WriteString(`      "status": "Статус",` + "\n")
	builder.WriteString(`      "top_scales": "Топ-шкалы",` + "\n")
	builder.WriteString(`      "top_professions": "Топ-профессии"` + "\n")
	builder.WriteString("    },\n")
	builder.WriteString(`    "section_titles": {` + "\n")
	builder.WriteString(`      "scales_list": "Ключевые результаты",` + "\n")
	builder.WriteString(`      "raw_scores": "Сырые показатели",` + "\n")
	builder.WriteString(`      "answers_table": "Таблица ответов"` + "\n")
	builder.WriteString("    },\n")
	builder.WriteString(`    "intro_paragraphs": ["Первый абзац", "Второй абзац"],` + "\n")
	builder.WriteString(`    "closing_paragraphs": ["Итоговый абзац"]` + "\n")
	builder.WriteString("  }\n")
	builder.WriteString("}\n")
	builder.WriteString("\nНе сокращай структуру. Не пропускай обязательные подписи и заголовки. Верни только один JSON-объект.")
	return builder.String()
}

func reportTemplateDraftJSONSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"name", "description", "client", "psychologist"},
		"properties": map[string]any{
			"name": map[string]any{
				"type":      "string",
				"minLength": 1,
			},
			"description": map[string]any{
				"type": "string",
			},
			"client":       reportAudienceTemplateSchema(),
			"psychologist": reportAudienceTemplateSchema(),
		},
	}
}

func reportAudienceTemplateSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"title": map[string]any{
				"type": "string",
			},
			"meta_labels": map[string]any{
				"type": "object",
				"additionalProperties": map[string]any{
					"type": "string",
				},
			},
			"section_titles": map[string]any{
				"type": "object",
				"additionalProperties": map[string]any{
					"type": "string",
				},
			},
			"intro_paragraphs": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
				},
			},
			"closing_paragraphs": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
				},
			},
			"chart_caption": map[string]any{
				"type": "string",
			},
		},
	}
}

func (s *AppService) GenerateReportTemplateDraft(ctx context.Context, user domain.AuthenticatedUser, input domain.GenerateReportTemplateDraftInput) (domain.GenerateReportTemplateDraftResponse, error) {
	input.Prompt = strings.TrimSpace(input.Prompt)
	if input.Prompt == "" || input.TestID < 0 {
		return domain.GenerateReportTemplateDraftResponse{}, ErrInvalidReportTemplateInput
	}
	if !domain.IsProSubscriptionPlan(user.SubscriptionPlan) {
		return domain.GenerateReportTemplateDraftResponse{}, ErrReportTemplateDraftForbidden
	}
	if s.reportTemplateDraftGenerator == nil || !s.reportTemplateDraftGenerator.Enabled() {
		return domain.GenerateReportTemplateDraftResponse{}, ErrReportTemplateDraftUnavailable
	}

	request := reportTemplateDraftRequest{
		Prompt: input.Prompt,
	}

	if input.TestID > 0 {
		test, err := s.GetPsychologistTestByID(ctx, user.ID, input.TestID)
		if err != nil {
			return domain.GenerateReportTemplateDraftResponse{}, err
		}

		questions, err := s.ListPsychologistQuestions(ctx, user.ID, input.TestID)
		if err != nil {
			return domain.GenerateReportTemplateDraftResponse{}, err
		}

		rules, err := s.ListFormulaRules(ctx, user.ID, input.TestID)
		if err != nil {
			return domain.GenerateReportTemplateDraftResponse{}, err
		}

		request.TestContext = buildReportTemplateDraftTestContext(test, questions, rules)
	}

	draft, err := s.reportTemplateDraftGenerator.GenerateReportTemplateDraft(ctx, request)
	if err != nil {
		return domain.GenerateReportTemplateDraftResponse{}, err
	}

	response, err := normalizeGeneratedReportTemplateDraft(draft)
	if err != nil {
		return domain.GenerateReportTemplateDraftResponse{}, fmt.Errorf("%w: %v", ErrReportTemplateDraftFailed, err)
	}
	response.Model = s.reportTemplateDraftGenerator.Model()

	return response, nil
}

func normalizeGeneratedReportTemplateDraft(draft generatedReportTemplateDraft) (domain.GenerateReportTemplateDraftResponse, error) {
	sanitizeGeneratedReportTemplateDraft(&draft)

	rawConfig, err := json.Marshal(reportTemplateConfig{
		Client:       draft.Client,
		Psychologist: draft.Psychologist,
	})
	if err != nil {
		return domain.GenerateReportTemplateDraftResponse{}, err
	}

	input, err := normalizeCreateReportTemplateInput(domain.CreateReportTemplateInput{
		Name:         draft.Name,
		Description:  draft.Description,
		TemplateBody: string(rawConfig),
	})
	if err != nil {
		return domain.GenerateReportTemplateDraftResponse{}, err
	}

	return domain.GenerateReportTemplateDraftResponse{
		Name:         input.Name,
		Description:  input.Description,
		TemplateBody: input.TemplateBody,
	}, nil
}

func sanitizeGeneratedReportTemplateDraft(draft *generatedReportTemplateDraft) {
	if draft == nil {
		return
	}

	draft.Name = strings.TrimSpace(draft.Name)
	draft.Description = strings.TrimSpace(draft.Description)
	if draft.Name == "" {
		draft.Name = "AI шаблон отчета"
	}
	if draft.Description == "" {
		draft.Description = "Черновик шаблона отчета, подготовленный нейросетью для последующей ручной доработки психологом."
	}

	sanitizeGeneratedReportAudienceTemplateConfig(&draft.Client, reportAudienceClient)
	sanitizeGeneratedReportAudienceTemplateConfig(&draft.Psychologist, reportAudiencePsychologist)
}

func sanitizeGeneratedReportAudienceTemplateConfig(config *reportAudienceTemplateConfig, audience reportAudience) {
	if config == nil {
		return
	}

	normalizeReportAudienceTemplateConfig(config)

	if len(config.MetaLabels) > 0 {
		filtered := make(map[string]string, len(config.MetaLabels))
		for key, value := range config.MetaLabels {
			if resolvedKey := resolveGeneratedMetaKey(key, audience); resolvedKey != "" {
				filtered[resolvedKey] = value
			}
		}
		config.MetaLabels = filtered
	}

	if len(config.SectionTitles) > 0 {
		filtered := make(map[string]string, len(config.SectionTitles))
		for key, value := range config.SectionTitles {
			setGeneratedSectionTitleAlias(filtered, key, value, audience)
		}
		config.SectionTitles = filtered
	}

	applyDefaultGeneratedReportAudienceTemplateConfig(config, audience)
}

func resolveGeneratedMetaKey(key string, audience reportAudience) string {
	normalized := normalizeLooseKey(key)
	allowed := allowedReportMetaKeys(audience)
	if _, ok := allowed[normalized]; ok {
		return normalized
	}

	switch normalized {
	case "respondent_name", "client":
		if _, ok := allowed[reportMetaRespondent]; ok {
			return reportMetaRespondent
		}
	case "session_id":
		if _, ok := allowed[reportMetaSession]; ok {
			return reportMetaSession
		}
	case "state":
		if _, ok := allowed[reportMetaStatus]; ok {
			return reportMetaStatus
		}
	case "top_metrics", "top_results":
		if _, ok := allowed[reportMetaTopScales]; ok {
			return reportMetaTopScales
		}
	}

	return ""
}

func setGeneratedSectionTitleAlias(target map[string]string, key string, value string, audience reportAudience) {
	if target == nil || strings.TrimSpace(value) == "" {
		return
	}

	temp := reportAudienceTemplateConfig{SectionTitles: map[string]string{}}
	setLooseSectionTitle(&temp, normalizeLooseKey(key), value, audience)
	for resolvedKey, resolvedValue := range temp.SectionTitles {
		target[resolvedKey] = resolvedValue
	}

	allowed := allowedReportSectionKeys(audience)
	normalized := normalizeLooseKey(key)
	if _, ok := allowed[normalized]; ok {
		target[normalized] = value
	}
}

func applyDefaultGeneratedReportAudienceTemplateConfig(config *reportAudienceTemplateConfig, audience reportAudience) {
	if config == nil {
		return
	}

	defaults := defaultGeneratedReportAudienceTemplateConfig(audience)

	if strings.TrimSpace(config.Title) == "" {
		config.Title = defaults.Title
	}
	if len(config.IntroParagraphs) == 0 {
		config.IntroParagraphs = append([]string(nil), defaults.IntroParagraphs...)
	}
	if len(config.ClosingParagraphs) == 0 && len(defaults.ClosingParagraphs) > 0 {
		config.ClosingParagraphs = append([]string(nil), defaults.ClosingParagraphs...)
	}
	if strings.TrimSpace(config.ChartCaption) == "" && strings.TrimSpace(defaults.ChartCaption) != "" {
		config.ChartCaption = defaults.ChartCaption
	}

	if config.MetaLabels == nil {
		config.MetaLabels = map[string]string{}
	}
	for key, value := range defaults.MetaLabels {
		if strings.TrimSpace(config.MetaLabels[key]) == "" {
			config.MetaLabels[key] = value
		}
	}

	if config.SectionTitles == nil {
		config.SectionTitles = map[string]string{}
	}
	for key, value := range defaults.SectionTitles {
		if strings.TrimSpace(config.SectionTitles[key]) == "" {
			config.SectionTitles[key] = value
		}
	}

	if len(defaults.IntroParagraphs) > 0 && len(config.IntroParagraphs) < len(defaults.IntroParagraphs) {
		config.IntroParagraphs = appendMissingParagraphs(config.IntroParagraphs, defaults.IntroParagraphs, len(defaults.IntroParagraphs))
	}
}

func defaultGeneratedReportAudienceTemplateConfig(audience reportAudience) reportAudienceTemplateConfig {
	if audience == reportAudiencePsychologist {
		return reportAudienceTemplateConfig{
			Title: "Рабочий отчет психолога",
			MetaLabels: map[string]string{
				reportMetaRespondent:     "Респондент",
				reportMetaSession:        "Сессия",
				reportMetaStatus:         "Статус",
				reportMetaTopScales:      "Топ-шкалы",
				reportMetaTopProfessions: "Топ-профессии",
			},
			SectionTitles: map[string]string{
				reportSectionScalesList:   "Ключевые результаты",
				reportSectionRawScores:    "Сырые показатели",
				reportSectionAnswersTable: "Таблица ответов",
			},
			IntroParagraphs: []string{
				"Этот черновик предназначен для профессионального разбора результатов и дальнейшей рабочей интерпретации на консультации.",
				"Используйте показатели, сырые значения и таблицу ответов как основу для обсуждения гипотез, уточняющих вопросов и рекомендаций клиенту.",
			},
		}
	}

	return reportAudienceTemplateConfig{
		Title: "Отчет по результатам теста",
		MetaLabels: map[string]string{
			reportMetaRespondent: "Респондент",
			reportMetaSession:    "Сессия",
			reportMetaStatus:     "Статус",
		},
		SectionTitles: map[string]string{
			reportSectionSummary:         "Краткий вывод",
			reportSectionChartData:       "Профиль результата",
			reportSectionScalesList:      "Результаты",
			reportSectionInterpretation:  "Интерпретация",
			reportSectionRecommendations: "Рекомендации",
		},
		IntroParagraphs: []string{
			"Этот отчет помогает спокойно посмотреть на результаты текущего прохождения и выделить наиболее заметные особенности профиля.",
			"Итоговые показатели лучше использовать как ориентир для обсуждения и следующих шагов, а не как жесткий окончательный вердикт.",
		},
		ChartCaption: "Диаграмма показывает наиболее выраженные результаты текущего прохождения и помогает быстро увидеть их соотношение.",
	}
}

func copyStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}

	result := make(map[string]string, len(values))
	for key, value := range values {
		result[key] = value
	}
	return result
}

func appendMissingParagraphs(current []string, defaults []string, limit int) []string {
	result := append([]string(nil), current...)
	if limit <= 0 {
		return result
	}

	seen := make(map[string]struct{}, len(result))
	for _, paragraph := range result {
		trimmed := strings.TrimSpace(paragraph)
		if trimmed != "" {
			seen[trimmed] = struct{}{}
		}
	}

	for _, paragraph := range defaults {
		if len(result) >= limit {
			break
		}
		trimmed := strings.TrimSpace(paragraph)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		result = append(result, trimmed)
		seen[trimmed] = struct{}{}
	}

	return result
}

func buildReportTemplateDraftTestContext(test domain.Test, questions []domain.Question, rules []domain.FormulaRule) string {
	lines := []string{
		fmt.Sprintf("Название теста: %s", strings.TrimSpace(test.Title)),
		fmt.Sprintf("Описание теста: %s", defaultIfEmpty(strings.TrimSpace(test.Description), "не указано")),
		fmt.Sprintf("Рекомендуемая длительность: %d мин.", test.RecommendedDuration),
		fmt.Sprintf("Количество вопросов: %d", len(questions)),
	}

	metricKeys := collectTemplateMetricKeys(questions, rules)
	if len(metricKeys) > 0 {
		lines = append(lines, "Ключи метрик/результатов: "+strings.Join(metricKeys, ", "))
	}

	if len(questions) > 0 {
		lines = append(lines, "Вопросы:")
		limit := len(questions)
		if limit > 12 {
			limit = 12
		}
		for idx := 0; idx < limit; idx++ {
			question := questions[idx]
			lines = append(lines, fmt.Sprintf("- #%d [%s] %s", question.OrderNumber, question.QuestionType, strings.TrimSpace(question.Text)))
		}
		if len(questions) > limit {
			lines = append(lines, fmt.Sprintf("- ... и еще %d вопросов", len(questions)-limit))
		}
	}

	return strings.Join(lines, "\n")
}

func collectTemplateMetricKeys(questions []domain.Question, rules []domain.FormulaRule) []string {
	seen := map[string]struct{}{}

	for _, question := range questions {
		for key := range question.ScaleWeights {
			normalized := strings.TrimSpace(key)
			if normalized != "" {
				seen[normalized] = struct{}{}
			}
		}
	}

	for _, rule := range rules {
		normalized := strings.TrimSpace(rule.ResultKey)
		if normalized == "" {
			normalized = "total"
		}
		seen[normalized] = struct{}{}
	}

	keys := make([]string, 0, len(seen))
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	return keys
}

func defaultIfEmpty(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
