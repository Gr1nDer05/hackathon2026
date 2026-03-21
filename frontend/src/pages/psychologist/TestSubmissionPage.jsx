import { Eye, FileDown, FileText, LayoutTemplate, LoaderCircle } from "lucide-react";
import { useMemo, useState } from "react";
import useSWR from "swr";
import { Link, useParams } from "react-router-dom";
import {
  formatDate,
  getResultProfessionEntries,
  normalizeResultItem,
} from "../../modules/psychologist/lib/psychologistUi";
import {
  getPsychologistSubmissionReportRequest,
  getPsychologistTestRequest,
  getPsychologistTestSubmissionRequest,
  listPsychologistReportTemplatesRequest,
  listPsychologistQuestionsRequest,
} from "../../modules/tests/api/testsApi";
import { buildClientSessionPath, ROUTES } from "../../shared/config/routes";
import PageCard from "../../shared/ui/PageCard";

function toBuilderPath(id) {
  return ROUTES.testBuilder.replace(":id", String(id));
}

function toResultsPath(id) {
  return ROUTES.testResults.replace(":id", String(id));
}

function getQuestionTypeLabel(questionType) {
  const labels = {
    single_choice: "Один вариант",
    multiple_choice: "Несколько вариантов",
    scale: "Шкала",
    text: "Текст",
    number: "Число",
  };

  return labels[questionType] || questionType || "—";
}

function getStatusLabel(status) {
  if (["completed", "submitted", "passed"].includes(status)) {
    return "Завершён";
  }

  if (["in_progress", "started"].includes(status)) {
    return "В процессе";
  }

  return status || "Неизвестно";
}

function getAnswerDisplay(answer, question) {
  const options = Array.isArray(question?.options) ? question.options : [];
  const optionLabelByValue = Object.fromEntries(
    options.map((option) => [String(option.value), option.label || option.value]),
  );

  if (Array.isArray(answer?.answer_values) && answer.answer_values.length) {
    return answer.answer_values
      .map((value) => optionLabelByValue[String(value)] || String(value))
      .join(", ");
  }

  if (String(answer?.answer_text || "").trim()) {
    return String(answer.answer_text).trim();
  }

  if (answer?.answer_value !== null && answer?.answer_value !== undefined && String(answer.answer_value).trim()) {
    return optionLabelByValue[String(answer.answer_value)] || String(answer.answer_value);
  }

  return "—";
}

