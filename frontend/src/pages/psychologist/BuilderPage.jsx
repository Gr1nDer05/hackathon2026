import {
  Calculator,
  Copy,
  Eye,
  GripVertical,
  LayoutTemplate,
  Link2,
  LoaderCircle,
  Plus,
  Rocket,
  Save,
  Trash2,
} from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import useSWR from "swr";
import { Link, useParams } from "react-router-dom";
import { DndContext, KeyboardSensor, PointerSensor, closestCenter, useSensor, useSensors } from "@dnd-kit/core";
import { SortableContext, arrayMove, sortableKeyboardCoordinates, useSortable, verticalListSortingStrategy } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import {
  calculateFormulaPreviewRequest,
  createPsychologistQuestionRequest,
  createFormulaRuleRequest,
  deleteFormulaRuleRequest,
  deletePsychologistQuestionRequest,
  getPsychologistTestRequest,
  listFormulaRulesRequest,
  listPsychologistQuestionsRequest,
  listPsychologistReportTemplatesRequest,
  publishPsychologistTestRequest,
  updateFormulaRuleRequest,
  updatePsychologistQuestionRequest,
  updatePsychologistTestRequest,
} from "../../modules/tests/api/testsApi";
import { getMetricLabel } from "../../modules/psychologist/lib/psychologistUi";
import { buildClientSessionPath, ROUTES } from "../../shared/config/routes";
import PageCard from "../../shared/ui/PageCard";

const QUESTION_TYPE_ITEMS = [
  { value: "single_choice", label: "Один вариант" },
  { value: "multiple_choice", label: "Несколько вариантов" },
  { value: "scale", label: "Шкала" },
  { value: "text", label: "Текстовый ответ" },
  { value: "number", label: "Числовой ответ" },
];

const QUESTION_TYPES_WITH_OPTIONS = new Set(["single_choice", "multiple_choice", "scale"]);
const CONDITION_TYPE_ITEMS = [
  { value: "always", label: "Всегда" },
  { value: "answer_equals", label: "Ответ равен" },
  { value: "answer_in", label: "Ответ в списке" },
  { value: "answer_numeric_gte", label: "Ответ >= значения" },
  { value: "answer_numeric_lte", label: "Ответ <= значения" },
];
const DEFAULT_METRIC_KEYS = [
  "total",
  "analytic",
  "creative",
  "social",
  "organizer",
  "practical",
  "stress_resistance",
  "leadership",
  "communication",
];

function toResultsPath(id) {
  return ROUTES.testResults.replace(":id", String(id));
}

function needsOptions(questionType) {
  return QUESTION_TYPES_WITH_OPTIONS.has(questionType);
}

function toOptionalInteger(value) {
  const normalizedValue = String(value ?? "").trim();

  if (!normalizedValue) {
    return undefined;
  }

  const parsed = Number(normalizedValue);
  if (!Number.isInteger(parsed) || parsed <= 0) {
    return undefined;
  }

  return parsed;
}

function toOptionalNumber(value) {
  const normalizedValue = String(value ?? "").trim();

  if (!normalizedValue) {
    return undefined;
  }

  const parsed = Number(normalizedValue);
  return Number.isFinite(parsed) ? parsed : undefined;
}

function createTestForm(test = {}) {
  return {
    title: String(test.title || ""),
    description: String(test.description || ""),
    report_template_id:
      test.report_template_id === null || test.report_template_id === undefined
        ? ""
        : String(test.report_template_id),
    recommended_duration:
      test.recommended_duration === null || test.recommended_duration === undefined
        ? ""
        : String(test.recommended_duration),
    has_participant_limit: Boolean(test.has_participant_limit),
    max_participants:
      test.max_participants === null || test.max_participants === undefined
        ? ""
        : String(test.max_participants),
    collect_respondent_age: Boolean(test.collect_respondent_age),
    collect_respondent_gender: Boolean(test.collect_respondent_gender),
    collect_respondent_education: Boolean(test.collect_respondent_education),
    show_client_report_immediately: Boolean(test.show_client_report_immediately),
  };
}

function createOptionDraft(index, overrides = {}) {
  const fallbackLabel = `Вариант ${index}`;
  const fallbackValue = `option_${index}`;

  return {
    label: String(overrides.label ?? fallbackLabel),
    value: String(overrides.value ?? fallbackValue),
    order_number: index,
    score:
      overrides.score === null || overrides.score === undefined || overrides.score === ""
        ? ""
        : String(overrides.score),
  };
}

function createDefaultOptions(questionType) {
  if (questionType === "scale") {
    return Array.from({ length: 5 }, (_, index) =>
      createOptionDraft(index + 1, {
        label: String(index + 1),
        value: String(index + 1),
        score: String(index + 1),
      }),
    );
  }

  if (needsOptions(questionType)) {
    return [createOptionDraft(1), createOptionDraft(2)];
  }

  return [];
}

function createQuestionDraft(orderNumber = 1) {
  return {
    text: "",
    question_type: "single_choice",
    is_required: true,
    order_number: orderNumber,
    options: createDefaultOptions("single_choice"),
  };
}

function createFormulaDraft() {
  return {
    name: "",
    question_id: "0",
    condition_type: "always",
    expected_value: "",
    score_delta: "0",
    result_key: "total",
    priority: "0",
  };
}

function normalizeOption(option, index) {
  return createOptionDraft(index + 1, {
    label: option?.label,
    value: option?.value,
    score: option?.score,
  });
}

function normalizeQuestion(question, index) {
  const questionType = question?.question_type || "text";

  return {
    ...question,
    text: String(question?.text || ""),
    question_type: questionType,
    is_required: Boolean(question?.is_required),
    order_number: Number(question?.order_number) || index + 1,
    options: needsOptions(questionType)
      ? (Array.isArray(question?.options) ? question.options : []).map(normalizeOption)
      : [],
  };
}

function normalizeFormulaRule(rule) {
  return {
    id: rule.id,
    name: String(rule.name || ""),
    question_id:
      rule.question_id === null || rule.question_id === undefined ? "0" : String(rule.question_id),
    condition_type: String(rule.condition_type || "always"),
    expected_value: String(rule.expected_value || ""),
    score_delta:
      rule.score_delta === null || rule.score_delta === undefined ? "0" : String(rule.score_delta),
    result_key: String(rule.result_key || "total"),
    priority: rule.priority === null || rule.priority === undefined ? "0" : String(rule.priority),
  };
}

function sortQuestions(questions) {
  return [...questions].sort((left, right) => {
    const orderDelta = (left.order_number || 0) - (right.order_number || 0);

    if (orderDelta !== 0) {
      return orderDelta;
    }

    return String(left.id).localeCompare(String(right.id));
  });
}

function buildTestPayload(form, currentTest) {
  const payload = {
    title: form.title.trim(),
    description: form.description.trim(),
    recommended_duration: toOptionalInteger(form.recommended_duration),
    has_participant_limit: Boolean(form.has_participant_limit),
    max_participants: form.has_participant_limit ? toOptionalInteger(form.max_participants) : undefined,
    collect_respondent_age: Boolean(form.collect_respondent_age),
    collect_respondent_gender: Boolean(form.collect_respondent_gender),
    collect_respondent_education: Boolean(form.collect_respondent_education),
    show_client_report_immediately: Boolean(form.show_client_report_immediately),
    status: currentTest?.status || "draft",
  };

  if (String(form.report_template_id || "").trim()) {
    payload.report_template_id = Number(form.report_template_id);
  } else if (currentTest?.report_template_id) {
    payload.report_template_id = currentTest.report_template_id;
  }

  return payload;
}

function buildQuestionPayload(question, index) {
  const payload = {
    text: String(question.text || "").trim(),
    question_type: question.question_type,
    is_required: Boolean(question.is_required),
    order_number: Number(question.order_number) || index + 1,
  };

  if (needsOptions(question.question_type)) {
    payload.options = (Array.isArray(question.options) ? question.options : []).map((option, optionIndex) => {
      const normalized = normalizeOption(option, optionIndex);
      const optionPayload = {
        label: normalized.label.trim() || `Вариант ${optionIndex + 1}`,
        value: normalized.value.trim() || `option_${optionIndex + 1}`,
        order_number: optionIndex + 1,
      };
      const numericScore = toOptionalNumber(normalized.score);

      if (numericScore !== undefined) {
        optionPayload.score = numericScore;
      }

      return optionPayload;
    });
  }

  return payload;
}

function buildFormulaRulePayload(rule) {
  const payload = {
    name: String(rule.name || "").trim(),
    condition_type: rule.condition_type,
  };

  const questionId = Number(rule.question_id || 0);
  if (questionId > 0) {
    payload.question_id = questionId;
  }

  const expectedValue = String(rule.expected_value || "").trim();
  if (expectedValue) {
    payload.expected_value = expectedValue;
  }

  const scoreDelta = String(rule.score_delta || "").trim();
  if (scoreDelta) {
    payload.score_delta = Number(scoreDelta);
  }

  const resultKey = String(rule.result_key || "").trim();
  if (resultKey) {
    payload.result_key = resultKey;
  }

  const priority = String(rule.priority || "").trim();
  if (priority) {
    payload.priority = Number(priority);
  }

  return payload;
}

function validateTestForm(form) {
  const errors = {};

  if (!String(form.title || "").trim()) {
    errors.title = "Укажите название методики.";
  }

  if (String(form.recommended_duration || "").trim() && !toOptionalInteger(form.recommended_duration)) {
    errors.recommended_duration = "Длительность должна быть положительным числом минут.";
  }

  if (form.has_participant_limit && !toOptionalInteger(form.max_participants)) {
    errors.max_participants = "Укажите положительное число участников.";
  }

  return errors;
}

