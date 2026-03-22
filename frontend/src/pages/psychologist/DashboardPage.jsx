import { ArrowRight, Clock3, FileText, LayoutTemplate, ListChecks, UserCircle2 } from "lucide-react";
import { useMemo } from "react";
import useSWR from "swr";
import { Link } from "react-router-dom";
import { useAuth } from "../../modules/auth/model/useAuth";
import {
  formatDate,
  getAccountStatus,
  getAccountStatusLabel,
  getPsychologistAccessUntil,
  getPsychologistDisplayName,
  getPsychologistInitials,
  getPsychologistLastActivityAt,
  getPsychologistSpecialization,
  getStatusTone,
  getSubscriptionPlan,
  getSubscriptionPlanLabel,
  getSubscriptionStatus,
  getSubscriptionStatusLabel,
  hasAiTemplateAccess,
  getTestActivityAt,
  sortTestsByRecent,
  summarizeTests,
} from "../../modules/psychologist/lib/psychologistUi";
import { listPsychologistReportTemplatesRequest, listPsychologistTestsRequest } from "../../modules/tests/api/testsApi";
import { ROUTES } from "../../shared/config/routes";
import PageCard from "../../shared/ui/PageCard";

function toBuilderPath(id) {
  return ROUTES.testBuilder.replace(":id", String(id));
}

function toResultsPath(id) {
  return ROUTES.testResults.replace(":id", String(id));
}

function getTestStatusLabel(status) {
  return status === "published" ? "Опубликован" : "Черновик";
}

