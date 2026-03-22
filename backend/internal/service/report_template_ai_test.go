package service

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
)

type fakeReportTemplateDraftGenerator struct {
	enabled bool
	model   string
	draft   generatedReportTemplateDraft
	err     error
}

func (g fakeReportTemplateDraftGenerator) Enabled() bool {
	return g.enabled
}

func (g fakeReportTemplateDraftGenerator) Model() string {
	return g.model
}

func (g fakeReportTemplateDraftGenerator) GenerateReportTemplateDraft(ctx context.Context, request reportTemplateDraftRequest) (generatedReportTemplateDraft, error) {
	return g.draft, g.err
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func TestGenerateReportTemplateDraftRequiresProPlan(t *testing.T) {
	service := &AppService{
		reportTemplateDraftGenerator: fakeReportTemplateDraftGenerator{
			enabled: true,
			model:   "gpt-5-mini",
		},
	}

	_, err := service.GenerateReportTemplateDraft(context.Background(), domain.AuthenticatedUser{
		ID:               7,
		Role:             domain.RolePsychologist,
		SubscriptionPlan: domain.SubscriptionPlanBasic,
	}, domain.GenerateReportTemplateDraftInput{
		Prompt: "Сделай шаблон для профориентационного теста",
	})
	if err != ErrReportTemplateDraftForbidden {
		t.Fatalf("expected ErrReportTemplateDraftForbidden, got %v", err)
	}
}

func TestGenerateReportTemplateDraftReturnsUnavailableWhenGeneratorIsMissing(t *testing.T) {
	service := &AppService{}

	_, err := service.GenerateReportTemplateDraft(context.Background(), domain.AuthenticatedUser{
		ID:               7,
		Role:             domain.RolePsychologist,
		SubscriptionPlan: domain.SubscriptionPlanPro,
	}, domain.GenerateReportTemplateDraftInput{
		Prompt: "Сделай шаблон для профориентационного теста",
	})
	if err != ErrReportTemplateDraftUnavailable {
		t.Fatalf("expected ErrReportTemplateDraftUnavailable, got %v", err)
	}
}

func TestGenerateReportTemplateDraftNormalizesGeneratorOutput(t *testing.T) {
	service := &AppService{
		reportTemplateDraftGenerator: fakeReportTemplateDraftGenerator{
			enabled: true,
			model:   "gpt-5-mini",
			draft: generatedReportTemplateDraft{
				Name:        "  Теплый шаблон для клиента  ",
				Description: "  Подходит для бережной выдачи результата.  ",
				Client: reportAudienceTemplateConfig{
					Title:           "  Отчет для клиента  ",
					IntroParagraphs: []string{"  В этом отчете собран предварительный результат теста.  "},
				},
				Psychologist: reportAudienceTemplateConfig{
					Title: "  Рабочий отчет психолога  ",
					SectionTitles: map[string]string{
						reportSectionAnswersTable: " Ответы по сессии ",
					},
				},
			},
		},
	}

	response, err := service.GenerateReportTemplateDraft(context.Background(), domain.AuthenticatedUser{
		ID:               7,
		Role:             domain.RolePsychologist,
		SubscriptionPlan: domain.SubscriptionPlanPro,
	}, domain.GenerateReportTemplateDraftInput{
		Prompt: "Сделай спокойный и профессиональный шаблон отчета",
	})
	if err != nil {
		t.Fatalf("GenerateReportTemplateDraft returned error: %v", err)
	}

	if response.Name != "Теплый шаблон для клиента" {
		t.Fatalf("expected normalized name, got %q", response.Name)
	}
	if response.Model != "gpt-5-mini" {
		t.Fatalf("expected model in response, got %q", response.Model)
	}
	if !strings.Contains(response.TemplateBody, `"title": "Отчет для клиента"`) {
		t.Fatalf("expected normalized template body, got %s", response.TemplateBody)
	}
	if !strings.Contains(response.TemplateBody, `"answers_table": "Ответы по сессии"`) {
		t.Fatalf("expected normalized section title, got %s", response.TemplateBody)
	}
}

func TestGenerateReportTemplateDraftFiltersUnsupportedTemplateKeys(t *testing.T) {
	service := &AppService{
		reportTemplateDraftGenerator: fakeReportTemplateDraftGenerator{
			enabled: true,
			model:   "gpt-5-mini",
			draft: generatedReportTemplateDraft{
				Name: "",
				Client: reportAudienceTemplateConfig{
					MetaLabels: map[string]string{
						"respondent": "Клиент",
						"email":      "Email",
					},
					SectionTitles: map[string]string{
						reportSectionSummary: "Краткий вывод",
						"email_block":        "Email блок",
					},
				},
				Psychologist: reportAudienceTemplateConfig{
					MetaLabels: map[string]string{
						reportMetaTopScales: "Топ шкалы",
						"phone":             "Телефон",
					},
					SectionTitles: map[string]string{
						reportSectionAnswersTable: "Ответы",
						"summary":                 "Нельзя для психолога",
					},
				},
			},
		},
	}

	response, err := service.GenerateReportTemplateDraft(context.Background(), domain.AuthenticatedUser{
		ID:               7,
		Role:             domain.RolePsychologist,
		SubscriptionPlan: domain.SubscriptionPlanPro,
	}, domain.GenerateReportTemplateDraftInput{
		Prompt: "Сделай спокойный и профессиональный шаблон отчета",
	})
	if err != nil {
		t.Fatalf("GenerateReportTemplateDraft returned error: %v", err)
	}

	if response.Name != "AI шаблон отчета" {
		t.Fatalf("expected fallback name, got %q", response.Name)
	}
	if strings.Contains(response.TemplateBody, `"email"`) || strings.Contains(response.TemplateBody, `"email_block"`) || strings.Contains(response.TemplateBody, `"phone"`) {
		t.Fatalf("expected unsupported keys to be removed, got %s", response.TemplateBody)
	}
	if strings.Contains(response.TemplateBody, `"summary": "Нельзя для психолога"`) {
		t.Fatalf("expected unsupported psychologist section to be removed, got %s", response.TemplateBody)
	}
	if !strings.Contains(response.TemplateBody, `"respondent": "Клиент"`) {
		t.Fatalf("expected supported client meta label to remain, got %s", response.TemplateBody)
	}
	if !strings.Contains(response.TemplateBody, `"answers_table": "Ответы"`) {
		t.Fatalf("expected supported psychologist section to remain, got %s", response.TemplateBody)
	}
}

func TestGenerateReportTemplateDraftBackfillsMissingTemplateContent(t *testing.T) {
	service := &AppService{
		reportTemplateDraftGenerator: fakeReportTemplateDraftGenerator{
			enabled: true,
			model:   "gpt-5-mini",
			draft: generatedReportTemplateDraft{
				Name:        "Черновик отчета",
				Description: "Минимальный AI-ответ",
				Client: reportAudienceTemplateConfig{
					Title: "",
				},
				Psychologist: reportAudienceTemplateConfig{
					Title: "",
				},
			},
		},
	}

	response, err := service.GenerateReportTemplateDraft(context.Background(), domain.AuthenticatedUser{
		ID:               7,
		Role:             domain.RolePsychologist,
		SubscriptionPlan: domain.SubscriptionPlanPro,
	}, domain.GenerateReportTemplateDraftInput{
		Prompt: "Сделай шаблон отчета",
	})
	if err != nil {
		t.Fatalf("GenerateReportTemplateDraft returned error: %v", err)
	}

	expectedFragments := []string{
		`"title": "Отчет по результатам теста"`,
		`"summary": "Краткий вывод"`,
		`"chart_data": "Профиль результата"`,
		`"scales_list": "Результаты"`,
		`"interpretation": "Интерпретация"`,
		`"recommendations": "Рекомендации"`,
		`"chart_caption": "Диаграмма показывает наиболее выраженные результаты текущего прохождения и помогает быстро увидеть их соотношение."`,
		`"title": "Рабочий отчет психолога"`,
		`"raw_scores": "Сырые показатели"`,
		`"answers_table": "Таблица ответов"`,
		`"intro_paragraphs": [`,
	}
	for _, fragment := range expectedFragments {
		if !strings.Contains(response.TemplateBody, fragment) {
			t.Fatalf("expected %q in template body, got %s", fragment, response.TemplateBody)
		}
	}
}

func TestGenerateReportTemplateDraftBackfillsMissingMetaLabelsAndIntroParagraphs(t *testing.T) {
	service := &AppService{
		reportTemplateDraftGenerator: fakeReportTemplateDraftGenerator{
			enabled: true,
			model:   "openrouter/free",
			draft: generatedReportTemplateDraft{
				Name:        "Черновик",
				Description: "",
				Client: reportAudienceTemplateConfig{
					MetaLabels: map[string]string{
						reportMetaRespondent: "Клиент",
					},
					IntroParagraphs: []string{"Только один вступительный абзац."},
				},
				Psychologist: reportAudienceTemplateConfig{
					MetaLabels: map[string]string{
						reportMetaStatus: "Статус",
					},
					IntroParagraphs: []string{"Один рабочий абзац."},
				},
			},
		},
	}

	response, err := service.GenerateReportTemplateDraft(context.Background(), domain.AuthenticatedUser{
		ID:               7,
		Role:             domain.RolePsychologist,
		SubscriptionPlan: domain.SubscriptionPlanPro,
	}, domain.GenerateReportTemplateDraftInput{
		Prompt: "Сделай подробный шаблон отчета",
	})
	if err != nil {
		t.Fatalf("GenerateReportTemplateDraft returned error: %v", err)
	}

	if !strings.Contains(response.Description, "Черновик шаблона отчета") {
		t.Fatalf("expected fallback description, got %q", response.Description)
	}
	expectedFragments := []string{
		`"respondent": "Клиент"`,
		`"session": "Сессия"`,
		`"status": "Статус"`,
		`"top_scales": "Топ-шкалы"`,
		`"top_professions": "Топ-профессии"`,
		`"Только один вступительный абзац."`,
		`"Этот отчет помогает спокойно посмотреть на результаты текущего прохождения и выделить наиболее заметные особенности профиля."`,
		`"Один рабочий абзац."`,
		`"Этот черновик предназначен для профессионального разбора результатов и дальнейшей рабочей интерпретации на консультации."`,
	}
	for _, fragment := range expectedFragments {
		if !strings.Contains(response.TemplateBody, fragment) {
			t.Fatalf("expected %q in template body, got %s", fragment, response.TemplateBody)
		}
	}
}

func TestBuildReportTemplateDraftUserPromptIncludesStrictSkeleton(t *testing.T) {
	prompt := buildReportTemplateDraftUserPrompt(reportTemplateDraftRequest{
		Prompt:      "Сделай теплый и понятный шаблон",
		TestContext: "Название теста: Профориентация",
	})

	expectedFragments := []string{
		"Промпт психолога:",
		"Контекст теста:",
		`"section_titles": {`,
		`"summary": "Краткий вывод"`,
		`"raw_scores": "Сырые показатели"`,
		`"top_professions": "Топ-профессии"`,
		"Не сокращай структуру.",
		"Верни только один JSON-объект.",
	}
	for _, fragment := range expectedFragments {
		if !strings.Contains(prompt, fragment) {
			t.Fatalf("expected %q in prompt, got %s", fragment, prompt)
		}
	}
}

func TestDecodeGeneratedReportTemplateDraftRepairsLooseJSON(t *testing.T) {
	draft, err := decodeGeneratedReportTemplateDraft(`{
  name: 'AI шаблон',
  description: 'Черновик для клиента и психолога',
  client: {
    title: 'Отчет для клиента',
    chart_caption: 'Пояснение к диаграмме',
    section_titles: {
      results: 'Результаты'
    },
  },
  psychologist: {
    title: 'Рабочий отчет психолога',
    section_titles: {
      raw_scores: 'Сырые показатели',
      answers: 'Таблица ответов',
    },
  },
}`)
	if err != nil {
		t.Fatalf("decodeGeneratedReportTemplateDraft returned error: %v", err)
	}
	if draft.Name != "AI шаблон" {
		t.Fatalf("expected repaired draft name, got %+v", draft)
	}
	if draft.Client.ChartCaption != "Пояснение к диаграмме" {
		t.Fatalf("expected chart caption after repair, got %+v", draft)
	}
	if draft.Client.SectionTitles[reportSectionScalesList] != "Результаты" {
		t.Fatalf("expected results section title after repair, got %+v", draft.Client.SectionTitles)
	}
	if draft.Psychologist.SectionTitles[reportSectionRawScores] != "Сырые показатели" {
		t.Fatalf("expected raw scores title after repair, got %+v", draft.Psychologist.SectionTitles)
	}
}

func TestDecodeGeneratedReportTemplateDraftBuildsFromPlainText(t *testing.T) {
	draft, err := decodeGeneratedReportTemplateDraft(`
Название: Бережный шаблон профориентационного отчета
Описание: Шаблон для спокойной выдачи результатов теста.

Клиент:
Заголовок: Отчет для клиента
Вступление: Этот отчет помогает увидеть сильные стороны и возможные направления развития.
Результаты: Результаты теста
Подпись к диаграмме: Диаграмма показывает наиболее выраженные показатели.

Психолог:
Заголовок: Рабочий отчет психолога
Вступление: Используйте этот шаблон как основу для профессиональной интерпретации результатов.
Raw scores: Сырые показатели
Таблица ответов: Ответы респондента
`)
	if err != nil {
		t.Fatalf("decodeGeneratedReportTemplateDraft returned error: %v", err)
	}
	if draft.Name != "Бережный шаблон профориентационного отчета" {
		t.Fatalf("expected name from prose, got %+v", draft)
	}
	if draft.Client.Title != "Отчет для клиента" {
		t.Fatalf("expected client title from prose, got %+v", draft.Client)
	}
	if draft.Client.SectionTitles[reportSectionScalesList] != "Результаты теста" {
		t.Fatalf("expected client results section from prose, got %+v", draft.Client.SectionTitles)
	}
	if draft.Client.ChartCaption != "Диаграмма показывает наиболее выраженные показатели." {
		t.Fatalf("expected client chart caption from prose, got %+v", draft.Client)
	}
	if draft.Psychologist.SectionTitles[reportSectionRawScores] != "Сырые показатели" {
		t.Fatalf("expected psychologist raw scores from prose, got %+v", draft.Psychologist.SectionTitles)
	}
	if draft.Psychologist.SectionTitles[reportSectionAnswersTable] != "Ответы респондента" {
		t.Fatalf("expected psychologist answers table from prose, got %+v", draft.Psychologist.SectionTitles)
	}
}

func TestOpenAIReportTemplateDraftGeneratorParsesResponse(t *testing.T) {
	generator := &openAIReportTemplateDraftGenerator{
		apiKey:  "test-key",
		baseURL: "https://api.openai.example/v1",
		model:   "gpt-5-mini",
		httpClient: &http.Client{
			Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				if r.URL.Path != "/v1/responses" {
					t.Fatalf("expected /v1/responses path, got %s", r.URL.Path)
				}
				if auth := r.Header.Get("Authorization"); auth != "Bearer test-key" {
					t.Fatalf("expected bearer auth, got %q", auth)
				}

				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatalf("failed to read request body: %v", err)
				}
				if !strings.Contains(string(body), `"json_schema"`) {
					t.Fatalf("expected structured output request, got %s", string(body))
				}
				if !strings.Contains(string(body), "бережный шаблон") {
					t.Fatalf("expected prompt in request, got %s", string(body))
				}

				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body: io.NopCloser(strings.NewReader(`{
					  "output": [
					    {
					      "type": "message",
					      "content": [
					        {
					          "type": "output_text",
					          "text": "{\"name\":\"AI шаблон\",\"description\":\"Черновик для мягкой выдачи результата\",\"client\":{\"title\":\"Отчет для клиента\"},\"psychologist\":{\"title\":\"Рабочий отчет психолога\"}}"
					        }
					      ]
					    }
					  ]
					}`)),
				}, nil
			}),
		},
	}

	draft, err := generator.GenerateReportTemplateDraft(context.Background(), reportTemplateDraftRequest{
		Prompt:      "Сделай бережный шаблон для клиента",
		TestContext: "Название теста: Профориентация",
	})
	if err != nil {
		t.Fatalf("GenerateReportTemplateDraft returned error: %v", err)
	}

	if draft.Name != "AI шаблон" {
		t.Fatalf("expected parsed name, got %+v", draft)
	}
	if draft.Client.Title != "Отчет для клиента" || draft.Psychologist.Title != "Рабочий отчет психолога" {
		t.Fatalf("unexpected parsed draft: %+v", draft)
	}
}

