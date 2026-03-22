import { Eye, LayoutTemplate, LoaderCircle } from "lucide-react";
import { useMemo } from "react";
import useSWR from "swr";
import { Link, useParams } from "react-router-dom";
import {
  formatDate,
  getStatusTone,
  normalizeResultItem,
} from "../../modules/psychologist/lib/psychologistUi";
import {
  getPsychologistTestRequest,
  listPsychologistReportTemplatesRequest,
  listPsychologistTestResultsRequest,
} from "../../modules/tests/api/testsApi";
import { buildClientSessionPath, buildTestSubmissionPath, ROUTES } from "../../shared/config/routes";
import PageCard from "../../shared/ui/PageCard";

function toBuilderPath(id) {
  return ROUTES.testBuilder.replace(":id", String(id));
}

function getResultStatusLabel(status) {
  if (["completed", "submitted", "passed"].includes(status)) {
    return "Завершён";
  }

  if (["in_progress", "started"].includes(status)) {
    return "В процессе";
  }

  return status || "Неизвестно";
}

export default function TestResultsPage() {
  const { id } = useParams();
  const testQuery = useSWR(id ? ["psychologist-test", id] : null, () => getPsychologistTestRequest(id), {
    revalidateOnFocus: false,
    shouldRetryOnError: false,
  });
  const resultsQuery = useSWR(
    id ? ["psychologist-test-results", id] : null,
    () => listPsychologistTestResultsRequest(id),
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

  const results = useMemo(
    () => (resultsQuery.data || []).map((item, index) => normalizeResultItem(item, index)),
    [resultsQuery.data],
  );
  const assignedTemplate = useMemo(() => {
    const templates = Array.isArray(reportTemplatesQuery.data) ? reportTemplatesQuery.data : [];
    const templateId = testQuery.data?.report_template_id;

    return templates.find((template) => String(template.id) === String(templateId || "")) || null;
  }, [reportTemplatesQuery.data, testQuery.data?.report_template_id]);
  const publicPath = testQuery.data?.public_slug ? buildClientSessionPath(testQuery.data.public_slug) : "";
  const publicUrl = publicPath ? `${window.location.origin}${publicPath}` : "";

  const summary = useMemo(() => {
    const total = results.length;
    const completed = results.filter((item) => ["completed", "submitted", "passed"].includes(item.status)).length;
    const inProgress = results.filter((item) => ["in_progress", "started"].includes(item.status)).length;
    const averageProgress = total
      ? Math.round(results.reduce((sum, item) => sum + item.progress, 0) / total)
      : 0;

    return { total, completed, inProgress, averageProgress };
  }, [results]);

  function isCompletedResult(status) {
    return ["completed", "submitted", "passed"].includes(status);
  }

  return (
    <PageCard
      wide
      title={testQuery.data?.title ? `Результаты: ${testQuery.data.title}` : `Результаты теста: ${id}`}
      description="Попытки прохождения, прогресс респондентов и базовая аналитика по тесту."
      links={[
        { to: ROUTES.tests, label: "К списку тестов" },
        { to: toBuilderPath(id), label: "В конструктор" },
      ]}
    >
      {testQuery.error ? (
        <p className="admin-form-message admin-form-message--error">
          {testQuery.error.message || "Не удалось загрузить параметры теста."}
        </p>
      ) : null}
      {resultsQuery.error ? (
        <p className="admin-form-message admin-form-message--error">
          {resultsQuery.error.message || "Не удалось загрузить результаты прохождений."}
        </p>
      ) : null}
      {reportTemplatesQuery.error ? (
        <p className="admin-form-message admin-form-message--error">
          {reportTemplatesQuery.error.message || "Не удалось загрузить шаблоны отчётов."}
        </p>
      ) : null}

      <div
        className={`workflow-note ${
          testQuery.data?.report_template_id && summary.completed ? "workflow-note--success" : "workflow-note--warning"
        }`}
      >
        <p>
          {testQuery.data?.report_template_id
            ? `Для теста подключён шаблон «${assignedTemplate?.name || `#${testQuery.data.report_template_id}` }».`
            : "Для теста пока не выбран шаблон отчёта."}{" "}
          {summary.completed
            ? "Открой завершённую сессию, чтобы сформировать HTML или DOCX отчёт."
            : "Отчёты станут доступны после первого завершённого прохождения."}
        </p>
        <div className="workflow-note__actions">
          <Link className="table-action-link" to={ROUTES.reportTemplates}>
            <LayoutTemplate size={15} strokeWidth={2.1} />
            <span>Шаблоны отчётов</span>
          </Link>
        </div>
      </div>

      <section className="psychologist-kpis">
        <article className="psychologist-kpi">
          <p className="psychologist-kpi__label">Всего попыток</p>
          <p className="psychologist-kpi__value">{summary.total}</p>
        </article>
        <article className="psychologist-kpi">
          <p className="psychologist-kpi__label">Завершено</p>
          <p className="psychologist-kpi__value">{summary.completed}</p>
        </article>
        <article className="psychologist-kpi">
          <p className="psychologist-kpi__label">В процессе</p>
          <p className="psychologist-kpi__value">{summary.inProgress}</p>
        </article>
        <article className="psychologist-kpi">
          <p className="psychologist-kpi__label">Средний прогресс</p>
          <p className="psychologist-kpi__value psychologist-kpi__value--small">{summary.averageProgress}%</p>
        </article>
      </section>

      <section className="admin-panels">
        <article className="admin-panel">
          <h3 className="admin-panel__title">Сводка по тесту</h3>
          <dl className="profile-meta-list">
            <div>
              <dt>Статус теста</dt>
              <dd>
                <span className={`status-badge status-badge--${testQuery.data?.status === "published" ? "active" : "draft"}`}>
                  {testQuery.data?.status === "published" ? "Опубликован" : "Черновик"}
                </span>
              </dd>
            </div>
            <div>
              <dt>Публичная ссылка</dt>
              <dd>{publicUrl || "Появится после публикации"}</dd>
            </div>
            <div>
              <dt>Обновлён</dt>
              <dd>{formatDate(testQuery.data?.updated_at || testQuery.data?.created_at)}</dd>
            </div>
            <div>
              <dt>Шаблон отчёта</dt>
              <dd>{assignedTemplate?.name || (testQuery.data?.report_template_id ? `Шаблон #${testQuery.data.report_template_id}` : "Не выбран")}</dd>
            </div>
          </dl>
        </article>

        <article className="admin-panel">
          <h3 className="admin-panel__title">Рабочие переходы</h3>
          <div className="psychologist-quick-grid psychologist-quick-grid--compact">
            <Link className="psychologist-quick-card" to={toBuilderPath(id)}>
              <LayoutTemplate size={18} strokeWidth={2.1} />
              <strong>Открыть конструктор</strong>
              <span>Редактировать структуру методики и вопросы.</span>
            </Link>
            {publicPath ? (
              <Link className="psychologist-quick-card" to={publicPath} target="_blank" rel="noreferrer">
                <Eye size={18} strokeWidth={2.1} />
                <strong>Открыть публичную форму</strong>
                <span>Проверить клиентский сценарий прохождения.</span>
              </Link>
            ) : null}
            <Link className="psychologist-quick-card" to={ROUTES.reportTemplates}>
              <LayoutTemplate size={18} strokeWidth={2.1} />
              <strong>Шаблоны отчётов</strong>
              <span>Открыть и изменить шаблоны отчётов.</span>
            </Link>
          </div>
        </article>
      </section>

      <div className="admin-table-wrap">
        <table className="admin-table">
          <thead>
            <tr>
              <th>Респондент</th>
              <th>Статус</th>
              <th>Метрики</th>
              <th>Прогресс</th>
              <th>Начало</th>
              <th>Завершение</th>
              <th>Детали</th>
            </tr>
          </thead>
          <tbody>
            {resultsQuery.isLoading ? (
              <tr>
                <td className="admin-table__empty" colSpan={7}>
                  <LoaderCircle className="icon-spin" size={16} strokeWidth={2.1} /> Загружаем прохождения...
                </td>
              </tr>
            ) : results.length ? (
              results.map((item) => {
                const canGenerateReport = isCompletedResult(item.status);

                return (
                  <tr key={item.id}>
                    <td>
                      <p className="admin-table__primary">{item.respondent}</p>
                      <p className="admin-table__secondary">{item.email || "Email не передан"}</p>
                    </td>
                    <td>
                      <span className={`status-badge status-badge--${getStatusTone(item.status)}`}>
                        {getResultStatusLabel(item.status)}
                      </span>
                    </td>
                    <td>
                      <p className="admin-table__primary admin-table__primary--small">
                        {item.topMetricSummary || "Метрики не рассчитаны"}
                      </p>
                      <p className="admin-table__secondary">
                        {item.metrics?.length ? `${item.metrics.length} показателей в результате` : "Показатели пока не рассчитаны"}
                      </p>
                    </td>
                    <td>{item.progress}%</td>
                    <td>{formatDate(item.startedAt)}</td>
                    <td>{formatDate(item.completedAt)}</td>
                    <td>
                      <div className="admin-table__actions admin-table__actions--reports">
                        <Link className="table-action-link" to={buildTestSubmissionPath(id, item.id)}>
                          <Eye size={15} strokeWidth={2.1} />
                          <span>Открыть сессию</span>
                        </Link>
                      </div>
                      {!canGenerateReport ? (
                        <p className="admin-table__secondary">До завершения доступен только просмотр прогресса и ответов.</p>
                      ) : null}
                    </td>
                  </tr>
                );
              })
            ) : (
              <tr>
                <td className="admin-table__empty" colSpan={7}>
                  По этому тесту пока нет прохождений. Опубликуй методику и выдай ссылку клиенту.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </PageCard>
  );
}
