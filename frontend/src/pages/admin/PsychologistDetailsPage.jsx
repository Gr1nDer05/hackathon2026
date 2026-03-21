import { Ban, CalendarPlus, ShieldCheck } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { useParams } from "react-router-dom";
import { useAdminData } from "../../modules/admin/model/useAdminData";
import { ROUTES } from "../../shared/config/routes";
import PageCard from "../../shared/ui/PageCard";

const EXTEND_OPTIONS = [7, 14, 30, 60, 90];

function getSubscriptionTone(status) {
  if (status === "active") return "Нормальный рабочий статус.";
  if (status === "expired") return "Доступ к рабочим маршрутам должен быть ограничен.";
  return "Требуется ручная проверка администратором.";
}

export default function PsychologistDetailsPage() {
  const { id } = useParams();
  const [extendDays, setExtendDays] = useState(30);
  const [actionError, setActionError] = useState("");
  const {
    psychologists,
    subscriptions,
    togglePsychologistStatus,
    extendPsychologistSubscription,
    ensureWorkspace,
    workspaceLoadingById,
    workspaceErrorById,
    isLoading,
    error,
  } = useAdminData();

  useEffect(() => {
    if (id) {
      ensureWorkspace(id).catch(() => {});
    }
  }, [ensureWorkspace, id]);

  const psychologist = useMemo(
    () => psychologists.find((item) => item.id === id) || psychologists[0],
    [id, psychologists],
  );

  const subscription = useMemo(
    () => subscriptions.find((item) => item.psychologistId === psychologist?.id) || subscriptions[0],
    [psychologist?.id, subscriptions],
  );

  const isWorkspaceLoading = psychologist ? workspaceLoadingById[psychologist.id] : false;
  const workspaceError = psychologist ? workspaceErrorById[psychologist.id] : null;

  async function handleToggleAccount() {
    if (!psychologist) {
      return;
    }

    setActionError("");

    try {
      await togglePsychologistStatus(psychologist.id);
    } catch (requestError) {
      setActionError(
        requestError?.message || "Не удалось обновить статус доступа. Повторите попытку.",
      );
    }
  }

  async function handleExtend() {
    if (!psychologist) {
      return;
    }

    setActionError("");

    try {
      await extendPsychologistSubscription(psychologist.id, extendDays);
    } catch (requestError) {
      setActionError(
        requestError?.message || "Не удалось продлить доступ. Повторите попытку.",
      );
    }
  }

  if (isLoading && !psychologist) {
    return (
      <PageCard
        embedded
        title="Профиль психолога"
        description="Загружаем данные психолога."
        links={[{ to: ROUTES.adminPsychologists, label: "К списку психологов" }]}
      >
        <p className="admin-form-message">Загружаем профиль психолога...</p>
      </PageCard>
    );
  }

  if (!psychologist || !subscription) {
    return (
      <PageCard
        embedded
        title="Психолог не найден"
        description="Запрошенный профиль не найден в системе."
        links={[{ to: ROUTES.adminPsychologists, label: "К списку психологов" }]}
      >
        <p className="admin-form-message admin-form-message--error">
          Не удалось найти психолога с указанным идентификатором.
        </p>
      </PageCard>
    );
  }

  return (
    <PageCard
      embedded
      title={`Психолог: ${psychologist.name}`}
      description="Профиль психолога, доступ и управление подпиской."
      links={[
        { to: ROUTES.adminPsychologists, label: "К списку психологов" },
        { to: ROUTES.adminSubscriptions, label: "К подпискам" },
      ]}
    >
      {error ? (
        <p className="admin-form-message admin-form-message--error">
          {error.message || "Не удалось загрузить данные администратора."}
        </p>
      ) : null}

      {workspaceError ? (
        <p className="admin-form-message admin-form-message--error">
          {workspaceError.message || "Не удалось загрузить рабочее пространство психолога."}
        </p>
      ) : null}

      {actionError ? (
        <p className="admin-form-message admin-form-message--error">{actionError}</p>
      ) : null}

      <section className="psychologist-profile-hero">
        <div className="psychologist-profile-hero__identity">
          <div className="psychologist-profile-hero__avatar" aria-hidden="true">
            {psychologist.name
              .split(" ")
              .map((part) => part[0])
              .join("")
              .slice(0, 2)}
          </div>
          <div>
            <p className="psychologist-profile-hero__eyebrow">Профиль психолога</p>
            <h2 className="psychologist-profile-hero__name">{psychologist.name}</h2>
            <p className="psychologist-profile-hero__role">{psychologist.specialization}</p>
          </div>
        </div>

        <div className="psychologist-profile-hero__badges">
          <span className={`status-badge status-badge--${psychologist.accountStatus}`}>
            {psychologist.accountStatus}
          </span>
          <span className={`status-badge status-badge--${subscription.status}`}>
            {subscription.status}
          </span>
        </div>
      </section>

      <section className="psychologist-kpis">
        <article className="psychologist-kpi">
          <p className="psychologist-kpi__label">Тестов</p>
          <p className="psychologist-kpi__value">{isWorkspaceLoading ? "..." : psychologist.testsCount}</p>
        </article>
        <article className="psychologist-kpi">
          <p className="psychologist-kpi__label">Прохождений</p>
          <p className="psychologist-kpi__value">{isWorkspaceLoading ? "..." : psychologist.sessionsCount}</p>
        </article>
        <article className="psychologist-kpi">
          <p className="psychologist-kpi__label">Completion rate</p>
          <p className="psychologist-kpi__value">
            {isWorkspaceLoading ? "..." : `${psychologist.completionRate}%`}
          </p>
        </article>
        <article className="psychologist-kpi">
          <p className="psychologist-kpi__label">Последняя активность</p>
          <p className="psychologist-kpi__value psychologist-kpi__value--small">
            {psychologist.lastActiveAt}
          </p>
        </article>
      </section>

      <section className="psychologist-profile-grid">
        <article className="admin-panel">
          <h2 className="admin-panel__title">Контактные данные</h2>
          <dl className="profile-meta-list">
            <div>
              <dt>ID</dt>
              <dd>{psychologist.backendId}</dd>
            </div>
            <div>
              <dt>Email</dt>
              <dd>{psychologist.email}</dd>
            </div>
            <div>
              <dt>Телефон</dt>
              <dd>{psychologist.phone}</dd>
            </div>
            <div>
              <dt>Город</dt>
              <dd>{psychologist.city}</dd>
            </div>
            <div>
              <dt>Пароль доступа</dt>
              <dd>
                {psychologist.hasPassword
                  ? `Задан ${psychologist.passwordUpdatedAt || "администратором"}`
                  : "Не задан"}
              </dd>
            </div>
            <div>
              <dt>В системе с</dt>
              <dd>{psychologist.joinedAt}</dd>
            </div>
          </dl>
        </article>

        <article className="admin-panel">
          <h2 className="admin-panel__title">Доступ и политика</h2>
          <p className="admin-panel__meta">
            Статус аккаунта:{" "}
            <span className={`status-badge status-badge--${psychologist.accountStatus}`}>
              {psychologist.accountStatus}
            </span>
          </p>
          <p className="admin-panel__meta">
            Если аккаунт заблокирован, психолог не должен работать с кабинетами и выдачей ссылок.
          </p>
          <div className="admin-table__actions">
            <button className="table-action-button" type="button" onClick={handleToggleAccount}>
              {psychologist.accountStatus === "blocked" ? (
                <ShieldCheck size={15} strokeWidth={2.1} />
              ) : (
                <Ban size={15} strokeWidth={2.1} />
              )}
              <span>
                {psychologist.accountStatus === "blocked"
                  ? "Разблокировать аккаунт"
                  : "Заблокировать аккаунт"}
                </span>
              </button>
          </div>
        </article>

        <article className="admin-panel">
          <h2 className="admin-panel__title">Подписка</h2>
          <dl className="profile-meta-list">
            <div>
              <dt>Тип доступа</dt>
              <dd>{subscription.plan}</dd>
            </div>
            <div>
              <dt>Статус</dt>
              <dd>
                <span className={`status-badge status-badge--${subscription.status}`}>
                  {subscription.status}
                </span>
              </dd>
            </div>
            <div>
              <dt>Старт</dt>
              <dd>{subscription.startedAt}</dd>
            </div>
            <div>
              <dt>Окончание</dt>
              <dd>{subscription.expiresAt}</dd>
            </div>
          </dl>
          <p className="admin-panel__meta">{getSubscriptionTone(subscription.status)}</p>
          <label className="admin-inline-field">
            <span className="admin-inline-field__label">Добавить дней</span>
            <input
              className="admin-inline-field__input"
              list="subscription-extend-options"
              inputMode="numeric"
              value={extendDays}
              onChange={(event) => setExtendDays(Number(event.target.value) || 1)}
            />
            <datalist id="subscription-extend-options">
              {EXTEND_OPTIONS.map((days) => (
                <option key={days} value={days} />
              ))}
            </datalist>
          </label>
          <div className="admin-table__actions admin-table__actions--subscription">
            <button className="table-action-button" type="button" onClick={handleExtend}>
              <CalendarPlus size={15} strokeWidth={2.1} />
              <span>Продлить на {extendDays} дн.</span>
            </button>
          </div>
        </article>

        <article className="admin-panel">
          <h2 className="admin-panel__title">Последние сигналы</h2>
          <ul className="admin-panel__list">
            <li>Последняя активность: {psychologist.lastActiveAt}</li>
            <li>Активных методик: {Math.max(psychologist.testsCount, 0)}</li>
            <li>Новых прохождений за неделю: {Math.max(Math.round(psychologist.sessionsCount / 12), 0)}</li>
            <li>Приоритет проверки: {subscription.status === "expired" ? "высокий" : "нормальный"}</li>
          </ul>
        </article>
      </section>
    </PageCard>
  );
}