func TestOpenRouterReportTemplateDraftGeneratorUsesChatCompletions(t *testing.T) {
	generator := &openAIReportTemplateDraftGenerator{
		apiKey:  "test-key",
		baseURL: "https://openrouter.ai/api/v1",
		model:   "qwen/qwen3-next-80b-a3b-instruct:free",
		httpClient: &http.Client{
			Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				if r.URL.Path != "/api/v1/chat/completions" {
					t.Fatalf("expected /api/v1/chat/completions path, got %s", r.URL.Path)
				}
				if auth := r.Header.Get("Authorization"); auth != "Bearer test-key" {
					t.Fatalf("expected bearer auth, got %q", auth)
				}

				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatalf("failed to read request body: %v", err)
				}
				payload := string(body)
				if !strings.Contains(payload, `"response_format"`) || !strings.Contains(payload, `"json_schema"`) {
					t.Fatalf("expected chat completions structured output request, got %s", payload)
				}
				if !strings.Contains(payload, `"allow_fallbacks":true`) || !strings.Contains(payload, `"require_parameters":true`) {
					t.Fatalf("expected provider preferences in request, got %s", payload)
				}
				if !strings.Contains(payload, "мягкий шаблон") {
					t.Fatalf("expected prompt in request, got %s", payload)
				}

				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body: io.NopCloser(strings.NewReader(`{
					  "choices": [
					    {
					      "message": {
					        "content": "{\"name\":\"AI шаблон\",\"description\":\"Черновик для мягкой выдачи результата\",\"client\":{\"title\":\"Отчет для клиента\"},\"psychologist\":{\"title\":\"Рабочий отчет психолога\"}}"
					      }
					    }
					  ]
					}`)),
				}, nil
			}),
		},
	}

	draft, err := generator.GenerateReportTemplateDraft(context.Background(), reportTemplateDraftRequest{
		Prompt:      "Сделай мягкий шаблон для клиента",
		TestContext: "Название теста: Профориентация",
	})
	if err != nil {
		t.Fatalf("GenerateReportTemplateDraft returned error: %v", err)
	}

	if draft.Name != "AI шаблон" {
		t.Fatalf("expected parsed name, got %+v", draft)
	}
	if draft.Client.Title != "Отчет для клиента" || draft.Psychologist.Title != "Рабочий отчет психолога" {
		t.Fatalf("unexpected parsed draft: %+v", draft)
	}
}

