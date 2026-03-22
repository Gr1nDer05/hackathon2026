import { LoaderCircle, Plus, Save, Trash2 } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import useSWR from "swr";
import {
  createPsychologistReportTemplateRequest,
  deletePsychologistReportTemplateRequest,
  generatePsychologistReportTemplateDraftRequest,
  listPsychologistReportTemplatesRequest,
  listPsychologistTestsRequest,
  updatePsychologistReportTemplateRequest,
} from "../../modules/tests/api/testsApi";
import { useAuth } from "../../modules/auth/model/useAuth";
import {
  getSubscriptionPlan,
  getSubscriptionPlanLabel,
  hasAiTemplateAccess,
} from "../../modules/psychologist/lib/psychologistUi";
import { ROUTES } from "../../shared/config/routes";
import PageCard from "../../shared/ui/PageCard";

function createReportTemplateDraft(overrides = {}) {
  return {
    id: overrides.id ?? null,
    name: String(overrides.name || ""),
    description: String(overrides.description || ""),
    client_title: String(overrides.client_title || ""),
    psychologist_title: String(overrides.psychologist_title || ""),
    client_summary_title: String(overrides.client_summary_title || ""),
    client_chart_caption: String(overrides.client_chart_caption || ""),
    psychologist_raw_scores_title: String(overrides.psychologist_raw_scores_title || ""),
    psychologist_answers_title: String(overrides.psychologist_answers_title || ""),
    client_intro: String(overrides.client_intro || ""),
    psychologist_intro: String(overrides.psychologist_intro || ""),
    client_closing: String(overrides.client_closing || ""),
    psychologist_closing: String(overrides.psychologist_closing || ""),
    source_body: overrides.source_body || {},
  };
}

const EMPTY_AI_FORM = {
  prompt: "",
  test_id: "",
};

function AutoResizeTextarea({
  className = "",
  minRows = 1,
  onChange,
  onKeyDown,
  singleLine = false,
  value,
  ...props
}) {
  const textareaRef = useRef(null);

  function resize(element = textareaRef.current) {
    if (!element) {
      return;
    }

    const styles = window.getComputedStyle(element);
    const lineHeight = parseFloat(styles.lineHeight) || 22;
    const paddingTop = parseFloat(styles.paddingTop) || 0;
    const paddingBottom = parseFloat(styles.paddingBottom) || 0;
    const borderTop = parseFloat(styles.borderTopWidth) || 0;
    const borderBottom = parseFloat(styles.borderBottomWidth) || 0;
    const minHeight =
      lineHeight * minRows + paddingTop + paddingBottom + borderTop + borderBottom;

    element.style.height = "0px";
    element.style.height = `${Math.max(element.scrollHeight, minHeight)}px`;
  }

  useEffect(() => {
    resize();
  }, [value]);

  return (
    <textarea
      {...props}
      ref={textareaRef}
      className={`admin-form-control report-template-field ${className}`.trim()}
      rows={minRows}
      value={value}
      onChange={(event) => {
        if (singleLine) {
          event.target.value = event.target.value.replace(/\s*\n+\s*/g, " ");
        }

        onChange?.(event);
        resize(event.target);
      }}
      onKeyDown={(event) => {
        if (singleLine && event.key === "Enter") {
          event.preventDefault();
          return;
        }

        onKeyDown?.(event);
      }}
    />
  );
}

function safeParseTemplateBody(rawValue) {
  try {
    return JSON.parse(String(rawValue || "{}"));
  } catch {
    return {};
  }
}

function toParagraphs(value) {
  return String(value || "")
    .split(/\n{2,}/)
    .map((item) => item.trim())
    .filter(Boolean);
}