export default function TestSubmissionPage() {
  const { id, sessionId } = useParams();
  const [reportError, setReportError] = useState("");
  const [reportLoadingKey, setReportLoadingKey] = useState("");

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
  const submissionQuery = useSWR(
    id && sessionId ? ["psychologist-test-submission", id, sessionId] : null,
    () => getPsychologistTestSubmissionRequest(id, sessionId),
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

  const submission = useMemo(
    () => (submissionQuery.data ? normalizeResultItem(submissionQuery.data, 0) : null),
    [submissionQuery.data],
  );
  const assignedTemplate = useMemo(() => {
    const templates = Array.isArray(reportTemplatesQuery.data) ? reportTemplatesQuery.data : [];
    const templateId = testQuery.data?.report_template_id;

    return templates.find((template) => String(template.id) === String(templateId || "")) || null;
  }, [reportTemplatesQuery.data, testQuery.data?.report_template_id]);
  const professions = useMemo(
    () => getResultProfessionEntries(submissionQuery.data || {}),
    [submissionQuery.data],
  );
  const questionById = useMemo(() => {
    const questions = Array.isArray(questionsQuery.data) ? questionsQuery.data : [];

    return Object.fromEntries(
      questions
        .filter((question) => question?.id !== null && question?.id !== undefined)
        .map((question) => [String(question.id), question]),
    );
  }, [questionsQuery.data]);
  const answers = Array.isArray(submissionQuery.data?.answers) ? submissionQuery.data.answers : [];
  const publicPath = testQuery.data?.public_slug ? buildClientSessionPath(testQuery.data.public_slug) : "";
  const canGenerateReports = ["completed", "submitted", "passed"].includes(submission?.status);

  function buildReportLoadingKey(audience, format) {
    return `${sessionId}:${audience}:${format}`;
  }

  async function handleOpenReport(audience, format) {
    const loadingKey = buildReportLoadingKey(audience, format);
    const reportTab =
      format === "html"
        ? window.open("about:blank", "_blank", "noopener,noreferrer")
        : null;

    setReportLoadingKey(loadingKey);
    setReportError("");

    try {
      const file = await getPsychologistSubmissionReportRequest(sessionId, { audience, format });
      const objectUrl = window.URL.createObjectURL(file.blob);

      if (format === "html") {
        if (!reportTab) {
          window.location.assign(objectUrl);
          return;
        }

        reportTab.location.href = objectUrl;
        window.setTimeout(() => window.URL.revokeObjectURL(objectUrl), 60_000);
      } else {
        const link = document.createElement("a");
        link.href = objectUrl;
        link.download = file.filename;
        document.body.append(link);
        link.click();
        link.remove();
        window.setTimeout(() => window.URL.revokeObjectURL(objectUrl), 1_000);
      }
    } catch (error) {
      if (reportTab) {
        reportTab.close();
      }
      if (error?.status === 409) {
        const likelyReason = !canGenerateReports
          ? "сессия ещё не завершена"
          : !testQuery.data?.report_template_id
            ? "у теста не выбран шаблон отчёта"
            : audience === "client" && !testQuery.data?.show_client_report_immediately
              ? "в тесте выключен мгновенный клиентский отчёт"
              : "backend отклонил генерацию в текущем состоянии";
        setReportError(
          `Отчёт не сформирован: ${likelyReason}. Проверь завершение сессии, шаблон отчёта и настройки теста.`,
        );
      } else {
        setReportError(error?.message || "Не удалось сформировать отчёт.");
      }
    } finally {
      setReportLoadingKey("");
    }
  }

  return (
    <PageCard
      wide
      title={testQuery.data?.title ? `Прохождение: ${testQuery.data.title}` : `Прохождение #${sessionId}`}
      description="Детали одной сессии: респондент, ответы, метрики и генерация отчётов."
      links={[
        { to: toResultsPath(id), label: "К списку результатов" },
        { to: toBuilderPath(id), label: "В конструктор" },
        { to: ROUTES.reportTemplates, label: "Шаблоны отчётов" },
      ]}
    >
      {testQuery.error ? <p className="admin-form-message admin-form-message--error">{testQuery.error.message || "Не удалось загрузить тест."}</p> : null}
      {questionsQuery.error ? <p className="admin-form-message admin-form-message--error">{questionsQuery.error.message || "Не удалось загрузить вопросы теста."}</p> : null}
      {submissionQuery.error ? <p className="admin-form-message admin-form-message--error">{submissionQuery.error.message || "Не удалось загрузить прохождение."}</p> : null}
      {reportTemplatesQuery.error ? <p className="admin-form-message admin-form-message--error">{reportTemplatesQuery.error.message || "Не удалось загрузить шаблоны отчётов."}</p> : null}
      {reportError ? <p className="admin-form-message admin-form-message--error">{reportError}</p> : null}

      {submissionQuery.isLoading ? (
        <div className="builder-empty">Загружаем данные прохождения...</div>
      ) : submission ? (
        <>
          <div className={`workflow-note ${canGenerateReports && testQuery.data?.report_template_id ? "workflow-note--success" : "workflow-note--warning"}`}>
            <p>
              {testQuery.data?.report_template_id
                ? `Для этой сессии используется шаблон «${assignedTemplate?.name || `#${testQuery.data.report_template_id}` }».`
                : "У теста пока нет привязанного шаблона отчёта."}{" "}
              {canGenerateReports
                ? "Сессия завершена, можно формировать клиентский и технический отчёт."
                : "Сессия ещё не завершена, поэтому генерация итоговых отчётов недоступна."}
            </p>
            <div className="workflow-note__actions">
              <Link className="table-action-link" to={ROUTES.reportTemplates}>
                <LayoutTemplate size={15} strokeWidth={2.1} />
                <span>Проверить шаблоны</span>
              </Link>
            </div>
          </div>

          <section className="psychologist-kpis">
            <article className="psychologist-kpi">
              <p className="psychologist-kpi__label">Статус</p>
              <p className="psychologist-kpi__value psychologist-kpi__value--small">{getStatusLabel(submission.status)}</p>
            </article>
            <article className="psychologist-kpi">
              <p className="psychologist-kpi__label">Прогресс</p>
              <p className="psychologist-kpi__value">{submission.progress}%</p>
            </article>
            <article className="psychologist-kpi">
              <p className="psychologist-kpi__label">Ответов</p>
              <p className="psychologist-kpi__value">{answers.length}</p>
            </article>
            <article className="psychologist-kpi">
              <p className="psychologist-kpi__label">Метрик</p>
              <p className="psychologist-kpi__value">{submission.metrics.length}</p>
            </article>
          </section>

          <section className="admin-panels">
            <article className="admin-panel">
              <h3 className="admin-panel__title">Респондент</h3>
              <dl className="profile-meta-list">
                <div>
                  <dt>ФИО</dt>
                  <dd>{submissionQuery.data?.respondent_name || submission.respondent}</dd>
                </div>
                <div>
                  <dt>Телефон</dt>
                  <dd>{submissionQuery.data?.respondent_phone || "—"}</dd>
                </div>
                <div>
                  <dt>Email</dt>
                  <dd>{submissionQuery.data?.respondent_email || "—"}</dd>
                </div>
                <div>
                  <dt>Возраст</dt>
                  <dd>{submissionQuery.data?.respondent_age || "—"}</dd>
                </div>
                <div>
                  <dt>Пол</dt>
                  <dd>{submissionQuery.data?.respondent_gender || "—"}</dd>
                </div>
                <div>
                  <dt>Образование</dt>
                  <dd>{submissionQuery.data?.respondent_education || "—"}</dd>
                </div>
                <div>
                  <dt>Старт</dt>
                  <dd>{formatDate(submission.startedAt)}</dd>
                </div>
                <div>
                  <dt>Завершение</dt>
                  <dd>{formatDate(submission.completedAt)}</dd>
                </div>
                <div>
                  <dt>Шаблон</dt>
                  <dd>{assignedTemplate?.name || (testQuery.data?.report_template_id ? `Шаблон #${testQuery.data.report_template_id}` : "Не выбран")}</dd>
                </div>
              </dl>
            </article>

            <article className="admin-panel">
              <h3 className="admin-panel__title">Отчёты</h3>
              <p className="admin-panel__meta">Генерация итоговых файлов для клиента и специалиста по текущей сессии.</p>
              <p className="admin-panel__meta">
                {testQuery.data?.show_client_report_immediately
                  ? "Мгновенный клиентский отчёт включён в настройках теста."
                  : "Мгновенный клиентский отчёт сейчас выключен в настройках теста."}
              </p>
              <div className="admin-table__actions admin-table__actions--reports">
                <button
                  className="table-action-button"
                  disabled={!canGenerateReports || reportLoadingKey === buildReportLoadingKey("psychologist", "html")}
                  type="button"
                  onClick={() => handleOpenReport("psychologist", "html")}
                >
                  {reportLoadingKey === buildReportLoadingKey("psychologist", "html") ? (
                    <LoaderCircle className="icon-spin" size={15} strokeWidth={2.1} />
                  ) : (
                    <FileText size={15} strokeWidth={2.1} />
                  )}
                  <span>HTML спец.</span>
                </button>
                <button
                  className="table-action-button"
                  disabled={!canGenerateReports || reportLoadingKey === buildReportLoadingKey("psychologist", "docx")}
                  type="button"
                  onClick={() => handleOpenReport("psychologist", "docx")}
                >
                  {reportLoadingKey === buildReportLoadingKey("psychologist", "docx") ? (
                    <LoaderCircle className="icon-spin" size={15} strokeWidth={2.1} />
                  ) : (
                    <FileDown size={15} strokeWidth={2.1} />
                  )}
                  <span>DOCX спец.</span>
                </button>
                <button
                  className="table-action-button"
                  disabled={!canGenerateReports || reportLoadingKey === buildReportLoadingKey("client", "html")}
                  type="button"
                  onClick={() => handleOpenReport("client", "html")}
                >
                  {reportLoadingKey === buildReportLoadingKey("client", "html") ? (
                    <LoaderCircle className="icon-spin" size={15} strokeWidth={2.1} />
                  ) : (
                    <FileText size={15} strokeWidth={2.1} />
                  )}
                  <span>HTML клиент</span>
                </button>
                <button
                  className="table-action-button"
                  disabled={!canGenerateReports || reportLoadingKey === buildReportLoadingKey("client", "docx")}
                  type="button"
                  onClick={() => handleOpenReport("client", "docx")}
                >
                  {reportLoadingKey === buildReportLoadingKey("client", "docx") ? (
                    <LoaderCircle className="icon-spin" size={15} strokeWidth={2.1} />
                  ) : (
                    <FileDown size={15} strokeWidth={2.1} />
                  )}
                  <span>DOCX клиент</span>
                </button>
              </div>
              {!canGenerateReports ? (
                <p className="admin-table__secondary">Пока сессия не завершена, можно только просмотреть ответы и прогресс.</p>
              ) : null}

              <div className="psychologist-quick-grid psychologist-quick-grid--compact">
                <Link className="psychologist-quick-card" to={toBuilderPath(id)}>
                  <LayoutTemplate size={18} strokeWidth={2.1} />
                  <strong>Открыть конструктор</strong>
                  <span>Проверить структуру теста и привязанный шаблон отчёта.</span>
                </Link>
                {publicPath ? (
                  <Link className="psychologist-quick-card" to={publicPath} target="_blank" rel="noreferrer">
                    <Eye size={18} strokeWidth={2.1} />
                    <strong>Публичная форма</strong>
                    <span>Открыть клиентский сценарий в новой вкладке.</span>
                  </Link>
                ) : null}
              </div>
            </article>
          </section>

          <section className="admin-panels">
            <article className="admin-panel">
              <h3 className="admin-panel__title">Метрики</h3>
              {submission.metrics.length ? (
                <div className="psychologist-summary-grid">
                  {submission.metrics.map((metric) => (
                    <div className="psychologist-summary-card" key={`metric-${metric.key}`}>
                      <span>{metric.label}</span>
                      <strong>{metric.displayValue}</strong>
                      <p>{metric.meta || "Рассчитано backend"}</p>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="psychologist-empty-state">Backend не вернул рассчитанные метрики по этой сессии.</p>
              )}
            </article>

            <article className="admin-panel">
              <h3 className="admin-panel__title">Профессии / выводы</h3>
              {professions.length ? (
                <div className="psychologist-summary-grid">
                  {professions.map((item) => (
                    <div className="psychologist-summary-card" key={`profession-${item.profession}`}>
                      <span>{item.profession}</span>
                      <strong>{item.score}</strong>
                      <p>Ранжирование по рассчитанному результату.</p>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="psychologist-empty-state">Дополнительные выводы по профессиям для этой сессии не рассчитаны.</p>
              )}
            </article>
          </section>

          <div className="admin-table-wrap">
            <table className="admin-table">
              <thead>
                <tr>
                  <th>Вопрос</th>
                  <th>Тип</th>
                  <th>Ответ</th>
                </tr>
              </thead>
              <tbody>
                {answers.length ? (
                  answers.map((answer) => {
                    const question = questionById[String(answer.question_id)];

                    return (
                      <tr key={`answer-${answer.id || `${answer.question_id}-${answer.answer_value || answer.answer_text || "empty"}`}`}>
                        <td>
                          <p className="admin-table__primary">{question?.text || `Вопрос #${answer.question_id}`}</p>
                          <p className="admin-table__secondary">ID вопроса: {answer.question_id}</p>
                        </td>
                        <td>{getQuestionTypeLabel(question?.question_type)}</td>
                        <td>{getAnswerDisplay(answer, question)}</td>
                      </tr>
                    );
                  })
                ) : (
                  <tr>
                    <td className="admin-table__empty" colSpan={3}>
                      Backend не вернул ответы по этой сессии.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </>
      ) : (
        <div className="psychologist-empty-state">Данные по этой сессии не найдены. Вернись к списку результатов и открой другую попытку.</div>
      )}
    </PageCard>
  );
}