func TestOpenRouterReportTemplateDraftGeneratorFallsBackWithoutStructuredOutput(t *testing.T) {
	requestCount := 0

	generator := &openAIReportTemplateDraftGenerator{
		apiKey:  "test-key",
		baseURL: "https://openrouter.ai/api/v1",
		model:   "qwen/qwen3-next-80b-a3b-instruct:free",
		httpClient: &http.Client{
			Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				requestCount++

				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatalf("failed to read request body: %v", err)
				}
				payload := string(body)

				if requestCount == 1 {
					if !strings.Contains(payload, `"response_format"`) {
						t.Fatalf("expected structured output on first request, got %s", payload)
					}
					return &http.Response{
						StatusCode: http.StatusBadRequest,
						Header:     make(http.Header),
						Body: io.NopCloser(strings.NewReader(`{
						  "error": {
						    "message": "response_format json_schema is not supported for this model"
						  }
						}`)),
					}, nil
				}

				if strings.Contains(payload, `"response_format"`) {
					t.Fatalf("did not expect response_format on fallback request, got %s", payload)
				}
				if !strings.Contains(payload, `"allow_fallbacks":true`) {
					t.Fatalf("expected provider fallback preference on retry, got %s", payload)
				}

				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body: io.NopCloser(strings.NewReader("{\n" +
						"  \"choices\": [\n" +
						"    {\n" +
						"      \"message\": {\n" +
						"        \"content\": \"```json\\n{\\\"name\\\":\\\"AI шаблон\\\",\\\"description\\\":\\\"Черновик для мягкой выдачи результата\\\",\\\"client\\\":{\\\"title\\\":\\\"Отчет для клиента\\\"},\\\"psychologist\\\":{\\\"title\\\":\\\"Рабочий отчет психолога\\\"}}\\n```\"\n" +
						"      }\n" +
						"    }\n" +
						"  ]\n" +
						"}")),
				}, nil
			}),
		},
	}

	draft, err := generator.GenerateReportTemplateDraft(context.Background(), reportTemplateDraftRequest{
		Prompt:      "Сделай мягкий шаблон для клиента",
		TestContext: "Название теста: Профориентация",
	})
	if err != nil {
		t.Fatalf("GenerateReportTemplateDraft returned error: %v", err)
	}
	if requestCount != 2 {
		t.Fatalf("expected two requests with fallback, got %d", requestCount)
	}
	if draft.Name != "AI шаблон" {
		t.Fatalf("expected parsed name after fallback, got %+v", draft)
	}
}