function normalizeReportTemplate(template) {
  const sourceBody = safeParseTemplateBody(template?.template_body);
  const clientSection = sourceBody?.client || {};
  const psychologistSection = sourceBody?.psychologist || {};
  const clientSectionTitles = clientSection?.section_titles || {};
  const psychologistSectionTitles = psychologistSection?.section_titles || {};

  return createReportTemplateDraft({
    id: template?.id,
    name: template?.name,
    description: template?.description,
    client_title: clientSection?.title,
    psychologist_title: psychologistSection?.title,
    client_summary_title: clientSectionTitles?.summary,
    client_chart_caption: clientSection?.chart_caption,
    psychologist_raw_scores_title: psychologistSectionTitles?.raw_scores,
    psychologist_answers_title: psychologistSectionTitles?.answers_table,
    client_intro: Array.isArray(clientSection?.intro_paragraphs) ? clientSection.intro_paragraphs.join("\n\n") : "",
    psychologist_intro: Array.isArray(psychologistSection?.intro_paragraphs)
      ? psychologistSection.intro_paragraphs.join("\n\n")
      : "",
    client_closing: Array.isArray(clientSection?.closing_paragraphs) ? clientSection.closing_paragraphs.join("\n\n") : "",
    psychologist_closing: Array.isArray(psychologistSection?.closing_paragraphs)
      ? psychologistSection.closing_paragraphs.join("\n\n")
      : "",
    source_body: sourceBody,
  });
}

function buildReportTemplatePayload(templateForm) {
  const sourceBody = templateForm?.source_body && typeof templateForm.source_body === "object" ? templateForm.source_body : {};
  const nextBody = {
    ...sourceBody,
    client: {
      ...(sourceBody.client || {}),
    },
    psychologist: {
      ...(sourceBody.psychologist || {}),
    },
  };

  nextBody.client.title = String(templateForm.client_title || "").trim();
  nextBody.psychologist.title = String(templateForm.psychologist_title || "").trim();
  nextBody.client.intro_paragraphs = toParagraphs(templateForm.client_intro);
  nextBody.psychologist.intro_paragraphs = toParagraphs(templateForm.psychologist_intro);
  nextBody.client.closing_paragraphs = toParagraphs(templateForm.client_closing);
  nextBody.psychologist.closing_paragraphs = toParagraphs(templateForm.psychologist_closing);
  nextBody.client.chart_caption = String(templateForm.client_chart_caption || "").trim();
  nextBody.client.section_titles = {
    ...(nextBody.client.section_titles || {}),
    summary: String(templateForm.client_summary_title || "").trim(),
  };
  nextBody.psychologist.section_titles = {
    ...(nextBody.psychologist.section_titles || {}),
    raw_scores: String(templateForm.psychologist_raw_scores_title || "").trim(),
    answers_table: String(templateForm.psychologist_answers_title || "").trim(),
  };

  return {
    name: String(templateForm.name || "").trim(),
    description: String(templateForm.description || "").trim(),
    template_body: JSON.stringify(nextBody, null, 2),
  };
}

function validateReportTemplateForm(templateForm) {
  const errors = {};

  if (!String(templateForm.name || "").trim()) {
    errors.name = "Укажите название шаблона.";
  }

  return errors;
}