export default function DashboardPage() {
  const { user } = useAuth();
  const testsQuery = useSWR("psychologist-tests", listPsychologistTestsRequest, {
    revalidateOnFocus: false,
    shouldRetryOnError: false,
  });
  const reportTemplatesQuery = useSWR(
    "psychologist-report-templates",
    listPsychologistReportTemplatesRequest,
    {
      revalidateOnFocus: false,
      shouldRetryOnError: false,
    },
  );

  const tests = testsQuery.data || user?.tests || user?.workspace?.tests || [];
  const reportTemplates = Array.isArray(reportTemplatesQuery.data) ? reportTemplatesQuery.data : [];
  const recentTests = useMemo(() => sortTestsByRecent(tests).slice(0, 4), [tests]);
  const stats = useMemo(() => summarizeTests(tests), [tests]);
  const testsWithoutReportTemplate = useMemo(
    () => tests.filter((item) => !item.report_template_id).length,
    [tests],
  );

  const displayName = getPsychologistDisplayName(user);
  const initials = getPsychologistInitials(user);
  const specialization = getPsychologistSpecialization(user);
  const accountStatus = getAccountStatus(user);
  const subscriptionStatus = getSubscriptionStatus(user);
  const subscriptionPlan = getSubscriptionPlan(user);
  const accessUntil = formatDate(getPsychologistAccessUntil(user));
  const lastActivity = formatDate(stats.lastActivityAt || getPsychologistLastActivityAt(user, recentTests));
  const featuredTest = recentTests[0];

  return (
    <PageCard
      wide
      title="Кабинет психолога"
      description="Рабочая сводка по методикам, доступу и последним действиям в системе."
      links={[
        { to: ROUTES.tests, label: "Мои опросники" },
        { to: ROUTES.profile, label: "Профиль" },
      ]}
    >
      <section className="psychologist-profile-hero">
        <div className="psychologist-profile-hero__identity">
          <div className="psychologist-profile-hero__avatar">{initials}</div>
          <div>
            <p className="psychologist-profile-hero__eyebrow">Рабочая зона психолога</p>
            <h2 className="psychologist-profile-hero__name">{displayName}</h2>
            <p className="psychologist-profile-hero__role">{specialization}</p>
          </div>
        </div>

        <div className="psychologist-profile-hero__badges">
          <span className={`status-badge status-badge--${getStatusTone(accountStatus)}`}>
            {getAccountStatusLabel(accountStatus)}
          </span>
          <span className={`status-badge status-badge--${getStatusTone(subscriptionStatus)}`}>
            {getSubscriptionStatusLabel(subscriptionStatus)}
          </span>
          <span className={`status-badge status-badge--${subscriptionPlan === "pro" ? "active" : "draft"}`}>
            {getSubscriptionPlanLabel(subscriptionPlan)}
          </span>
        </div>
      </section>

      {testsQuery.error ? (
        <p className="admin-form-message admin-form-message--error">
          {testsQuery.error.message || "Не удалось загрузить сводку по тестам."}
        </p>
      ) : null}
      {reportTemplatesQuery.error ? (
        <p className="admin-form-message admin-form-message--error">
          {reportTemplatesQuery.error.message || "Не удалось загрузить шаблоны отчётов."}
        </p>
      ) : null}

      <div className={`workflow-note ${testsWithoutReportTemplate ? "workflow-note--warning" : "workflow-note--success"}`}>
        <p>
          {testsWithoutReportTemplate
            ? `У ${testsWithoutReportTemplate} тестов ещё не настроен шаблон отчёта.`
            : "Для текущих тестов шаблоны отчётов уже готовы."}
        </p>
        <div className="workflow-note__actions">
          <Link className="table-action-link" to={ROUTES.reportTemplates}>
            <FileText size={15} strokeWidth={2.1} />
            <span>Шаблоны отчётов</span>
          </Link>
        </div>
      </div>

      <div className={`workflow-note ${hasAiTemplateAccess(user) ? "workflow-note--success" : "workflow-note--warning"}`}>
        <p>
          {hasAiTemplateAccess(user)
            ? "Автоматическое заполнение шаблонов отчётов доступно."
            : "Автоматическое заполнение шаблонов сейчас недоступно на вашем плане."}
        </p>
      </div>

      <section className="psychologist-kpis">
        <article className="psychologist-kpi">
          <p className="psychologist-kpi__label">Всего методик</p>
          <p className="psychologist-kpi__value">{stats.totalCount}</p>
        </article>
        <article className="psychologist-kpi">
          <p className="psychologist-kpi__label">Опубликовано</p>
          <p className="psychologist-kpi__value">{stats.publishedCount}</p>
        </article>
        <article className="psychologist-kpi">
          <p className="psychologist-kpi__label">Прохождений</p>
          <p className="psychologist-kpi__value">{stats.completedSessions}</p>
        </article>
        <article className="psychologist-kpi">
          <p className="psychologist-kpi__label">В работе</p>
          <p className="psychologist-kpi__value">{stats.inProgressSessions}</p>
        </article>
        <article className="psychologist-kpi">
          <p className="psychologist-kpi__label">Доступ до</p>
          <p className="psychologist-kpi__value psychologist-kpi__value--small">{accessUntil}</p>
        </article>
        <article className="psychologist-kpi">
          <p className="psychologist-kpi__label">План</p>
          <p className="psychologist-kpi__value psychologist-kpi__value--small">{getSubscriptionPlanLabel(subscriptionPlan)}</p>
        </article>
      </section>

      <section className="admin-panels">
        <article className="admin-panel">
          <h3 className="admin-panel__title">Последние опросники</h3>
          <p className="admin-panel__meta">Быстрый доступ к методикам, которые недавно редактировались.</p>

          {testsQuery.isLoading ? (
            <p className="psychologist-empty-state">Загружаем список методик...</p>
          ) : recentTests.length ? (
            <div className="psychologist-test-list">
              {recentTests.map((item) => (
                <Link className="psychologist-test-item" key={item.id} to={toBuilderPath(item.id)}>
                  <div className="psychologist-test-item__head">
                    <div>
                      <p className="psychologist-test-item__title">{item.title}</p>
                      <p className="psychologist-test-item__desc">{item.description || "Описание пока не заполнено"}</p>
                    </div>
                    <ArrowRight size={16} strokeWidth={2.1} />
                  </div>
                  <div className="psychologist-test-item__meta">
                    <span className={`status-badge status-badge--${item.status === "published" ? "active" : "draft"}`}>
                      {getTestStatusLabel(item.status)}
                    </span>
                    <span>Завершено: {item.completed_sessions_count || 0}</span>
                    <span>В работе: {item.in_progress_sessions_count || 0}</span>
                    <span>Активность: {formatDate(getTestActivityAt(item))}</span>
                  </div>
                </Link>
              ))}
            </div>
          ) : (
            <p className="psychologist-empty-state">Методик пока нет. Создай первый опросник и собери структуру теста.</p>
          )}
        </article>

        <article className="admin-panel">
          <h3 className="admin-panel__title">Быстрые действия</h3>
          <p className="admin-panel__meta">Основные разделы для работы с тестами и отчётами.</p>

          <div className="psychologist-quick-grid">
            <Link className="psychologist-quick-card" to={ROUTES.tests}>
              <LayoutTemplate size={18} strokeWidth={2.1} />
              <strong>Мои опросники</strong>
              <span>Создание, публикация и выдача клиентских ссылок.</span>
            </Link>

            <Link className="psychologist-quick-card" to={featuredTest ? toBuilderPath(featuredTest.id) : ROUTES.tests}>
              <Clock3 size={18} strokeWidth={2.1} />
              <strong>Последняя методика</strong>
              <span>{featuredTest ? `Продолжить работу с «${featuredTest.title}».` : "Открой список тестов и создай первую методику."}</span>
            </Link>

            <Link className="psychologist-quick-card" to={featuredTest ? toResultsPath(featuredTest.id) : ROUTES.tests}>
              <ListChecks size={18} strokeWidth={2.1} />
              <strong>Результаты</strong>
              <span>{featuredTest ? "Посмотреть прохождения и прогресс по последнему тесту." : "Результаты появятся после первых прохождений."}</span>
            </Link>

            <Link className="psychologist-quick-card" to={ROUTES.reportTemplates}>
              <FileText size={18} strokeWidth={2.1} />
              <strong>Шаблоны отчётов</strong>
              <span>{reportTemplates.length ? `Доступно шаблонов: ${reportTemplates.length}.` : "Создайте первый шаблон отчёта."}</span>
            </Link>

            <Link className="psychologist-quick-card" to={ROUTES.profile}>
              <UserCircle2 size={18} strokeWidth={2.1} />
              <strong>Профиль</strong>
              <span>Контакты, специализация и данные специалиста.</span>
            </Link>
          </div>
        </article>
      </section>

      <section className="admin-panel admin-panel--spaced">
        <h3 className="admin-panel__title">Обзор</h3>
        <div className="psychologist-summary-grid">
          <div className="psychologist-summary-card">
            <span>Черновики</span>
            <strong>{stats.draftCount}</strong>
            <p>Тесты, которые ещё не опубликованы.</p>
          </div>
          <div className="psychologist-summary-card">
            <span>Стартов</span>
            <strong>{stats.startedSessions}</strong>
            <p>Сколько раз пользователи начинали прохождение.</p>
          </div>
          <div className="psychologist-summary-card">
            <span>Последняя активность</span>
            <strong>{lastActivity}</strong>
            <p>Последнее действие по тестам и результатам.</p>
          </div>
          <div className="psychologist-summary-card">
            <span>Публичная выдача</span>
            <strong>{stats.publishedCount ? "Готова" : "Не настроена"}</strong>
            <p>Опубликованные тесты можно отправлять клиентам по ссылке.</p>
          </div>
          <div className="psychologist-summary-card">
            <span>Шаблоны</span>
            <strong>{reportTemplates.length}</strong>
            <p>Шаблоны, которые можно привязать к тестам.</p>
          </div>
        </div>
      </section>
    </PageCard>
  );
}
