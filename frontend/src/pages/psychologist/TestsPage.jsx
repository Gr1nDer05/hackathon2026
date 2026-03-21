import { Copy, Eye, FilePlus2, LayoutTemplate, ListChecks, LoaderCircle, Plus, X } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import useSWR from "swr";
import useSWRMutation from "swr/mutation";
import { Link } from "react-router-dom";
import {
  createPsychologistTestRequest,
  listPsychologistReportTemplatesRequest,
  listPsychologistTestsRequest,
} from "../../modules/tests/api/testsApi";
import { formatDate as formatUiDate, getTestActivityAt, summarizeTests } from "../../modules/psychologist/lib/psychologistUi";
import { buildClientSessionPath, ROUTES } from "../../shared/config/routes";
import PageCard from "../../shared/ui/PageCard";

const EMPTY_FORM = {
  title: "",
  description: "",
  status: "draft",
};

function toBuilderPath(id) {
  return ROUTES.testBuilder.replace(":id", String(id));
}

function toResultsPath(id) {
  return ROUTES.testResults.replace(":id", String(id));
}

function formatDate(value) {
  if (!value) {
    return "—";
  }

  const timestamp = new Date(value).getTime();
  if (!Number.isFinite(timestamp)) {
    return "—";
  }

  return new Date(value).toLocaleDateString("ru-RU");
}

function getLastActivityHint(test) {
  if (test?.last_completed_at) {
    return `Последнее завершение: ${formatUiDate(test.last_completed_at)}`;
  }

  if (test?.last_started_at) {
    return `Последний старт: ${formatUiDate(test.last_started_at)}`;
  }

  return "Ручное обновление или создание теста.";
}

function validateTestForm(form) {
  const errors = {};

  if (!form.title.trim()) {
    errors.title = "Укажите название опросника.";
  }

  if (form.status && !["draft", "published"].includes(form.status)) {
    errors.status = "Укажите корректный статус.";
  }

  return errors;
}

function getStatusLabel(status) {
  if (status === "published") {
    return "published";
  }

  return "draft";
}