export default function ReportTemplatesPage() {
  const { user } = useAuth();
  const templatesQuery = useSWR("psychologist-report-templates", listPsychologistReportTemplatesRequest, {
    revalidateOnFocus: false,
    shouldRetryOnError: false,
  });
  const testsQuery = useSWR("psychologist-tests", listPsychologistTestsRequest, {
    revalidateOnFocus: false,
    shouldRetryOnError: false,
  });

  const [draftTemplates, setDraftTemplates] = useState([]);
  const [newTemplateForm, setNewTemplateForm] = useState(createReportTemplateDraft());
  const [newTemplateErrors, setNewTemplateErrors] = useState({});
  const [templateErrorsById, setTemplateErrorsById] = useState({});
  const [aiForm, setAiForm] = useState(EMPTY_AI_FORM);
  const [feedbackMessage, setFeedbackMessage] = useState("");
  const [feedbackError, setFeedbackError] = useState("");
  const [isCreatingTemplate, setIsCreatingTemplate] = useState(false);
  const [isGeneratingDraft, setIsGeneratingDraft] = useState(false);
  const [activeTemplateId, setActiveTemplateId] = useState(null);

  const subscriptionPlan = getSubscriptionPlan(user);
  const canGenerateAiDraft = hasAiTemplateAccess(user);
  const tests = Array.isArray(testsQuery.data) ? testsQuery.data : [];

  useEffect(() => {
    if (!templatesQuery.data) {
      return;
    }

    setDraftTemplates(templatesQuery.data.map(normalizeReportTemplate));
  }, [templatesQuery.data]);

  useEffect(() => {
    if (!feedbackMessage) {
      return undefined;
    }

    const timeoutId = window.setTimeout(() => setFeedbackMessage(""), 2600);
    return () => window.clearTimeout(timeoutId);
  }, [feedbackMessage]);

  function clearNewTemplateError(field) {
    setNewTemplateErrors((prev) => {
      if (!prev[field]) {
        return prev;
      }

      const next = { ...prev };
      delete next[field];
      return next;
    });
  }

  function clearTemplateError(templateId, field) {
    setTemplateErrorsById((prev) => {
      const current = prev[templateId];
      if (!current || !current[field]) {
        return prev;
      }

      const next = { ...prev };
      next[templateId] = { ...current };
      delete next[templateId][field];

      if (!Object.keys(next[templateId]).length) {
        delete next[templateId];
      }

      return next;
    });
  }

  function handleNewTemplateFieldChange(field, value) {
    setNewTemplateForm((prev) => ({
      ...prev,
      [field]: value,
    }));
    clearNewTemplateError(field);
    if (feedbackError) {
      setFeedbackError("");
    }
  }

  function handleTemplateFieldChange(templateId, field, value) {
    setDraftTemplates((prev) =>
      prev.map((template) =>
        String(template.id) === String(templateId)
          ? {
              ...template,
              [field]: value,
            }
          : template,
      ),
    );
    clearTemplateError(templateId, field);
    if (feedbackError) {
      setFeedbackError("");
    }
  }

  function handleAiFieldChange(field, value) {
    setAiForm((prev) => ({
      ...prev,
      [field]: value,
    }));
    if (feedbackError) {
      setFeedbackError("");
    }
  }

  async function handleCreateTemplate(event) {
    event.preventDefault();
    const errors = validateReportTemplateForm(newTemplateForm);
    setNewTemplateErrors(errors);

    if (Object.keys(errors).length) {
      return;
    }

    setIsCreatingTemplate(true);
    setFeedbackError("");

    try {
      await createPsychologistReportTemplateRequest(buildReportTemplatePayload(newTemplateForm));
      setNewTemplateForm(createReportTemplateDraft());
      setNewTemplateErrors({});
      await templatesQuery.mutate();
      setFeedbackMessage("Шаблон отчёта создан.");
    } catch (error) {
      setFeedbackError(error?.message || "Не удалось создать шаблон отчёта.");
    } finally {
      setIsCreatingTemplate(false);
    }
  }

  async function handleGenerateTemplateDraft(event) {
    event.preventDefault();

    if (!canGenerateAiDraft) {
      setFeedbackError("Автозаполнение шаблона сейчас недоступно на вашем плане.");
      return;
    }

    const prompt = String(aiForm.prompt || "").trim();
    if (!prompt) {
      setFeedbackError("Опишите, какой шаблон отчёта нужно сгенерировать.");
      return;
    }

    setIsGeneratingDraft(true);
    setFeedbackError("");

    try {
      const payload = {
        prompt,
      };
      const testId = Number(aiForm.test_id);
      if (Number.isInteger(testId) && testId > 0) {
        payload.test_id = testId;
      }

      const response = await generatePsychologistReportTemplateDraftRequest(payload);
      setNewTemplateForm(normalizeReportTemplate(response));
      setAiForm(EMPTY_AI_FORM);
      setFeedbackMessage("Шаблон заполнен автоматически. Проверьте текст и сохраните его.");
    } catch (error) {
      setFeedbackError(
        error?.message || "Проблемы с OpenRouter, попробуйте ещё раз.",
      );
    } finally {
      setIsGeneratingDraft(false);
    }
  }

  async function handleSaveTemplate(templateId) {
    const template = draftTemplates.find((item) => String(item.id) === String(templateId));
    if (!template) {
      return;
    }

    const errors = validateReportTemplateForm(template);
    if (Object.keys(errors).length) {
      setTemplateErrorsById((prev) => ({ ...prev, [templateId]: errors }));
      return;
    }

    setActiveTemplateId(templateId);
    setFeedbackError("");

    try {
      await updatePsychologistReportTemplateRequest(templateId, buildReportTemplatePayload(template));
      await templatesQuery.mutate();
      setFeedbackMessage("Шаблон отчёта сохранён.");
    } catch (error) {
      setFeedbackError(error?.message || "Не удалось сохранить шаблон отчёта.");
    } finally {
      setActiveTemplateId(null);
    }
  }

  async function handleDeleteTemplate(templateId) {
    setActiveTemplateId(templateId);
    setFeedbackError("");

    try {
      await deletePsychologistReportTemplateRequest(templateId);
      await templatesQuery.mutate();
      setFeedbackMessage("Шаблон отчёта удалён.");
    } catch (error) {
      setFeedbackError(error?.message || "Не удалось удалить шаблон отчёта.");
    } finally {
      setActiveTemplateId(null);
    }
  }

  return (
    <PageCard
      wide
      title="Шаблоны отчётов"
      description="Здесь можно настроить шаблоны отчётов для клиента и специалиста."
      links={[
        { to: ROUTES.tests, label: "К опросникам" },
        { to: ROUTES.dashboard, label: "В кабинет" },
      ]}
    >
      {templatesQuery.error ? (
        <p className="admin-form-message admin-form-message--error">
          {templatesQuery.error.message || "Не удалось загрузить шаблоны отчётов."}
        </p>
      ) : null}
      {testsQuery.error ? (
        <p className="admin-form-message admin-form-message--error">
          {testsQuery.error.message || "Не удалось загрузить список тестов."}
        </p>
      ) : null}
      {feedbackError ? <p className="admin-form-message admin-form-message--error">{feedbackError}</p> : null}
      {feedbackMessage ? <p className="admin-form-message">{feedbackMessage}</p> : null}

      <div className={`workflow-note ${canGenerateAiDraft ? "workflow-note--success" : "workflow-note--warning"}`}>
        <p>
          Текущий план: <strong>{getSubscriptionPlanLabel(subscriptionPlan)}</strong>.{" "}
          {canGenerateAiDraft
            ? "Можно быстро заполнить шаблон по описанию и потом поправить текст вручную."
            : "Автозаполнение шаблона сейчас недоступно на вашем плане."}
        </p>
      </div>

      <section className="builder-panel">
        <div className="builder-section-head">
          <div>
            <p className="builder-section-head__eyebrow">Быстрое заполнение</p>
            <h3 className="builder-section-head__title">Заполнить шаблон по описанию</h3>
          </div>
        </div>
        <p className="builder-panel__description">
          Опишите, каким должен быть отчёт. Сервис предложит основу, которую можно сразу поправить под себя.
        </p>

        <form className="admin-form-grid" onSubmit={handleGenerateTemplateDraft}>
          <label className="admin-form-field admin-form-field--wide">
            <span>Описание шаблона</span>
            <AutoResizeTextarea
              className="report-template-field--long"
              disabled={!canGenerateAiDraft || isGeneratingDraft}
              minRows={4}
              value={aiForm.prompt}
              onChange={(event) => handleAiFieldChange("prompt", event.target.value)}
              placeholder="Например: нужен спокойный клиентский отчёт с коротким выводом, диаграммой и пояснением по сильным сторонам."
            />
          </label>

          <label className="admin-form-field">
            <span>Тест для контекста</span>
            <select
              className="admin-form-control"
              disabled={!canGenerateAiDraft || isGeneratingDraft}
              value={aiForm.test_id}
              onChange={(event) => handleAiFieldChange("test_id", event.target.value)}
            >
              <option value="">Без теста</option>
              {tests.map((test) => (
                <option key={test.id} value={test.id}>
                  {test.title}
                </option>
              ))}
            </select>
          </label>

          <div className="admin-form-actions">
            <button
              className="admin-primary-button"
              disabled={!canGenerateAiDraft || isGeneratingDraft}
              type="submit"
            >
              {isGeneratingDraft ? (
                <LoaderCircle className="icon-spin" size={16} strokeWidth={2.1} />
              ) : (
                <Plus size={16} strokeWidth={2.1} />
              )}
              <span>{isGeneratingDraft ? "Подготовка..." : "Заполнить шаблон"}</span>
            </button>
          </div>
        </form>
      </section>

      <section className="builder-panel">
        <div className="builder-section-head">
          <div>
            <p className="builder-section-head__eyebrow">Новый шаблон</p>
            <h3 className="builder-section-head__title">Новый шаблон отчёта</h3>
          </div>
        </div>
        <p className="builder-panel__description">
          Один и тот же шаблон можно использовать в нескольких тестах.
        </p>

        <form className="admin-form-grid" onSubmit={handleCreateTemplate}>
          <label className="admin-form-field">
            <span>Название шаблона</span>
            <AutoResizeTextarea
              aria-invalid={Boolean(newTemplateErrors.name)}
              className={`${newTemplateErrors.name ? "admin-form-control--invalid " : ""}report-template-field--compact`}
              minRows={1}
              singleLine
              value={newTemplateForm.name}
              onChange={(event) => handleNewTemplateFieldChange("name", event.target.value)}
              placeholder="Например: Базовый шаблон отчёта"
            />
            {newTemplateErrors.name ? <small className="admin-form-error admin-form-error--inline">{newTemplateErrors.name}</small> : null}
          </label>

          <label className="admin-form-field">
            <span>Описание</span>
            <AutoResizeTextarea
              className="report-template-field--compact"
              minRows={1}
              singleLine
              value={newTemplateForm.description}
              onChange={(event) => handleNewTemplateFieldChange("description", event.target.value)}
              placeholder="Для каких методик подходит"
            />
          </label>

          <label className="admin-form-field">
            <span>Заголовок клиентского отчёта</span>
            <AutoResizeTextarea
              className="report-template-field--compact"
              minRows={1}
              singleLine
              value={newTemplateForm.client_title}
              onChange={(event) => handleNewTemplateFieldChange("client_title", event.target.value)}
            />
          </label>

          <label className="admin-form-field">
            <span>Заголовок отчёта психолога</span>
            <AutoResizeTextarea
              className="report-template-field--compact"
              minRows={1}
              singleLine
              value={newTemplateForm.psychologist_title}
              onChange={(event) => handleNewTemplateFieldChange("psychologist_title", event.target.value)}
            />
          </label>

          <label className="admin-form-field">
            <span>Заголовок блока результатов</span>
            <AutoResizeTextarea
              className="report-template-field--compact"
              minRows={1}
              singleLine
              value={newTemplateForm.client_summary_title}
              onChange={(event) => handleNewTemplateFieldChange("client_summary_title", event.target.value)}
            />
          </label>

          <label className="admin-form-field">
            <span>Подпись к диаграмме</span>
            <AutoResizeTextarea
              className="report-template-field--compact"
              minRows={1}
              singleLine
              value={newTemplateForm.client_chart_caption}
              onChange={(event) => handleNewTemplateFieldChange("client_chart_caption", event.target.value)}
            />
          </label>

          <label className="admin-form-field">
            <span>Заголовок блока с баллами</span>
            <AutoResizeTextarea
              className="report-template-field--compact"
              minRows={1}
              singleLine
              value={newTemplateForm.psychologist_raw_scores_title}
              onChange={(event) => handleNewTemplateFieldChange("psychologist_raw_scores_title", event.target.value)}
            />
          </label>

          <label className="admin-form-field">
            <span>Заголовок таблицы ответов</span>
            <AutoResizeTextarea
              className="report-template-field--compact"
              minRows={1}
              singleLine
              value={newTemplateForm.psychologist_answers_title}
              onChange={(event) => handleNewTemplateFieldChange("psychologist_answers_title", event.target.value)}
            />
          </label>

          <label className="admin-form-field admin-form-field--wide">
            <span>Вступление для клиента</span>
            <AutoResizeTextarea
              className="report-template-field--long"
              minRows={3}
              value={newTemplateForm.client_intro}
              onChange={(event) => handleNewTemplateFieldChange("client_intro", event.target.value)}
              placeholder="Новый абзац отделяйте пустой строкой."
            />
          </label>

          <label className="admin-form-field admin-form-field--wide">
            <span>Вступление для психолога</span>
            <AutoResizeTextarea
              className="report-template-field--long"
              minRows={3}
              value={newTemplateForm.psychologist_intro}
              onChange={(event) => handleNewTemplateFieldChange("psychologist_intro", event.target.value)}
            />
          </label>

          <label className="admin-form-field admin-form-field--wide">
            <span>Заключение для клиента</span>
            <AutoResizeTextarea
              className="report-template-field--long"
              minRows={3}
              value={newTemplateForm.client_closing}
              onChange={(event) => handleNewTemplateFieldChange("client_closing", event.target.value)}
            />
          </label>

          <label className="admin-form-field admin-form-field--wide">
            <span>Заключение для психолога</span>
            <AutoResizeTextarea
              className="report-template-field--long"
              minRows={3}
              value={newTemplateForm.psychologist_closing}
              onChange={(event) => handleNewTemplateFieldChange("psychologist_closing", event.target.value)}
            />
          </label>

          <div className="admin-form-actions">
            <button className="admin-primary-button" disabled={isCreatingTemplate} type="submit">
              {isCreatingTemplate ? (
                <LoaderCircle className="icon-spin" size={16} strokeWidth={2.1} />
              ) : (
                <Plus size={16} strokeWidth={2.1} />
              )}
              <span>{isCreatingTemplate ? "Создание..." : "Создать шаблон"}</span>
            </button>
          </div>
        </form>
      </section>

      <section className="builder-panel">
        <div className="builder-section-head">
          <div>
            <p className="builder-section-head__eyebrow">Список шаблонов</p>
            <h3 className="builder-section-head__title">Сохранённые шаблоны отчётов</h3>
          </div>
        </div>
        <div className="builder-formula-list">
          {templatesQuery.isLoading && !draftTemplates.length ? (
            <div className="builder-empty">Загружаем шаблоны отчётов...</div>
          ) : draftTemplates.length ? (
            draftTemplates.map((template) => {
              const templateErrors = templateErrorsById[template.id] || {};
              const isBusy = activeTemplateId === template.id;

              return (
                <article className="builder-formula-card" key={`report-template-page-${template.id}`}>
                  <div className="builder-formula-card__head">
                    <strong>{template.name || "Шаблон без названия"}</strong>
                    <div className="admin-table__actions">
                      <button className="table-action-button" disabled={isBusy} type="button" onClick={() => handleSaveTemplate(template.id)}>
                        {isBusy ? (
                          <LoaderCircle className="icon-spin" size={15} strokeWidth={2.1} />
                        ) : (
                          <Save size={15} strokeWidth={2.1} />
                        )}
                        <span>Сохранить</span>
                      </button>
                      <button className="table-action-button" disabled={isBusy} type="button" onClick={() => handleDeleteTemplate(template.id)}>
                        <Trash2 size={15} strokeWidth={2.1} />
                        <span>Удалить</span>
                      </button>
                    </div>
                  </div>

                  <div className="builder-formula-grid">
                    <label className="admin-form-field">
                      <span>Название</span>
                      <AutoResizeTextarea
                        aria-invalid={Boolean(templateErrors.name)}
                        className={`${templateErrors.name ? "admin-form-control--invalid " : ""}report-template-field--compact`}
                        minRows={1}
                        singleLine
                        value={template.name}
                        onChange={(event) => handleTemplateFieldChange(template.id, "name", event.target.value)}
                      />
                      {templateErrors.name ? <small className="admin-form-error admin-form-error--inline">{templateErrors.name}</small> : null}
                    </label>

                    <label className="admin-form-field">
                      <span>Описание</span>
                      <AutoResizeTextarea
                        className="report-template-field--compact"
                        minRows={1}
                        singleLine
                        value={template.description}
                        onChange={(event) => handleTemplateFieldChange(template.id, "description", event.target.value)}
                      />
                    </label>

                    <label className="admin-form-field">
                      <span>Клиентский заголовок</span>
                      <AutoResizeTextarea
                        className="report-template-field--compact"
                        minRows={1}
                        singleLine
                        value={template.client_title}
                        onChange={(event) => handleTemplateFieldChange(template.id, "client_title", event.target.value)}
                      />
                    </label>

                    <label className="admin-form-field">
                      <span>Заголовок психолога</span>
                      <AutoResizeTextarea
                        className="report-template-field--compact"
                        minRows={1}
                        singleLine
                        value={template.psychologist_title}
                        onChange={(event) => handleTemplateFieldChange(template.id, "psychologist_title", event.target.value)}
                      />
                    </label>

                    <label className="admin-form-field">
                      <span>Заголовок блока результатов</span>
                      <AutoResizeTextarea
                        className="report-template-field--compact"
                        minRows={1}
                        singleLine
                        value={template.client_summary_title}
                        onChange={(event) => handleTemplateFieldChange(template.id, "client_summary_title", event.target.value)}
                      />
                    </label>

                    <label className="admin-form-field">
                      <span>Подпись к диаграмме</span>
                      <AutoResizeTextarea
                        className="report-template-field--compact"
                        minRows={1}
                        singleLine
                        value={template.client_chart_caption}
                        onChange={(event) => handleTemplateFieldChange(template.id, "client_chart_caption", event.target.value)}
                      />
                    </label>

                    <label className="admin-form-field">
                      <span>Заголовок блока с баллами</span>
                      <AutoResizeTextarea
                        className="report-template-field--compact"
                        minRows={1}
                        singleLine
                        value={template.psychologist_raw_scores_title}
                        onChange={(event) => handleTemplateFieldChange(template.id, "psychologist_raw_scores_title", event.target.value)}
                      />
                    </label>

                    <label className="admin-form-field">
                      <span>Заголовок таблицы ответов</span>
                      <AutoResizeTextarea
                        className="report-template-field--compact"
                        minRows={1}
                        singleLine
                        value={template.psychologist_answers_title}
                        onChange={(event) => handleTemplateFieldChange(template.id, "psychologist_answers_title", event.target.value)}
                      />
                    </label>

                    <label className="admin-form-field admin-form-field--wide">
                      <span>Вступление для клиента</span>
                      <AutoResizeTextarea
                        className="report-template-field--long"
                        minRows={3}
                        value={template.client_intro}
                        onChange={(event) => handleTemplateFieldChange(template.id, "client_intro", event.target.value)}
                      />
                    </label>

                    <label className="admin-form-field admin-form-field--wide">
                      <span>Вступление для психолога</span>
                      <AutoResizeTextarea
                        className="report-template-field--long"
                        minRows={3}
                        value={template.psychologist_intro}
                        onChange={(event) => handleTemplateFieldChange(template.id, "psychologist_intro", event.target.value)}
                      />
                    </label>

                    <label className="admin-form-field admin-form-field--wide">
                      <span>Заключение для клиента</span>
                      <AutoResizeTextarea
                        className="report-template-field--long"
                        minRows={3}
                        value={template.client_closing}
                        onChange={(event) => handleTemplateFieldChange(template.id, "client_closing", event.target.value)}
                      />
                    </label>

                    <label className="admin-form-field admin-form-field--wide">
                      <span>Заключение для психолога</span>
                      <AutoResizeTextarea
                        className="report-template-field--long"
                        minRows={3}
                        value={template.psychologist_closing}
                        onChange={(event) => handleTemplateFieldChange(template.id, "psychologist_closing", event.target.value)}
                      />
                    </label>
                  </div>
                </article>
              );
            })
          ) : (
            <div className="builder-empty">Шаблонов отчётов пока нет. Создайте первый шаблон.</div>
          )}
        </div>
      </section>
    </PageCard>
  );
}