function validateQuestionForm(form) {
  const errors = {};

  if (!String(form.text || "").trim()) {
    errors.text = "Добавьте текст вопроса.";
  }

  if (!QUESTION_TYPE_ITEMS.some((item) => item.value === form.question_type)) {
    errors.question_type = "Укажите корректный тип вопроса.";
  }

  if (needsOptions(form.question_type)) {
    const options = Array.isArray(form.options) ? form.options : [];

    if (options.length < 2) {
      errors.options = "Нужно минимум два варианта ответа.";
      return errors;
    }

    const seenValues = new Set();

    for (const option of options) {
      const label = String(option?.label || "").trim();
      const value = String(option?.value || "").trim();
      const score = String(option?.score ?? "").trim();

      if (!label) {
        errors.options = "У каждого варианта должно быть название.";
        break;
      }

      if (!value) {
        errors.options = "У каждого варианта должен быть value.";
        break;
      }

      if (seenValues.has(value)) {
        errors.options = "Value у вариантов ответа должны быть уникальными.";
        break;
      }

      seenValues.add(value);

      if (form.question_type === "scale" && score === "") {
        errors.options = "Для шкалы у каждого варианта должен быть балл.";
        break;
      }

      if (score !== "" && !Number.isFinite(Number(score))) {
        errors.options = "Балл варианта должен быть числом.";
        break;
      }
    }
  }

  return errors;
}

function validateFormulaRuleForm(rule) {
  const errors = {};
  const conditionType = String(rule.condition_type || "");
  const name = String(rule.name || "").trim();
  const expectedValue = String(rule.expected_value || "").trim();
  const scoreDelta = String(rule.score_delta || "").trim();
  const priority = String(rule.priority || "").trim();

  if (!name) {
    errors.name = "Укажите название правила.";
  }

  if (!CONDITION_TYPE_ITEMS.some((item) => item.value === conditionType)) {
    errors.condition_type = "Укажите корректное условие.";
  }

  if (conditionType !== "always" && !expectedValue) {
    errors.expected_value = "Для выбранного условия нужно expected_value.";
  }

  if (scoreDelta && !Number.isFinite(Number(scoreDelta))) {
    errors.score_delta = "Изменение балла должно быть числом.";
  }

  if (priority && !Number.isInteger(Number(priority))) {
    errors.priority = "Приоритет должен быть целым числом.";
  }

  return errors;
}

function getFirstError(errors) {
  return Object.values(errors || {})[0] || "";
}

function collectPublishIssues(testForm, questions) {
  const issues = [];
  const testErrors = validateTestForm(testForm);

  Object.values(testErrors).forEach((message) => {
    issues.push(message);
  });

  if (!questions.length) {
    issues.push("Добавьте хотя бы один вопрос.");
  }

  questions.forEach((question, index) => {
    const questionErrors = validateQuestionForm(question);
    const firstError = getFirstError(questionErrors);

    if (firstError) {
      issues.push(`Вопрос #${index + 1}: ${firstError}`);
    }
  });

  return issues;
}

function getQuestionTypeLabel(questionType) {
  return QUESTION_TYPE_ITEMS.find((item) => item.value === questionType)?.label || questionType;
}

function getTestStatusLabel(status) {
  return status === "published" ? "Опубликован" : "Черновик";
}

function formatDate(value) {
  if (!value) {
    return "—";
  }

  const timestamp = new Date(value).getTime();
  if (!Number.isFinite(timestamp)) {
    return "—";
  }

  return new Date(value).toLocaleDateString("ru-RU", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
  });
}

function reorderQuestions(questions, sourceId, targetId) {
  const sourceIndex = questions.findIndex((item) => String(item.id) === String(sourceId));
  const targetIndex = questions.findIndex((item) => String(item.id) === String(targetId));

  if (sourceIndex === -1 || targetIndex === -1 || sourceIndex === targetIndex) {
    return questions;
  }

  return arrayMove(questions, sourceIndex, targetIndex).map((item, index) => ({
    ...item,
    order_number: index + 1,
  }));
}

function SortableQuestionCard({
  question,
  index,
  questionErrors,
  isBusy,
  isReordering,
  onSave,
  onDelete,
  onQuestionFieldChange,
  onQuestionOptionChange,
  onRemoveOption,
  onAddOption,
}) {
  const { attributes, listeners, setActivatorNodeRef, setNodeRef, transform, transition, isDragging, isSorting } =
    useSortable({
      id: String(question.id),
      disabled: isReordering,
    });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
  };

  return (
    <article
      ref={setNodeRef}
      style={style}
      className={`builder-question-card ${isDragging ? "builder-question-card--dragging" : ""} ${
        isSorting ? "builder-question-card--sorting" : ""
      }`}
    >
      <div className="builder-question-card__head">
        <div className="builder-question-card__meta">
          <button
            ref={setActivatorNodeRef}
            className="builder-drag-handle"
            title="Перетащите вопрос"
            type="button"
            {...attributes}
            {...listeners}
          >
            <GripVertical size={16} strokeWidth={2.1} />
          </button>
          <strong>#{index + 1}</strong>
          <span>{getQuestionTypeLabel(question.question_type)}</span>
          <span>{question.is_required ? "обязательный" : "необязательный"}</span>
        </div>

        <div className="admin-table__actions">
          <button className="table-action-button" disabled={isBusy} type="button" onClick={() => onSave(question.id)}>
            {isBusy ? (
              <LoaderCircle className="icon-spin" size={15} strokeWidth={2.1} />
            ) : (
              <Save size={15} strokeWidth={2.1} />
            )}
            <span>Сохранить</span>
          </button>
          <button className="table-action-button" disabled={isBusy} type="button" onClick={() => onDelete(question.id)}>
            <Trash2 size={15} strokeWidth={2.1} />
            <span>Удалить</span>
          </button>
        </div>
      </div>

      <div className="builder-question-card__body">
        <label className="admin-form-field admin-form-field--wide">
          <span>Текст вопроса</span>
          <textarea
            aria-invalid={Boolean(questionErrors.text)}
            className={
              questionErrors.text
                ? "admin-form-control admin-form-control--invalid builder-textarea"
                : "admin-form-control builder-textarea"
            }
            rows={3}
            value={question.text}
            onChange={(event) => onQuestionFieldChange(question.id, "text", event.target.value)}
            placeholder="Сформулируйте вопрос"
          />
          {questionErrors.text ? (
            <small className="admin-form-error admin-form-error--inline">{questionErrors.text}</small>
          ) : null}
        </label>

        <div className="builder-question-card__row">
          <label className="admin-form-field">
            <span>Тип вопроса</span>
            <select
              aria-invalid={Boolean(questionErrors.question_type)}
              className={questionErrors.question_type ? "admin-form-control admin-form-control--invalid" : "admin-form-control"}
              value={question.question_type}
              onChange={(event) => onQuestionFieldChange(question.id, "question_type", event.target.value)}
            >
              {QUESTION_TYPE_ITEMS.map((item) => (
                <option key={`${question.id}-${item.value}`} value={item.value}>
                  {item.label}
                </option>
              ))}
            </select>
          </label>

          <label className="builder-toggle builder-toggle--inline">
            <span>Обязательный вопрос</span>
            <input
              checked={question.is_required}
              type="checkbox"
              onChange={(event) => onQuestionFieldChange(question.id, "is_required", event.target.checked)}
            />
          </label>
        </div>

        {needsOptions(question.question_type) ? (
          <div className="admin-form-field admin-form-field--wide">
            <span>Варианты ответа</span>
            <div className="builder-options">
              <div className="builder-option-row builder-option-row--head" aria-hidden="true">
                <span>Подпись</span>
                <span>Value</span>
                <span>Балл</span>
                <span />
              </div>
              {question.options.map((option, optionIndex) => (
                <div className="builder-option-row" key={`${question.id}-${optionIndex + 1}`}>
                  <input
                    className="admin-form-control"
                    value={option.label}
                    onChange={(event) => onQuestionOptionChange(question.id, optionIndex, "label", event.target.value)}
                    placeholder={`Вариант ${optionIndex + 1}`}
                  />
                  <input
                    className="admin-form-control"
                    value={option.value}
                    onChange={(event) => onQuestionOptionChange(question.id, optionIndex, "value", event.target.value)}
                    placeholder={`option_${optionIndex + 1}`}
                  />
                  <input
                    className="admin-form-control"
                    inputMode="decimal"
                    type="number"
                    value={option.score}
                    onChange={(event) => onQuestionOptionChange(question.id, optionIndex, "score", event.target.value)}
                    placeholder={question.question_type === "scale" ? "1" : "опц."}
                  />
                  <button
                    className="table-action-button"
                    disabled={question.options.length <= 2}
                    type="button"
                    onClick={() => onRemoveOption(question.id, optionIndex)}
                  >
                    <Trash2 size={15} strokeWidth={2.1} />
                    <span>Убрать</span>
                  </button>
                </div>
              ))}
            </div>
            <div className="builder-option-actions">
              <button className="table-action-button" type="button" onClick={() => onAddOption(question.id)}>
                <Plus size={15} strokeWidth={2.1} />
                <span>Добавить вариант</span>
              </button>
            </div>
            {questionErrors.options ? (
              <small className="admin-form-error admin-form-error--inline">{questionErrors.options}</small>
            ) : null}
          </div>
        ) : null}
      </div>
    </article>
  );
}