func TestOpenRouterReportTemplateDraftGeneratorFallsBackToFreeRouterOnProviderError(t *testing.T) {
	requestCount := 0

	generator := &openAIReportTemplateDraftGenerator{
		apiKey:  "test-key",
		baseURL: "https://openrouter.ai/api/v1",
		model:   "qwen/qwen3-next-80b-a3b-instruct:free",
		httpClient: &http.Client{
			Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				requestCount++

				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatalf("failed to read request body: %v", err)
				}
				payload := string(body)

				if requestCount == 1 {
					if !strings.Contains(payload, `"model":"qwen/qwen3-next-80b-a3b-instruct:free"`) {
						t.Fatalf("expected original model on first request, got %s", payload)
					}
					return &http.Response{
						StatusCode: http.StatusBadGateway,
						Header:     make(http.Header),
						Body: io.NopCloser(strings.NewReader(`{
						  "error": {
						    "message": "No endpoints found for this model"
						  }
						}`)),
					}, nil
				}

				if !strings.Contains(payload, `"model":"openrouter/free"`) {
					t.Fatalf("expected router fallback model on second request, got %s", payload)
				}
				if strings.Contains(payload, `"response_format"`) {
					t.Fatalf("did not expect response_format on router fallback request, got %s", payload)
				}

				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body: io.NopCloser(strings.NewReader(`{
					  "choices": [
					    {
					      "message": {
					        "content": "{\"name\":\"Router шаблон\",\"description\":\"Черновик через free router\",\"client\":{\"title\":\"Отчет для клиента\"},\"psychologist\":{\"title\":\"Рабочий отчет психолога\"}}"
					      }
					    }
					  ]
					}`)),
				}, nil
			}),
		},
	}

	draft, err := generator.GenerateReportTemplateDraft(context.Background(), reportTemplateDraftRequest{
		Prompt:      "Сделай рабочий шаблон для клиента",
		TestContext: "Название теста: Профориентация",
	})
	if err != nil {
		t.Fatalf("GenerateReportTemplateDraft returned error: %v", err)
	}
	if requestCount != 2 {
		t.Fatalf("expected router fallback with two requests, got %d", requestCount)
	}
	if draft.Name != "Router шаблон" {
		t.Fatalf("expected parsed fallback name, got %+v", draft)
	}
}