export default function TestsPage() {
  const testsQuery = useSWR("psychologist-tests", listPsychologistTestsRequest, {
    revalidateOnFocus: false,
    shouldRetryOnError: false,
  });
  const reportTemplatesQuery = useSWR("psychologist-report-templates", listPsychologistReportTemplatesRequest, {
    revalidateOnFocus: false,
    shouldRetryOnError: false,
  });
  const createTest = useSWRMutation(
    "psychologist-tests-create",
    async (_, { arg }) => createPsychologistTestRequest(arg),
    {
      onSuccess() {
        testsQuery.mutate();
      },
    },
  );

  const tests = testsQuery.data || [];
  const testsWithoutReportTemplate = useMemo(
    () => tests.filter((item) => !item.report_template_id).length,
    [tests],
  );
  const stats = useMemo(() => summarizeTests(tests), [tests]);
  const reportTemplateNameById = useMemo(() => {
    const entries = Array.isArray(reportTemplatesQuery.data) ? reportTemplatesQuery.data : [];

    return Object.fromEntries(
      entries
        .filter((template) => template?.id !== null && template?.id !== undefined)
        .map((template) => [String(template.id), template.name || `Шаблон #${template.id}`]),
    );
  }, [reportTemplatesQuery.data]);
  const [query, setQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState("all");
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [form, setForm] = useState(EMPTY_FORM);
  const [formErrors, setFormErrors] = useState({});
  const [submitError, setSubmitError] = useState("");
  const [actionMessage, setActionMessage] = useState("");
  const [actionError, setActionError] = useState("");

  useEffect(() => {
    if (!actionMessage) {
      return undefined;
    }

    const timeoutId = window.setTimeout(() => setActionMessage(""), 2400);
    return () => window.clearTimeout(timeoutId);
  }, [actionMessage]);

  const filteredTests = useMemo(() => {
    const normalizedQuery = query.trim().toLowerCase();

    return tests.filter((item) => {
      const matchesQuery =
        !normalizedQuery ||
        item.title.toLowerCase().includes(normalizedQuery) ||
        String(item.description || "").toLowerCase().includes(normalizedQuery);

      const matchesStatus = statusFilter === "all" || item.status === statusFilter;

      return matchesQuery && matchesStatus;
    });
  }, [query, statusFilter, tests]);

  function handleFormChange(field, value) {
    setForm((prev) => ({ ...prev, [field]: value }));
    setSubmitError("");
    setFormErrors((prev) => {
      if (!prev[field]) {
        return prev;
      }

      const next = { ...prev };
      delete next[field];
      return next;
    });
  }

  async function handleCreateTest(event) {
    event.preventDefault();
    const errors = validateTestForm(form);

    if (Object.keys(errors).length > 0) {
      setFormErrors(errors);
      return;
    }

    setSubmitError("");
    setActionMessage("");
    setActionError("");

    try {
      await createTest.trigger({
        title: form.title.trim(),
        description: form.description.trim(),
        status: form.status,
      });

      setForm(EMPTY_FORM);
      setFormErrors({});
      setIsCreateOpen(false);
      setActionMessage("Опросник создан.");
    } catch (error) {
      setSubmitError(error?.message || "Не удалось создать опросник.");
    }
  }

  async function handleCopyPublicUrl(publicSlug) {
    if (!publicSlug) {
      setActionError("У этого опросника пока нет публичной ссылки.");
      return;
    }

    setActionError("");

    try {
      const publicClientUrl = `${window.location.origin}${buildClientSessionPath(publicSlug)}`;
      await navigator.clipboard.writeText(publicClientUrl);
      setActionMessage("Публичная ссылка скопирована.");
    } catch {
      setActionError("Не удалось скопировать ссылку в буфер обмена.");
    }
  }

  return (
    <PageCard
      wide
      title="Мои опросники"
      description="Список методик психолога, статусы публикации и быстрые действия."
      links={[
        { to: ROUTES.dashboard, label: "Назад в кабинет" },
        { to: ROUTES.reportTemplates, label: "Шаблоны отчётов" },
        { to: ROUTES.profile, label: "Профиль" },
      ]}
    >
      <section className="admin-page-actions">
        <button
          className="admin-primary-button"
          type="button"
          onClick={() => {
            setIsCreateOpen((prev) => !prev);
            setSubmitError("");
          }}
        >
          {isCreateOpen ? <X size={16} strokeWidth={2.1} /> : <Plus size={16} strokeWidth={2.1} />}
          <span>{isCreateOpen ? "Скрыть форму" : "Создать опросник"}</span>
        </button>
      </section>

      {isCreateOpen ? (
        <section className="admin-create-panel">
          <div className="admin-create-panel__header">
            <div>
              <p className="admin-create-panel__eyebrow">Создание методики</p>
              <h2 className="admin-create-panel__title">Новый опросник</h2>
            </div>
            <p className="admin-create-panel__meta">
              Сначала создаём основу теста, затем настраиваем вопросы, формулы и публикацию.
            </p>
          </div>

          <form className="admin-form-grid" onSubmit={handleCreateTest}>
            <label className="admin-form-field">
              <span>Название</span>
              <input
                aria-invalid={Boolean(formErrors.title)}
                className={formErrors.title ? "admin-form-control admin-form-control--invalid" : "admin-form-control"}
                value={form.title}
                onChange={(event) => handleFormChange("title", event.target.value)}
                placeholder="Например: Профориентационный базовый тест"
              />
              {formErrors.title ? (
                <small className="admin-form-error admin-form-error--inline">{formErrors.title}</small>
              ) : null}
            </label>

            <label className="admin-form-field">
              <span>Статус</span>
              <select
                aria-invalid={Boolean(formErrors.status)}
                className={formErrors.status ? "admin-form-control admin-form-control--invalid" : "admin-form-control"}
                value={form.status}
                onChange={(event) => handleFormChange("status", event.target.value)}
              >
                <option value="draft">Черновик</option>
                <option value="published">Опубликован</option>
              </select>
              {formErrors.status ? (
                <small className="admin-form-error admin-form-error--inline">{formErrors.status}</small>
              ) : null}
            </label>

            <label className="admin-form-field admin-form-field--wide">
              <span>Описание</span>
              <input
                className="admin-form-control"
                value={form.description}
                onChange={(event) => handleFormChange("description", event.target.value)}
                placeholder="Краткое описание методики и её назначения"
              />
            </label>

            <div className="admin-form-actions">
              <button className="admin-primary-button" disabled={createTest.isMutating} type="submit">
                {createTest.isMutating ? (
                  <LoaderCircle className="icon-spin" size={16} strokeWidth={2.1} />
                ) : (
                  <FilePlus2 size={16} strokeWidth={2.1} />
                )}
                <span>{createTest.isMutating ? "Создание..." : "Создать опросник"}</span>
              </button>
            </div>
            {submitError ? (
              <p className="admin-form-message admin-form-message--error">{submitError}</p>
            ) : null}
          </form>
        </section>
      ) : null}

      {testsQuery.error ? (
        <p className="admin-form-message admin-form-message--error">
          {testsQuery.error.message || "Не удалось загрузить список опросников."}
        </p>
      ) : null}

      {actionError ? (
        <p className="admin-form-message admin-form-message--error">{actionError}</p>
      ) : null}

      {actionMessage ? <p className="admin-form-message">{actionMessage}</p> : null}
      {reportTemplatesQuery.error ? (
        <p className="admin-form-message admin-form-message--error">
          {reportTemplatesQuery.error.message || "Не удалось загрузить шаблоны отчётов."}
        </p>
      ) : null}

      <section className="psychologist-summary-grid">
        <div className="psychologist-summary-card">
          <span>Всего тестов</span>
          <strong>{stats.totalCount}</strong>
          <p>Общее число методик, доступных в рабочем контуре психолога.</p>
        </div>
        <div className="psychologist-summary-card">
          <span>Стартов</span>
          <strong>{stats.startedSessions}</strong>
          <p>Сколько раз опубликованные тесты уже были открыты и запущены пользователями.</p>
        </div>
        <div className="psychologist-summary-card">
          <span>В работе</span>
          <strong>{stats.inProgressSessions}</strong>
          <p>Сессии, которые ещё не завершены и могут вернуться в отчёты позже.</p>
        </div>
        <div className="psychologist-summary-card">
          <span>Последняя активность</span>
          <strong>{formatDate(stats.lastActivityAt)}</strong>
          <p>Последний старт, completion или ручное изменение одной из методик.</p>
        </div>
      </section>

      <div className={`workflow-note ${testsWithoutReportTemplate ? "workflow-note--warning" : "workflow-note--success"}`}>
        <p>
          {testsWithoutReportTemplate
            ? `У ${testsWithoutReportTemplate} тестов ещё нет шаблона отчёта. Чтобы получать HTML и DOCX без ручной доработки, привяжи шаблон в конструкторе.`
            : "У всех текущих тестов есть привязанный шаблон отчёта или отчётный контур уже готов к настройке."}
        </p>
        <div className="workflow-note__actions">
          <Link className="table-action-link" to={ROUTES.reportTemplates}>
            <LayoutTemplate size={15} strokeWidth={2.1} />
            <span>Открыть шаблоны отчётов</span>
          </Link>
        </div>
      </div>

      <section className="admin-tools admin-tools--compact">
        <input
          className="admin-tools__input"
          type="text"
          value={query}
          onChange={(event) => setQuery(event.target.value)}
          placeholder="Поиск по названию или описанию"
        />

        <select
          className="admin-tools__select"
          value={statusFilter}
          onChange={(event) => setStatusFilter(event.target.value)}
        >
          <option value="all">Все статусы</option>
          <option value="draft">Черновики</option>
          <option value="published">Опубликованные</option>
        </select>
      </section>

      <div className="admin-table-wrap">
        <table className="admin-table">
          <thead>
            <tr>
              <th>Опросник</th>
              <th>Статус</th>
              <th>Активность</th>
              <th>Отчёт</th>
              <th>Публичная ссылка</th>
              <th>Последнее действие</th>
              <th>Действия</th>
            </tr>
          </thead>
          <tbody>
            {testsQuery.isLoading ? (
              <tr>
                <td className="admin-table__empty" colSpan={7}>
                  Загружаем опросники...
                </td>
              </tr>
            ) : filteredTests.length ? (
              filteredTests.map((item) => (
                <tr key={item.id}>
                  <td>
                    <p className="admin-table__primary">{item.title}</p>
                    <p className="admin-table__secondary">
                      {item.description || "Описание пока не заполнено"}
                    </p>
                  </td>
                  <td>
                    <span
                      className={`status-badge status-badge--${
                        item.status === "published" ? "active" : "draft"
                      }`}
                    >
                      {getStatusLabel(item.status)}
                    </span>
                  </td>
                  <td>
                    <p className="admin-table__primary admin-table__primary--small">
                      Завершено: {item.completed_sessions_count || 0}
                    </p>
                    <p className="admin-table__secondary">
                      Стартов: {item.started_sessions_count || 0} · В работе: {item.in_progress_sessions_count || 0}
                    </p>
                  </td>
                  <td>
                    <p className="admin-table__primary admin-table__primary--small">
                      {item.report_template_id ? "Подключён" : "Не выбран"}
                    </p>
                    <p className="admin-table__secondary">
                      {item.report_template_id
                        ? reportTemplateNameById[String(item.report_template_id)] || `Шаблон #${item.report_template_id}`
                        : "Назначается в конструкторе"}
                    </p>
                  </td>
                  <td>
                    <p className="admin-table__primary admin-table__primary--small">
                      {item.public_slug ? "Ссылка готова" : "Не опубликован"}
                    </p>
                    <p className="admin-table__secondary">
                      {item.public_slug || "—"}
                    </p>
                  </td>
                  <td>
                    <p className="admin-table__primary admin-table__primary--small">
                      {formatDate(getTestActivityAt(item))}
                    </p>
                    <p className="admin-table__secondary">{getLastActivityHint(item)}</p>
                  </td>
                  <td>
                    <div className="admin-table__actions">
                      <Link className="table-action-link" to={toBuilderPath(item.id)}>
                        <LayoutTemplate size={15} strokeWidth={2.1} />
                        <span>Конструктор</span>
                      </Link>
                      <Link className="table-action-link" to={toResultsPath(item.id)}>
                        <ListChecks size={15} strokeWidth={2.1} />
                        <span>Результаты</span>
                      </Link>
                      {item.public_slug ? (
                        <>
                          <button
                            className="table-action-button"
                            type="button"
                            onClick={() => handleCopyPublicUrl(item.public_slug)}
                          >
                            <Copy size={15} strokeWidth={2.1} />
                            <span>Скопировать</span>
                          </button>
                          <Link className="table-action-link" to={buildClientSessionPath(item.public_slug)} target="_blank" rel="noreferrer">
                            <Eye size={15} strokeWidth={2.1} />
                            <span>Открыть</span>
                          </Link>
                        </>
                      ) : null}
                    </div>
                  </td>
                </tr>
              ))
            ) : (
              <tr>
                <td className="admin-table__empty" colSpan={7}>
                  {tests.length
                    ? "Ничего не найдено. Измени фильтр или поисковый запрос."
                    : "У вас пока нет опросников. Создайте первый тест, чтобы начать работу."}
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </PageCard>
  );
}