export default function BuilderPage() {
  const { id } = useParams();
  const testQuery = useSWR(id ? ["psychologist-test", id] : null, () => getPsychologistTestRequest(id), {
    revalidateOnFocus: false,
    shouldRetryOnError: false,
  });
  const questionsQuery = useSWR(
    id ? ["psychologist-test-questions", id] : null,
    () => listPsychologistQuestionsRequest(id),
    {
      revalidateOnFocus: false,
      shouldRetryOnError: false,
    },
  );
  const formulaRulesQuery = useSWR(
    id ? ["psychologist-test-formulas", id] : null,
    () => listFormulaRulesRequest(id),
    {
      revalidateOnFocus: false,
      shouldRetryOnError: false,
    },
  );
  const reportTemplatesQuery = useSWR(
    "psychologist-report-templates",
    listPsychologistReportTemplatesRequest,
    {
      revalidateOnFocus: false,
      shouldRetryOnError: false,
    },
  );

  const [testForm, setTestForm] = useState(createTestForm());
  const [testErrors, setTestErrors] = useState({});
  const [draftQuestions, setDraftQuestions] = useState([]);
  const [newQuestionForm, setNewQuestionForm] = useState(createQuestionDraft());
  const [newQuestionErrors, setNewQuestionErrors] = useState({});
  const [questionErrorsById, setQuestionErrorsById] = useState({});
  const [draftFormulaRules, setDraftFormulaRules] = useState([]);
  const [newFormulaForm, setNewFormulaForm] = useState(createFormulaDraft());
  const [newFormulaErrors, setNewFormulaErrors] = useState({});
  const [formulaErrorsById, setFormulaErrorsById] = useState({});
  const [previewAnswersByQuestion, setPreviewAnswersByQuestion] = useState({});
  const [previewResult, setPreviewResult] = useState(null);
  const [previewError, setPreviewError] = useState("");
  const [feedbackMessage, setFeedbackMessage] = useState("");
  const [feedbackError, setFeedbackError] = useState("");
  const [isSavingTest, setIsSavingTest] = useState(false);
  const [isPublishing, setIsPublishing] = useState(false);
  const [isCreatingQuestion, setIsCreatingQuestion] = useState(false);
  const [isCreatingFormula, setIsCreatingFormula] = useState(false);
  const [isPreviewingFormula, setIsPreviewingFormula] = useState(false);
  const [activeQuestionId, setActiveQuestionId] = useState(null);
  const [activeFormulaId, setActiveFormulaId] = useState(null);
  const [isReordering, setIsReordering] = useState(false);

  const dndSensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: 8,
      },
    }),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    }),
  );

  useEffect(() => {
    if (!testQuery.data) {
      return;
    }

    setTestForm(createTestForm(testQuery.data));
  }, [testQuery.data]);

  useEffect(() => {
    if (!questionsQuery.data) {
      return;
    }

    const normalized = sortQuestions(questionsQuery.data.map(normalizeQuestion));
    setDraftQuestions(normalized);
    setNewQuestionForm((prev) => ({
      ...prev,
      order_number: normalized.length + 1,
    }));
  }, [questionsQuery.data]);

  useEffect(() => {
    if (!formulaRulesQuery.data) {
      return;
    }

    setDraftFormulaRules(formulaRulesQuery.data.map(normalizeFormulaRule));
  }, [formulaRulesQuery.data]);

  useEffect(() => {
    if (!feedbackMessage) {
      return undefined;
    }

    const timeoutId = window.setTimeout(() => setFeedbackMessage(""), 2600);
    return () => window.clearTimeout(timeoutId);
  }, [feedbackMessage]);

  const questionCount = draftQuestions.length;
  const publicSlug = testQuery.data?.public_slug || "";
  const publicPath = publicSlug ? buildClientSessionPath(publicSlug) : "";
  const publicUrl = publicPath ? `${window.location.origin}${publicPath}` : "";
  const reportTemplates = useMemo(
    () => (Array.isArray(reportTemplatesQuery.data) ? reportTemplatesQuery.data : []),
    [reportTemplatesQuery.data],
  );
  const selectedReportTemplate = useMemo(
    () =>
      reportTemplates.find((template) => String(template.id) === String(testForm.report_template_id || "")) || null,
    [reportTemplates, testForm.report_template_id],
  );
  const publishIssues = useMemo(() => collectPublishIssues(testForm, draftQuestions), [testForm, draftQuestions]);
  const metricKeySuggestions = useMemo(
    () =>
      Array.from(
        new Set([
          ...DEFAULT_METRIC_KEYS,
          ...draftFormulaRules.map((rule) => String(rule.result_key || "").trim()),
          ...Object.keys(previewResult?.metrics || {}),
        ]),
      ).filter(Boolean),
    [draftFormulaRules, previewResult],
  );

  function clearFeedbackError() {
    if (feedbackError) {
      setFeedbackError("");
    }
  }

  function clearTestError(field) {
    setTestErrors((prev) => {
      if (!prev[field]) {
        return prev;
      }

      const next = { ...prev };
      delete next[field];
      return next;
    });
  }

  function clearNewQuestionError(field) {
    setNewQuestionErrors((prev) => {
      if (!prev[field]) {
        return prev;
      }

      const next = { ...prev };
      delete next[field];
      return next;
    });
  }

  function clearQuestionError(questionId, field) {
    setQuestionErrorsById((prev) => {
      const current = prev[questionId];
      if (!current || !current[field]) {
        return prev;
      }

      const next = { ...prev };
      next[questionId] = { ...current };
      delete next[questionId][field];

      if (!Object.keys(next[questionId]).length) {
        delete next[questionId];
      }

      return next;
    });
  }

  function clearNewFormulaError(field) {
    setNewFormulaErrors((prev) => {
      if (!prev[field]) {
        return prev;
      }

      const next = { ...prev };
      delete next[field];
      return next;
    });
  }

  function clearFormulaError(ruleId, field) {
    setFormulaErrorsById((prev) => {
      const current = prev[ruleId];
      if (!current || !current[field]) {
        return prev;
      }

      const next = { ...prev };
      next[ruleId] = { ...current };
      delete next[ruleId][field];

      if (!Object.keys(next[ruleId]).length) {
        delete next[ruleId];
      }

      return next;
    });
  }

  function handleTestFieldChange(field, value) {
    setTestForm((prev) => {
      const next = { ...prev, [field]: value };

      if (field === "has_participant_limit" && !value) {
        next.max_participants = "";
      }

      return next;
    });
    clearFeedbackError();
    clearTestError(field);

    if (field === "has_participant_limit") {
      clearTestError("max_participants");
    }
  }

  function handleNewQuestionFieldChange(field, value) {
    setNewQuestionForm((prev) => {
      const next = { ...prev, [field]: value };

      if (field === "question_type") {
        next.options = createDefaultOptions(value);
      }

      return next;
    });

    clearFeedbackError();
    clearNewQuestionError(field);
    clearNewQuestionError("options");
  }

  function handleNewQuestionOptionChange(optionIndex, field, value) {
    setNewQuestionForm((prev) => ({
      ...prev,
      options: prev.options.map((option, index) =>
        index === optionIndex ? { ...option, [field]: value } : option,
      ),
    }));
    clearFeedbackError();
    clearNewQuestionError("options");
  }

  function addNewQuestionOption() {
    setNewQuestionForm((prev) => ({
      ...prev,
      options: [...prev.options, createOptionDraft(prev.options.length + 1)],
    }));
    clearFeedbackError();
    clearNewQuestionError("options");
  }

  function removeNewQuestionOption(optionIndex) {
    setNewQuestionForm((prev) => ({
      ...prev,
      options: prev.options.filter((_, index) => index !== optionIndex),
    }));
    clearFeedbackError();
    clearNewQuestionError("options");
  }

  function handleQuestionFieldChange(questionId, field, value) {
    setDraftQuestions((prev) =>
      prev.map((question) => {
        if (String(question.id) !== String(questionId)) {
          return question;
        }

        const next = { ...question, [field]: value };

        if (field === "question_type") {
          next.options = createDefaultOptions(value);
        }

        return next;
      }),
    );

    clearFeedbackError();
    clearQuestionError(questionId, field);
    clearQuestionError(questionId, "options");
  }

  function handleQuestionOptionChange(questionId, optionIndex, field, value) {
    setDraftQuestions((prev) =>
      prev.map((question) => {
        if (String(question.id) !== String(questionId)) {
          return question;
        }

        return {
          ...question,
          options: question.options.map((option, index) =>
            index === optionIndex ? { ...option, [field]: value } : option,
          ),
        };
      }),
    );

    clearFeedbackError();
    clearQuestionError(questionId, "options");
  }

  function addOptionToQuestion(questionId) {
    setDraftQuestions((prev) =>
      prev.map((question) => {
        if (String(question.id) !== String(questionId)) {
          return question;
        }

        return {
          ...question,
          options: [...question.options, createOptionDraft(question.options.length + 1)],
        };
      }),
    );
    clearFeedbackError();
    clearQuestionError(questionId, "options");
  }

  function removeOptionFromQuestion(questionId, optionIndex) {
    setDraftQuestions((prev) =>
      prev.map((question) => {
        if (String(question.id) !== String(questionId)) {
          return question;
        }

        return {
          ...question,
          options: question.options.filter((_, index) => index !== optionIndex),
        };
      }),
    );
    clearFeedbackError();
    clearQuestionError(questionId, "options");
  }

  function handleNewFormulaFieldChange(field, value) {
    setNewFormulaForm((prev) => ({
      ...prev,
      [field]: value,
      ...(field === "condition_type" && value === "always" ? { expected_value: "" } : {}),
    }));
    clearFeedbackError();
    clearNewFormulaError(field);
  }

  function handleFormulaFieldChange(ruleId, field, value) {
    setDraftFormulaRules((prev) =>
      prev.map((rule) =>
        String(rule.id) === String(ruleId)
          ? {
              ...rule,
              [field]: value,
              ...(field === "condition_type" && value === "always" ? { expected_value: "" } : {}),
            }
          : rule,
      ),
    );
    clearFeedbackError();
    clearFormulaError(ruleId, field);
  }

  function handlePreviewAnswerChange(questionId, patch) {
    setPreviewAnswersByQuestion((prev) => ({
      ...prev,
      [String(questionId)]: {
        ...prev[String(questionId)],
        ...patch,
      },
    }));
    setPreviewError("");
  }

  async function handleCreateFormula(event) {
    event.preventDefault();
    const errors = validateFormulaRuleForm(newFormulaForm);
    setNewFormulaErrors(errors);

    if (Object.keys(errors).length) {
      return;
    }

    setIsCreatingFormula(true);
    setFeedbackError("");

    try {
      await createFormulaRuleRequest(id, buildFormulaRulePayload(newFormulaForm));
      setNewFormulaForm(createFormulaDraft());
      setNewFormulaErrors({});
      await formulaRulesQuery.mutate();
      setFeedbackMessage("Правило расчёта добавлено.");
    } catch (error) {
      setFeedbackError(error?.message || "Не удалось создать правило расчёта.");
    } finally {
      setIsCreatingFormula(false);
    }
  }

  async function handleSaveFormula(ruleId) {
    const rule = draftFormulaRules.find((item) => String(item.id) === String(ruleId));
    if (!rule) {
      return;
    }

    const errors = validateFormulaRuleForm(rule);
    if (Object.keys(errors).length) {
      setFormulaErrorsById((prev) => ({ ...prev, [ruleId]: errors }));
      return;
    }

    setActiveFormulaId(ruleId);
    setFeedbackError("");

    try {
      await updateFormulaRuleRequest(id, ruleId, buildFormulaRulePayload(rule));
      await formulaRulesQuery.mutate();
      setFeedbackMessage("Правило обновлено.");
    } catch (error) {
      setFeedbackError(error?.message || "Не удалось сохранить правило.");
    } finally {
      setActiveFormulaId(null);
    }
  }

  async function handleDeleteFormula(ruleId) {
    setActiveFormulaId(ruleId);
    setFeedbackError("");

    try {
      await deleteFormulaRuleRequest(id, ruleId);
      await formulaRulesQuery.mutate();
      setFormulaErrorsById((prev) => {
        if (!prev[ruleId]) {
          return prev;
        }

        const next = { ...prev };
        delete next[ruleId];
        return next;
      });
      setFeedbackMessage("Правило удалено.");
    } catch (error) {
      setFeedbackError(error?.message || "Не удалось удалить правило.");
    } finally {
      setActiveFormulaId(null);
    }
  }

  async function handleCalculateFormulaPreview() {
    const answers = draftQuestions.reduce((accumulator, question) => {
      const value = previewAnswersByQuestion[String(question.id)];

      if (!value) {
        return accumulator;
      }

      if (question.question_type === "multiple_choice" && Array.isArray(value.answer_values) && value.answer_values.length) {
        accumulator.push({
          question_id: Number(question.id),
          answer_values: value.answer_values,
        });
      } else if (value.answer_value) {
        accumulator.push({
          question_id: Number(question.id),
          answer_value: String(value.answer_value),
        });
      }

      return accumulator;
    }, []);

    setIsPreviewingFormula(true);
    setPreviewError("");

    try {
      const response = await calculateFormulaPreviewRequest(id, {
        answers,
      });
      setPreviewResult(response);
    } catch (error) {
      setPreviewError(error?.message || "Не удалось рассчитать preview.");
    } finally {
      setIsPreviewingFormula(false);
    }
  }

  async function handleSaveTest(event) {
    event.preventDefault();
    const errors = validateTestForm(testForm);
    setTestErrors(errors);

    if (Object.keys(errors).length) {
      return;
    }

    setIsSavingTest(true);
    setFeedbackError("");

    try {
      await updatePsychologistTestRequest(id, buildTestPayload(testForm, testQuery.data));
      await testQuery.mutate();
      setFeedbackMessage("Параметры теста сохранены.");
    } catch (error) {
      setFeedbackError(error?.message || "Не удалось сохранить параметры теста.");
    } finally {
      setIsSavingTest(false);
    }
  }

  async function handlePublishTest() {
    const nextTestErrors = validateTestForm(testForm);
    const nextQuestionErrors = {};

    draftQuestions.forEach((question) => {
      const questionErrors = validateQuestionForm(question);
      if (Object.keys(questionErrors).length) {
        nextQuestionErrors[question.id] = questionErrors;
      }
    });

    setTestErrors(nextTestErrors);
    setQuestionErrorsById(nextQuestionErrors);

    if (Object.keys(nextTestErrors).length || Object.keys(nextQuestionErrors).length || !draftQuestions.length) {
      setFeedbackError("Перед публикацией исправь проблемы в методике. Список есть в правой колонке.");
      return;
    }

    setIsPublishing(true);
    setFeedbackError("");

    try {
      await updatePsychologistTestRequest(id, buildTestPayload(testForm, testQuery.data));
      await publishPsychologistTestRequest(id);
      await Promise.all([testQuery.mutate(), questionsQuery.mutate()]);
      setFeedbackMessage("Тест опубликован. Публичная ссылка уже доступна.");
    } catch (error) {
      setFeedbackError(error?.message || "Не удалось опубликовать тест.");
    } finally {
      setIsPublishing(false);
    }
  }

  async function handleCreateQuestion(event) {
    event.preventDefault();
    const payload = {
      ...newQuestionForm,
      order_number: draftQuestions.length + 1,
    };
    const errors = validateQuestionForm(payload);
    setNewQuestionErrors(errors);

    if (Object.keys(errors).length) {
      return;
    }

    setIsCreatingQuestion(true);
    setFeedbackError("");

    try {
      await createPsychologistQuestionRequest(id, buildQuestionPayload(payload, draftQuestions.length));
      setNewQuestionForm(createQuestionDraft(draftQuestions.length + 2));
      setNewQuestionErrors({});
      await questionsQuery.mutate();
      setFeedbackMessage("Вопрос добавлен в методику.");
    } catch (error) {
      setFeedbackError(error?.message || "Не удалось создать вопрос.");
    } finally {
      setIsCreatingQuestion(false);
    }
  }

  async function handleSaveQuestion(questionId) {
    const question = draftQuestions.find((item) => String(item.id) === String(questionId));
    if (!question) {
      return;
    }

    const errors = validateQuestionForm(question);
    if (Object.keys(errors).length) {
      setQuestionErrorsById((prev) => ({ ...prev, [questionId]: errors }));
      return;
    }

    setActiveQuestionId(questionId);
    setFeedbackError("");

    try {
      await updatePsychologistQuestionRequest(
        id,
        questionId,
        buildQuestionPayload(question, draftQuestions.findIndex((item) => String(item.id) === String(questionId))),
      );
      await questionsQuery.mutate();
      setFeedbackMessage("Вопрос сохранён.");
    } catch (error) {
      setFeedbackError(error?.message || "Не удалось сохранить вопрос.");
    } finally {
      setActiveQuestionId(null);
    }
  }

  async function handleDeleteQuestion(questionId) {
    const previousQuestions = draftQuestions;
    setActiveQuestionId(questionId);
    setFeedbackError("");

    try {
      await deletePsychologistQuestionRequest(id, questionId);

      const remainingQuestions = previousQuestions.filter((item) => String(item.id) !== String(questionId));
      const normalizedRemaining = remainingQuestions.map((question, index) => ({
        ...question,
        order_number: index + 1,
      }));
      const changedQuestions = normalizedRemaining.filter(
        (question, index) => Number(remainingQuestions[index].order_number) !== index + 1,
      );

      if (changedQuestions.length) {
        for (const question of changedQuestions) {
          await updatePsychologistQuestionRequest(
            id,
            question.id,
            buildQuestionPayload(question, question.order_number - 1),
          );
        }
      }

      await questionsQuery.mutate();
      setQuestionErrorsById((prev) => {
        if (!prev[questionId]) {
          return prev;
        }

        const next = { ...prev };
        delete next[questionId];
        return next;
      });
      setFeedbackMessage("Вопрос удалён.");
    } catch (error) {
      setFeedbackError(error?.message || "Не удалось удалить вопрос.");
    } finally {
      setActiveQuestionId(null);
    }
  }

  async function handleCopyPublicUrl() {
    if (!publicUrl) {
      setFeedbackError("Публичная ссылка появится после публикации теста.");
      return;
    }

    setFeedbackError("");

    try {
      await navigator.clipboard.writeText(publicUrl);
      setFeedbackMessage("Публичная ссылка скопирована.");
    } catch {
      setFeedbackError("Не удалось скопировать ссылку в буфер обмена.");
    }
  }

  async function handleQuestionDragEnd(event) {
    const activeId = event?.active?.id;
    const overId = event?.over?.id;

    if (!activeId || !overId || String(activeId) === String(overId) || isReordering) {
      return;
    }

    const previousQuestions = draftQuestions;
    const reorderedQuestions = reorderQuestions(previousQuestions, activeId, overId);

    setDraftQuestions(reorderedQuestions);
    setIsReordering(true);
    setFeedbackError("");

    try {
      const changedQuestions = reorderedQuestions.filter((question, index) => {
        const previousQuestion = previousQuestions.find((item) => String(item.id) === String(question.id));
        return (previousQuestion?.order_number || 0) !== index + 1;
      });

      const temporaryOrderBase = previousQuestions.length + 1000;

      for (const [index, question] of changedQuestions.entries()) {
        await updatePsychologistQuestionRequest(id, question.id, {
          ...buildQuestionPayload(question, question.order_number - 1),
          order_number: temporaryOrderBase + index,
        });
      }

      for (const question of changedQuestions) {
        await updatePsychologistQuestionRequest(
          id,
          question.id,
          buildQuestionPayload(question, question.order_number - 1),
        );
      }

      await questionsQuery.mutate();
      setFeedbackMessage("Порядок вопросов обновлён.");
    } catch (error) {
      setDraftQuestions(previousQuestions);
      setFeedbackError(error?.message || "Не удалось обновить порядок вопросов.");
    } finally {
      setIsReordering(false);
    }
  }

  const sidebarStats = useMemo(
    () => [
      { label: "Вопросов", value: questionCount },
      { label: "Правил", value: draftFormulaRules.length },
      { label: "Статус", value: getTestStatusLabel(testQuery.data?.status) },
      { label: "Обновлён", value: formatDate(testQuery.data?.updated_at || testQuery.data?.created_at) },
    ],
    [draftFormulaRules.length, questionCount, testQuery.data],
  );

  return (
    <PageCard
      wide
      title={testQuery.data?.title || `Конструктор методики #${id}`}
      description="Редактор структуры теста, параметров сбора данных и порядка выдачи вопросов. Карточки можно перетаскивать мышью прямо в списке."
      links={[
        { to: ROUTES.tests, label: "К списку тестов" },
        { to: toResultsPath(id), label: "К результатам" },
        { to: ROUTES.reportTemplates, label: "Шаблоны отчётов" },
      ]}
    >
      {testQuery.error ? (
        <p className="admin-form-message admin-form-message--error">
          {testQuery.error.message || "Не удалось загрузить тест."}
        </p>
      ) : null}

      {questionsQuery.error ? (
        <p className="admin-form-message admin-form-message--error">
          {questionsQuery.error.message || "Не удалось загрузить вопросы."}
        </p>
      ) : null}

      {formulaRulesQuery.error ? (
        <p className="admin-form-message admin-form-message--error">
          {formulaRulesQuery.error.message || "Не удалось загрузить правила расчёта."}
        </p>
      ) : null}

      {feedbackError ? <p className="admin-form-message admin-form-message--error">{feedbackError}</p> : null}
      {feedbackMessage ? <p className="admin-form-message">{feedbackMessage}</p> : null}

      {testQuery.isLoading ? (
        <div className="builder-empty">Загружаем параметры теста...</div>
      ) : testQuery.data ? (
        <section className="builder-grid">
          <div className="builder-main">
            <section className="builder-panel builder-panel--hero">
              <div>
                <p className="builder-panel__eyebrow">Конструктор методики</p>
                <h2 className="builder-panel__title">Настройка теста</h2>
              </div>
              <div className="builder-panel__actions">
                <button className="table-action-button" type="button" onClick={handleCopyPublicUrl}>
                  <Copy size={15} strokeWidth={2.1} />
                  <span>Скопировать ссылку</span>
                </button>
                {publicPath ? (
                  <Link className="table-action-link" to={publicPath} target="_blank" rel="noreferrer">
                    <Eye size={15} strokeWidth={2.1} />
                    <span>Открыть форму</span>
                  </Link>
                ) : null}
                <button className="admin-primary-button" disabled={isPublishing} type="button" onClick={handlePublishTest}>
                  {isPublishing ? (
                    <LoaderCircle className="icon-spin" size={16} strokeWidth={2.1} />
                  ) : (
                    <Rocket size={16} strokeWidth={2.1} />
                  )}
                  <span>{testQuery.data.status === "published" ? "Переопубликовать" : "Опубликовать"}</span>
                </button>
              </div>
            </section>

            <section className="builder-panel">
              <div className="builder-section-head">
                <div>
                  <p className="builder-section-head__eyebrow">Паспорт методики</p>
                  <h3 className="builder-section-head__title">Основные параметры</h3>
                </div>
                <span className={`status-badge status-badge--${testQuery.data.status === "published" ? "active" : "draft"}`}>
                  {getTestStatusLabel(testQuery.data.status)}
                </span>
              </div>

              <form className="admin-form-grid" onSubmit={handleSaveTest}>
                <label className="admin-form-field">
                  <span>Название теста</span>
                  <input
                    aria-invalid={Boolean(testErrors.title)}
                    className={testErrors.title ? "admin-form-control admin-form-control--invalid" : "admin-form-control"}
                    value={testForm.title}
                    onChange={(event) => handleTestFieldChange("title", event.target.value)}
                    placeholder="Например: ПрофДНК. Базовый профориентационный опросник"
                  />
                  {testErrors.title ? <small className="admin-form-error admin-form-error--inline">{testErrors.title}</small> : null}
                </label>

                <label className="admin-form-field">
                  <span>Рекомендуемая длительность, мин</span>
                  <input
                    aria-invalid={Boolean(testErrors.recommended_duration)}
                    className={
                      testErrors.recommended_duration
                        ? "admin-form-control admin-form-control--invalid"
                        : "admin-form-control"
                    }
                    inputMode="numeric"
                    min="1"
                    type="number"
                    value={testForm.recommended_duration}
                    onChange={(event) => handleTestFieldChange("recommended_duration", event.target.value)}
                    placeholder="20"
                  />
                  {testErrors.recommended_duration ? (
                    <small className="admin-form-error admin-form-error--inline">{testErrors.recommended_duration}</small>
                  ) : null}
                </label>

                <label className="admin-form-field admin-form-field--wide">
                  <span>Описание</span>
                  <textarea
                    className="admin-form-control builder-textarea"
                    rows={4}
                    value={testForm.description}
                    onChange={(event) => handleTestFieldChange("description", event.target.value)}
                    placeholder="Кратко опишите методику, целевую аудиторию и что будет на выходе."
                  />
                </label>

                <label className="admin-form-field">
                  <span>Шаблон отчёта</span>
                  <select
                    className="admin-form-control"
                    value={testForm.report_template_id}
                    onChange={(event) => handleTestFieldChange("report_template_id", event.target.value)}
                  >
                    <option value="">Не выбран</option>
                    {reportTemplates.map((template) => (
                      <option key={`report-template-select-${template.id}`} value={template.id}>
                        {template.name}
                      </option>
                    ))}
                  </select>
                </label>

                <div className="admin-form-field admin-form-field--wide">
                  {reportTemplatesQuery.isLoading ? (
                    <div className="workflow-note">
                      <p>Загружаем доступные шаблоны отчётов.</p>
                    </div>
                  ) : selectedReportTemplate ? (
                    <div className="workflow-note workflow-note--success">
                      <p>
                        К тесту привязан шаблон <strong>{selectedReportTemplate.name}</strong>.
                        {selectedReportTemplate.description ? ` ${selectedReportTemplate.description}` : ""}
                      </p>
                      <div className="workflow-note__actions">
                        <Link className="table-action-link" to={ROUTES.reportTemplates}>
                          <LayoutTemplate size={15} strokeWidth={2.1} />
                          <span>Редактировать шаблоны</span>
                        </Link>
                      </div>
                    </div>
                  ) : reportTemplates.length ? (
                    <div className="workflow-note workflow-note--warning">
                      <p>Шаблон отчёта пока не выбран. Без него отчётный контур будет неполным.</p>
                      <div className="workflow-note__actions">
                        <Link className="table-action-link" to={ROUTES.reportTemplates}>
                          <LayoutTemplate size={15} strokeWidth={2.1} />
                          <span>Открыть шаблоны</span>
                        </Link>
                      </div>
                    </div>
                  ) : (
                    <div className="workflow-note workflow-note--warning">
                      <p>Шаблонов отчётов пока нет. Сначала создай шаблон, потом вернись и привяжи его к тесту.</p>
                      <div className="workflow-note__actions">
                        <Link className="table-action-link" to={ROUTES.reportTemplates}>
                          <LayoutTemplate size={15} strokeWidth={2.1} />
                          <span>Создать шаблон</span>
                        </Link>
                      </div>
                    </div>
                  )}
                </div>

                <div className="builder-toggle-grid admin-form-field--wide">
                  <label className="builder-toggle">
                    <span>Ограничить число участников</span>
                    <input
                      checked={testForm.has_participant_limit}
                      type="checkbox"
                      onChange={(event) => handleTestFieldChange("has_participant_limit", event.target.checked)}
                    />
                  </label>
                  <label className="builder-toggle">
                    <span>Собирать возраст</span>
                    <input
                      checked={testForm.collect_respondent_age}
                      type="checkbox"
                      onChange={(event) => handleTestFieldChange("collect_respondent_age", event.target.checked)}
                    />
                  </label>
                  <label className="builder-toggle">
                    <span>Собирать пол</span>
                    <input
                      checked={testForm.collect_respondent_gender}
                      type="checkbox"
                      onChange={(event) => handleTestFieldChange("collect_respondent_gender", event.target.checked)}
                    />
                  </label>
                  <label className="builder-toggle">
                    <span>Собирать образование</span>
                    <input
                      checked={testForm.collect_respondent_education}
                      type="checkbox"
                      onChange={(event) => handleTestFieldChange("collect_respondent_education", event.target.checked)}
                    />
                  </label>
                  <label className="builder-toggle">
                    <span>Показывать клиентский отчёт сразу</span>
                    <input
                      checked={testForm.show_client_report_immediately}
                      type="checkbox"
                      onChange={(event) =>
                        handleTestFieldChange("show_client_report_immediately", event.target.checked)
                      }
                    />
                  </label>
                </div>

                <div className="admin-form-field admin-form-field--wide">
                  <div
                    className={`workflow-note ${
                      testForm.show_client_report_immediately
                        ? "workflow-note--success"
                        : "workflow-note--warning"
                    }`}
                  >
                    <p>
                      {testForm.show_client_report_immediately
                        ? "После завершения теста клиент сможет сразу открыть HTML-отчёт по своей публичной сессии."
                        : "Мгновенный клиентский отчёт сейчас выключен. В этом режиме backend может возвращать 409 на публичный report endpoint, и итог клиенту останется доступен только через психолога."}
                    </p>
                  </div>
                </div>

                {testForm.has_participant_limit ? (
                  <label className="admin-form-field">
                    <span>Максимум участников</span>
                    <input
                      aria-invalid={Boolean(testErrors.max_participants)}
                      className={
                        testErrors.max_participants
                          ? "admin-form-control admin-form-control--invalid"
                          : "admin-form-control"
                      }
                      inputMode="numeric"
                      min="1"
                      type="number"
                      value={testForm.max_participants}
                      onChange={(event) => handleTestFieldChange("max_participants", event.target.value)}
                      placeholder="100"
                    />
                    {testErrors.max_participants ? (
                      <small className="admin-form-error admin-form-error--inline">{testErrors.max_participants}</small>
                    ) : null}
                  </label>
                ) : null}

                <div className="admin-form-actions">
                  <button className="admin-primary-button" disabled={isSavingTest} type="submit">
                    {isSavingTest ? (
                      <LoaderCircle className="icon-spin" size={16} strokeWidth={2.1} />
                    ) : (
                      <Save size={16} strokeWidth={2.1} />
                    )}
                    <span>{isSavingTest ? "Сохранение..." : "Сохранить параметры"}</span>
                  </button>
                </div>
              </form>
            </section>

            <section className="builder-panel">
              <div className="builder-section-head">
                <div>
                  <p className="builder-section-head__eyebrow">Сборка теста</p>
                  <h3 className="builder-section-head__title">Добавить вопрос</h3>
                </div>
                <span className="builder-section-head__meta">Новый вопрос сразу попадает в конец структуры.</span>
              </div>

              <form className="admin-form-grid" onSubmit={handleCreateQuestion}>
                <label className="admin-form-field admin-form-field--wide">
                  <span>Текст вопроса</span>
                  <textarea
                    aria-invalid={Boolean(newQuestionErrors.text)}
                    className={
                      newQuestionErrors.text
                        ? "admin-form-control admin-form-control--invalid builder-textarea"
                        : "admin-form-control builder-textarea"
                    }
                    rows={3}
                    value={newQuestionForm.text}
                    onChange={(event) => handleNewQuestionFieldChange("text", event.target.value)}
                    placeholder="Например: Какие виды деятельности вызывают у вас наибольший интерес?"
                  />
                  {newQuestionErrors.text ? (
                    <small className="admin-form-error admin-form-error--inline">{newQuestionErrors.text}</small>
                  ) : null}
                </label>

                <label className="admin-form-field">
                  <span>Тип вопроса</span>
                  <select
                    aria-invalid={Boolean(newQuestionErrors.question_type)}
                    className={
                      newQuestionErrors.question_type
                        ? "admin-form-control admin-form-control--invalid"
                        : "admin-form-control"
                    }
                    value={newQuestionForm.question_type}
                    onChange={(event) => handleNewQuestionFieldChange("question_type", event.target.value)}
                  >
                    {QUESTION_TYPE_ITEMS.map((item) => (
                      <option key={item.value} value={item.value}>
                        {item.label}
                      </option>
                    ))}
                  </select>
                </label>

                <label className="builder-toggle">
                  <span>Обязательный вопрос</span>
                  <input
                    checked={newQuestionForm.is_required}
                    type="checkbox"
                    onChange={(event) => handleNewQuestionFieldChange("is_required", event.target.checked)}
                  />
                </label>

                {needsOptions(newQuestionForm.question_type) ? (
                  <div className="admin-form-field admin-form-field--wide">
                    <span>Варианты ответа</span>
                    <div className="builder-options">
                      <div className="builder-option-row builder-option-row--head" aria-hidden="true">
                        <span>Подпись</span>
                        <span>Value</span>
                        <span>Балл</span>
                        <span />
                      </div>
                      {newQuestionForm.options.map((option, index) => (
                        <div className="builder-option-row" key={`${newQuestionForm.question_type}-${index + 1}`}>
                          <input
                            className="admin-form-control"
                            value={option.label}
                            onChange={(event) => handleNewQuestionOptionChange(index, "label", event.target.value)}
                            placeholder={`Вариант ${index + 1}`}
                          />
                          <input
                            className="admin-form-control"
                            value={option.value}
                            onChange={(event) => handleNewQuestionOptionChange(index, "value", event.target.value)}
                            placeholder={`option_${index + 1}`}
                          />
                          <input
                            className="admin-form-control"
                            inputMode="decimal"
                            type="number"
                            value={option.score}
                            onChange={(event) => handleNewQuestionOptionChange(index, "score", event.target.value)}
                            placeholder={newQuestionForm.question_type === "scale" ? "1" : "опц."}
                          />
                          <button
                            className="table-action-button"
                            disabled={newQuestionForm.options.length <= 2}
                            type="button"
                            onClick={() => removeNewQuestionOption(index)}
                          >
                            <Trash2 size={15} strokeWidth={2.1} />
                            <span>Убрать</span>
                          </button>
                        </div>
                      ))}
                    </div>
                    <div className="builder-option-actions">
                      <button className="table-action-button" type="button" onClick={addNewQuestionOption}>
                        <Plus size={15} strokeWidth={2.1} />
                        <span>Добавить вариант</span>
                      </button>
                    </div>
                    {newQuestionErrors.options ? (
                      <small className="admin-form-error admin-form-error--inline">{newQuestionErrors.options}</small>
                    ) : null}
                  </div>
                ) : null}

                <div className="admin-form-actions">
                  <button className="admin-primary-button" disabled={isCreatingQuestion} type="submit">
                    {isCreatingQuestion ? (
                      <LoaderCircle className="icon-spin" size={16} strokeWidth={2.1} />
                    ) : (
                      <Plus size={16} strokeWidth={2.1} />
                    )}
                    <span>{isCreatingQuestion ? "Добавление..." : "Добавить вопрос"}</span>
                  </button>
                </div>
              </form>
            </section>

            <section className="builder-panel">
              <div className="builder-section-head">
                <div>
                  <p className="builder-section-head__eyebrow">Структура теста</p>
                  <h3 className="builder-section-head__title">Порядок вопросов</h3>
                </div>
                <span className="builder-section-head__meta">Перетаскивай карточки мышью, чтобы менять порядок.</span>
              </div>

              {questionsQuery.isLoading ? (
                <div className="builder-empty">Загружаем структуру методики...</div>
              ) : draftQuestions.length ? (
                <DndContext collisionDetection={closestCenter} onDragEnd={handleQuestionDragEnd} sensors={dndSensors}>
                  <SortableContext items={draftQuestions.map((question) => String(question.id))} strategy={verticalListSortingStrategy}>
                    <div className="builder-question-list">
                      {draftQuestions.map((question, index) => (
                        <SortableQuestionCard
                          key={question.id}
                          index={index}
                          isBusy={activeQuestionId === question.id}
                          isReordering={isReordering}
                          onAddOption={addOptionToQuestion}
                          onDelete={handleDeleteQuestion}
                          onQuestionFieldChange={handleQuestionFieldChange}
                          onQuestionOptionChange={handleQuestionOptionChange}
                          onRemoveOption={removeOptionFromQuestion}
                          onSave={handleSaveQuestion}
                          question={question}
                          questionErrors={questionErrorsById[question.id] || {}}
                        />
                      ))}
                    </div>
                  </SortableContext>
                </DndContext>
              ) : (
                <div className="builder-empty">В тесте пока нет вопросов. Добавь первый блок выше.</div>
              )}
            </section>

            <section className="builder-panel">
              <div className="builder-section-head">
                <div>
                  <p className="builder-section-head__eyebrow">Интерпретация результатов</p>
                  <h3 className="builder-section-head__title">Метрики и правила расчёта</h3>
                </div>
              </div>
              <p className="builder-panel__description">
                Для каждого ответа можно увеличивать или уменьшать любую метрику теста: интерес, лидерство, стрессоустойчивость и другие.
              </p>

              <datalist id="builder-metric-key-options">
                {metricKeySuggestions.map((key) => (
                  <option key={`metric-option-${key}`} value={key} />
                ))}
              </datalist>

              <form className="admin-form-grid" onSubmit={handleCreateFormula}>
                <label className="admin-form-field">
                  <span>Название правила</span>
                  <input
                    aria-invalid={Boolean(newFormulaErrors.name)}
                    className={newFormulaErrors.name ? "admin-form-control admin-form-control--invalid" : "admin-form-control"}
                    value={newFormulaForm.name}
                    onChange={(event) => handleNewFormulaFieldChange("name", event.target.value)}
                    placeholder="Например: Аналитический уклон"
                  />
                  {newFormulaErrors.name ? <small className="admin-form-error admin-form-error--inline">{newFormulaErrors.name}</small> : null}
                </label>

                <label className="admin-form-field">
                  <span>Вопрос</span>
                  <select className="admin-form-control" value={newFormulaForm.question_id} onChange={(event) => handleNewFormulaFieldChange("question_id", event.target.value)}>
                    <option value="0">Глобальное правило</option>
                    {draftQuestions.map((question, index) => (
                      <option key={`formula-new-question-${question.id}`} value={question.id}>
                        #{index + 1} {question.text.slice(0, 60)}
                      </option>
                    ))}
                  </select>
                </label>

                <label className="admin-form-field">
                  <span>Условие</span>
                  <select
                    aria-invalid={Boolean(newFormulaErrors.condition_type)}
                    className={newFormulaErrors.condition_type ? "admin-form-control admin-form-control--invalid" : "admin-form-control"}
                    value={newFormulaForm.condition_type}
                    onChange={(event) => handleNewFormulaFieldChange("condition_type", event.target.value)}
                  >
                    {CONDITION_TYPE_ITEMS.map((item) => (
                      <option key={`formula-condition-${item.value}`} value={item.value}>
                        {item.label}
                      </option>
                    ))}
                  </select>
                  {newFormulaErrors.condition_type ? <small className="admin-form-error admin-form-error--inline">{newFormulaErrors.condition_type}</small> : null}
                </label>

                <label className="admin-form-field">
                  <span>Ожидаемое значение</span>
                  <input
                    aria-invalid={Boolean(newFormulaErrors.expected_value)}
                    className={newFormulaErrors.expected_value ? "admin-form-control admin-form-control--invalid" : "admin-form-control"}
                    disabled={newFormulaForm.condition_type === "always"}
                    value={newFormulaForm.expected_value}
                    onChange={(event) => handleNewFormulaFieldChange("expected_value", event.target.value)}
                    placeholder={newFormulaForm.condition_type === "answer_in" ? "math,science,logic" : "Например: 5"}
                  />
                  {newFormulaErrors.expected_value ? <small className="admin-form-error admin-form-error--inline">{newFormulaErrors.expected_value}</small> : null}
                </label>

                <label className="admin-form-field">
                  <span>Ключ метрики</span>
                  <input
                    className="admin-form-control"
                    list="builder-metric-key-options"
                    value={newFormulaForm.result_key}
                    onChange={(event) => handleNewFormulaFieldChange("result_key", event.target.value)}
                    placeholder="stress_resistance / leadership / total"
                  />
                </label>

                <label className="admin-form-field">
                  <span>Изменение метрики</span>
                  <input
                    aria-invalid={Boolean(newFormulaErrors.score_delta)}
                    className={newFormulaErrors.score_delta ? "admin-form-control admin-form-control--invalid" : "admin-form-control"}
                    inputMode="decimal"
                    type="number"
                    value={newFormulaForm.score_delta}
                    onChange={(event) => handleNewFormulaFieldChange("score_delta", event.target.value)}
                    placeholder="+1 / -2 / 3.5"
                  />
                  {newFormulaErrors.score_delta ? <small className="admin-form-error admin-form-error--inline">{newFormulaErrors.score_delta}</small> : null}
                </label>

                <label className="admin-form-field">
                  <span>Priority</span>
                  <input
                    aria-invalid={Boolean(newFormulaErrors.priority)}
                    className={newFormulaErrors.priority ? "admin-form-control admin-form-control--invalid" : "admin-form-control"}
                    inputMode="numeric"
                    type="number"
                    value={newFormulaForm.priority}
                    onChange={(event) => handleNewFormulaFieldChange("priority", event.target.value)}
                    placeholder="0"
                  />
                  {newFormulaErrors.priority ? <small className="admin-form-error admin-form-error--inline">{newFormulaErrors.priority}</small> : null}
                </label>

                <div className="builder-metric-helper admin-form-field--wide">
                  <span className="builder-metric-helper__label">Быстрые ключи метрик</span>
                  <div className="builder-metric-chip-list">
                    {metricKeySuggestions.map((key) => (
                      <button
                        key={`metric-chip-${key}`}
                        className="builder-metric-chip"
                        type="button"
                        onClick={() => handleNewFormulaFieldChange("result_key", key)}
                      >
                        <strong>{getMetricLabel(key)}</strong>
                        <span>{key}</span>
                      </button>
                    ))}
                  </div>
                </div>

                <div className="admin-form-actions">
                  <button className="admin-primary-button" disabled={isCreatingFormula} type="submit">
                    {isCreatingFormula ? <LoaderCircle className="icon-spin" size={16} strokeWidth={2.1} /> : <Plus size={16} strokeWidth={2.1} />}
                    <span>{isCreatingFormula ? "Добавление..." : "Добавить правило"}</span>
                  </button>
                </div>
              </form>

              <div className="builder-formula-list">
                {formulaRulesQuery.isLoading && !draftFormulaRules.length ? (
                  <div className="builder-empty">Загружаем правила расчёта...</div>
                ) : draftFormulaRules.length ? (
                  draftFormulaRules.map((rule) => {
                    const formulaErrors = formulaErrorsById[rule.id] || {};
                    const isBusy = activeFormulaId === rule.id;

                    return (
                      <article className="builder-formula-card" key={rule.id}>
                        <div className="builder-formula-card__head">
                          <strong>{rule.name || "Правило без названия"}</strong>
                          <div className="admin-table__actions">
                            <button className="table-action-button" disabled={isBusy} type="button" onClick={() => handleSaveFormula(rule.id)}>
                              {isBusy ? <LoaderCircle className="icon-spin" size={15} strokeWidth={2.1} /> : <Save size={15} strokeWidth={2.1} />}
                              <span>Сохранить</span>
                            </button>
                            <button className="table-action-button" disabled={isBusy} type="button" onClick={() => handleDeleteFormula(rule.id)}>
                              <Trash2 size={15} strokeWidth={2.1} />
                              <span>Удалить</span>
                            </button>
                          </div>
                        </div>

                        <div className="builder-formula-grid">
                          <label className="admin-form-field">
                            <span>Название</span>
                            <input
                              aria-invalid={Boolean(formulaErrors.name)}
                              className={formulaErrors.name ? "admin-form-control admin-form-control--invalid" : "admin-form-control"}
                              value={rule.name}
                              onChange={(event) => handleFormulaFieldChange(rule.id, "name", event.target.value)}
                            />
                            {formulaErrors.name ? <small className="admin-form-error admin-form-error--inline">{formulaErrors.name}</small> : null}
                          </label>

                          <label className="admin-form-field">
                            <span>Вопрос</span>
                            <select className="admin-form-control" value={rule.question_id} onChange={(event) => handleFormulaFieldChange(rule.id, "question_id", event.target.value)}>
                              <option value="0">Глобальное правило</option>
                              {draftQuestions.map((question, index) => (
                                <option key={`formula-question-${rule.id}-${question.id}`} value={question.id}>
                                  #{index + 1} {question.text.slice(0, 60)}
                                </option>
                              ))}
                            </select>
                          </label>

                          <label className="admin-form-field">
                            <span>Условие</span>
                            <select
                              aria-invalid={Boolean(formulaErrors.condition_type)}
                              className={formulaErrors.condition_type ? "admin-form-control admin-form-control--invalid" : "admin-form-control"}
                              value={rule.condition_type}
                              onChange={(event) => handleFormulaFieldChange(rule.id, "condition_type", event.target.value)}
                            >
                              {CONDITION_TYPE_ITEMS.map((item) => (
                                <option key={`formula-condition-${rule.id}-${item.value}`} value={item.value}>
                                  {item.label}
                                </option>
                              ))}
                            </select>
                            {formulaErrors.condition_type ? <small className="admin-form-error admin-form-error--inline">{formulaErrors.condition_type}</small> : null}
                          </label>

                          <label className="admin-form-field">
                            <span>Ожидаемое значение</span>
                            <input
                              aria-invalid={Boolean(formulaErrors.expected_value)}
                              className={formulaErrors.expected_value ? "admin-form-control admin-form-control--invalid" : "admin-form-control"}
                              disabled={rule.condition_type === "always"}
                              value={rule.expected_value}
                              onChange={(event) => handleFormulaFieldChange(rule.id, "expected_value", event.target.value)}
                            />
                            {formulaErrors.expected_value ? <small className="admin-form-error admin-form-error--inline">{formulaErrors.expected_value}</small> : null}
                          </label>

                          <label className="admin-form-field">
                            <span>Ключ метрики</span>
                            <input
                              className="admin-form-control"
                              list="builder-metric-key-options"
                              value={rule.result_key}
                              onChange={(event) => handleFormulaFieldChange(rule.id, "result_key", event.target.value)}
                            />
                          </label>

                          <label className="admin-form-field">
                            <span>Изменение метрики</span>
                            <input
                              aria-invalid={Boolean(formulaErrors.score_delta)}
                              className={formulaErrors.score_delta ? "admin-form-control admin-form-control--invalid" : "admin-form-control"}
                              inputMode="decimal"
                              type="number"
                              value={rule.score_delta}
                              onChange={(event) => handleFormulaFieldChange(rule.id, "score_delta", event.target.value)}
                            />
                            {formulaErrors.score_delta ? <small className="admin-form-error admin-form-error--inline">{formulaErrors.score_delta}</small> : null}
                          </label>

                          <label className="admin-form-field">
                            <span>Priority</span>
                            <input
                              aria-invalid={Boolean(formulaErrors.priority)}
                              className={formulaErrors.priority ? "admin-form-control admin-form-control--invalid" : "admin-form-control"}
                              inputMode="numeric"
                              type="number"
                              value={rule.priority}
                              onChange={(event) => handleFormulaFieldChange(rule.id, "priority", event.target.value)}
                            />
                            {formulaErrors.priority ? <small className="admin-form-error admin-form-error--inline">{formulaErrors.priority}</small> : null}
                          </label>
                        </div>
                      </article>
                    );
                  })
                ) : (
                  <div className="builder-empty">Правила расчёта ещё не добавлены. Создай первое правило выше.</div>
                )}
              </div>
            </section>

            <section className="builder-panel">
              <div className="builder-section-head">
                <div>
                  <p className="builder-section-head__eyebrow">Preview</p>
                  <h3 className="builder-section-head__title">Проверка метрик</h3>
                </div>
                <button className="table-action-button" disabled={isPreviewingFormula} type="button" onClick={handleCalculateFormulaPreview}>
                  {isPreviewingFormula ? <LoaderCircle className="icon-spin" size={15} strokeWidth={2.1} /> : <Calculator size={15} strokeWidth={2.1} />}
                  <span>{isPreviewingFormula ? "Расчёт..." : "Посчитать preview"}</span>
                </button>
              </div>

              {previewError ? <p className="admin-form-message admin-form-message--error">{previewError}</p> : null}

              <div className="builder-preview-grid">
                <div className="builder-preview-inputs">
                  {draftQuestions.length ? (
                    draftQuestions.map((question, index) => {
                      const previewValue = previewAnswersByQuestion[String(question.id)] || {};

                      return (
                        <div className="builder-preview-card" key={`preview-question-${question.id}`}>
                          <strong>#{index + 1} {question.text}</strong>

                          {question.question_type === "multiple_choice" ? (
                            <div className="client-answer-options">
                              {(question.options || []).map((option) => (
                                <label className="client-answer-option" key={`preview-option-${question.id}-${option.value}`}>
                                  <input
                                    checked={Array.isArray(previewValue.answer_values) && previewValue.answer_values.includes(option.value)}
                                    type="checkbox"
                                    onChange={(event) => {
                                      const current = Array.isArray(previewValue.answer_values) ? previewValue.answer_values : [];
                                      const nextValues = event.target.checked
                                        ? [...current, option.value]
                                        : current.filter((item) => item !== option.value);
                                      handlePreviewAnswerChange(question.id, { answer_values: nextValues });
                                    }}
                                  />
                                  <span>{option.label}</span>
                                </label>
                              ))}
                            </div>
                          ) : question.question_type === "text" ? (
                            <input
                              className="admin-form-control"
                              value={previewValue.answer_value || ""}
                              onChange={(event) => handlePreviewAnswerChange(question.id, { answer_value: event.target.value })}
                              placeholder="Текст ответа"
                            />
                          ) : needsOptions(question.question_type) ? (
                            <select
                              className="admin-form-control"
                              value={previewValue.answer_value || ""}
                              onChange={(event) => handlePreviewAnswerChange(question.id, { answer_value: event.target.value })}
                            >
                              <option value="">Не выбрано</option>
                              {(question.options || []).map((option) => (
                                <option key={`preview-select-${question.id}-${option.value}`} value={option.value}>
                                  {option.label}
                                </option>
                              ))}
                            </select>
                          ) : (
                            <input
                              className="admin-form-control"
                              inputMode={question.question_type === "number" ? "numeric" : "text"}
                              type={question.question_type === "number" ? "number" : "text"}
                              value={previewValue.answer_value || ""}
                              onChange={(event) => handlePreviewAnswerChange(question.id, { answer_value: event.target.value })}
                              placeholder={question.question_type === "number" ? "Число" : "Ответ"}
                            />
                          )}
                        </div>
                      );
                    })
                  ) : (
                    <div className="builder-empty">Добавьте вопросы, чтобы тестировать формулы на примерах ответов.</div>
                  )}
                </div>

                <div className="builder-preview-result">
                  <div className="builder-stat-card">
                    <span>Суммарный счёт</span>
                    <strong>{previewResult?.total_score ?? "—"}</strong>
                  </div>

                  <div className="builder-health">
                    <p className="builder-section-head__eyebrow">Сработавшие правила</p>
                    {previewResult?.triggered_rule_ids?.length ? (
                      <div className="builder-trigger-list">
                        {previewResult.triggered_rule_ids.map((ruleId) => (
                          <span className="status-badge status-badge--active" key={`triggered-rule-${ruleId}`}>
                            #{ruleId}
                          </span>
                        ))}
                      </div>
                    ) : (
                      <p className="builder-health__item">Пока нет сработавших правил.</p>
                    )}
                  </div>

                  <div className="builder-health">
                    <p className="builder-section-head__eyebrow">Метрики</p>
                    {previewResult?.metrics && Object.keys(previewResult.metrics).length ? (
                      <div className="builder-metrics-list">
                        {Object.entries(previewResult.metrics).map(([key, value]) => (
                          <div className="builder-metric-row" key={`metric-${key}`}>
                            <span>{getMetricLabel(key)}</span>
                            <strong>{value}</strong>
                          </div>
                        ))}
                      </div>
                    ) : (
                      <p className="builder-health__item">Метрики появятся после первого preview-расчёта.</p>
                    )}
                  </div>
                </div>
              </div>
            </section>
          </div>

          <aside className="builder-sidebar">
            <section className="builder-panel builder-panel--sticky">
              <div className="builder-section-head">
                <div>
                  <p className="builder-section-head__eyebrow">Сводка</p>
                  <h3 className="builder-section-head__title">Текущее состояние</h3>
                </div>
              </div>

              <div className="builder-stats">
                {sidebarStats.map((item) => (
                  <div className="builder-stat-card" key={item.label}>
                    <span>{item.label}</span>
                    <strong>{item.value}</strong>
                  </div>
                ))}
              </div>

              <div className="builder-health">
                <p className="builder-section-head__eyebrow">Проверка перед публикацией</p>
                {publishIssues.length ? (
                  <ul className="builder-health__list">
                    {publishIssues.map((issue) => (
                      <li className="builder-health__item" key={issue}>
                        {issue}
                      </li>
                    ))}
                  </ul>
                ) : (
                  <p className="builder-health__item builder-health__item--ok">Методика готова к публикации.</p>
                )}
              </div>

              <div className="builder-notes">
                <div className="builder-note">
                  <LayoutTemplate size={16} strokeWidth={2.1} />
                  <p>Сначала сохраняй структуру и параметры теста, потом публикуй и выдавай ссылку клиенту.</p>
                </div>
                <div className="builder-note">
                  <Link2 size={16} strokeWidth={2.1} />
                  <p>Для вопросов со шкалой и вариантами лучше сразу заполнять value и score, чтобы потом проще строить отчёты.</p>
                </div>
              </div>

              <div className="builder-public-link">
                <span>Публичная ссылка</span>
                <strong>{publicUrl || "Появится после публикации"}</strong>
              </div>

              <div className="admin-table__actions">
                <Link className="table-action-link" to={ROUTES.tests}>
                  <LayoutTemplate size={15} strokeWidth={2.1} />
                  <span>Все опросники</span>
                </Link>
                <Link className="table-action-link" to={toResultsPath(id)}>
                  <Eye size={15} strokeWidth={2.1} />
                  <span>Результаты</span>
                </Link>
              </div>
            </section>

            <section className="builder-panel">
              <div className="builder-section-head">
                <div>
                  <p className="builder-section-head__eyebrow">Отчётность</p>
                  <h3 className="builder-section-head__title">Шаблоны отчётов</h3>
                </div>
              </div>
              <p className="builder-panel__description">
                Шаблон задаёт структуру клиентского и технического отчёта. В конструкторе выбирается только привязка к тесту, а создание и редактирование вынесены в отдельную страницу.
              </p>

              {reportTemplatesQuery.error ? (
                <p className="admin-form-message admin-form-message--error">
                  {reportTemplatesQuery.error.message || "Не удалось загрузить шаблоны отчётов."}
                </p>
              ) : null}

              <div className="builder-formula-list">
                {reportTemplatesQuery.isLoading ? (
                  <div className="builder-empty">Загружаем шаблоны отчётов...</div>
                ) : reportTemplates.length ? (
                  <>
                    <div className="builder-note">
                      <Link2 size={16} strokeWidth={2.1} />
                      <p>
                        {selectedReportTemplate
                          ? `Сейчас к тесту привязан шаблон «${selectedReportTemplate.name}».`
                          : "Шаблон к тесту пока не привязан. Выберите его в параметрах выше."}
                      </p>
                    </div>
                    <div className="admin-table__actions">
                      <Link className="table-action-link" to={ROUTES.reportTemplates}>
                        <LayoutTemplate size={15} strokeWidth={2.1} />
                        <span>Открыть страницу шаблонов</span>
                      </Link>
                    </div>
                  </>
                ) : (
                  <>
                    <div className="builder-empty">Шаблонов отчётов пока нет. Сначала создайте шаблон на отдельной странице.</div>
                    <div className="admin-table__actions">
                      <Link className="table-action-link" to={ROUTES.reportTemplates}>
                        <LayoutTemplate size={15} strokeWidth={2.1} />
                        <span>Создать шаблон отчёта</span>
                      </Link>
                    </div>
                  </>
                )}
              </div>
            </section>
          </aside>
        </section>
      ) : null}
    </PageCard>
  );
}