func TestOpenRouterReportTemplateDraftGeneratorRetriesTransientProviderFailure(t *testing.T) {
	requestCount := 0

	generator := &openAIReportTemplateDraftGenerator{
		apiKey:  "test-key",
		baseURL: "https://openrouter.ai/api/v1",
		model:   "openrouter/free",
		httpClient: &http.Client{
			Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				requestCount++
				if requestCount == 1 {
					return &http.Response{
						StatusCode: http.StatusBadGateway,
						Header:     make(http.Header),
						Body: io.NopCloser(strings.NewReader(`{
						  "error": {
						    "message": "Venice: mistralai/mistral-small-3.1-24b-instruct:free is temporarily rate-limited upstream. Please retry shortly."
						  }
						}`)),
					}, nil
				}

				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body: io.NopCloser(strings.NewReader(`{
					  "choices": [
					    {
					      "message": {
					        "content": "{\"name\":\"Успешный шаблон\",\"description\":\"Шаблон после retry\",\"client\":{\"title\":\"Отчет для клиента\"},\"psychologist\":{\"title\":\"Рабочий отчет психолога\"}}"
					      }
					    }
					  ]
					}`)),
				}, nil
			}),
		},
	}

	draft, err := generator.GenerateReportTemplateDraft(context.Background(), reportTemplateDraftRequest{
		Prompt: "Сделай шаблон отчета",
	})
	if err != nil {
		t.Fatalf("GenerateReportTemplateDraft returned error: %v", err)
	}
	if requestCount != 2 {
		t.Fatalf("expected retry on transient provider failure, got %d requests", requestCount)
	}
	if draft.Name != "Успешный шаблон" {
		t.Fatalf("expected successful draft after retry, got %+v", draft)
	}
}

func TestOpenRouterReportTemplateDraftGeneratorWrapsTransportErrors(t *testing.T) {
	generator := &openAIReportTemplateDraftGenerator{
		apiKey:  "test-key",
		baseURL: "https://openrouter.ai/api/v1",
		model:   "openrouter/free",
		httpClient: &http.Client{
			Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				return nil, errors.New("upstream connection reset")
			}),
		},
	}

	_, err := generator.GenerateReportTemplateDraft(context.Background(), reportTemplateDraftRequest{
		Prompt: "Сделай шаблон отчета",
	})
	if err == nil {
		t.Fatal("expected provider error, got nil")
	}
	if !errors.Is(err, ErrReportTemplateDraftFailed) {
		t.Fatalf("expected ErrReportTemplateDraftFailed, got %v", err)
	}
	if !strings.Contains(err.Error(), "upstream connection reset") {
		t.Fatalf("expected wrapped transport message, got %v", err)
	}
}
