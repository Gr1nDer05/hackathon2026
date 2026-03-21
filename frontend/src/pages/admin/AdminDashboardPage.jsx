import { ROUTES } from "../../shared/config/routes";
import PageCard from "../../shared/ui/PageCard";
import { useAdminData } from "../../modules/admin/model/useAdminData";

function getExpiringSoonCount(subscriptions) {
  const sevenDaysFromNow = Date.now() + 7 * 24 * 60 * 60 * 1000;

  return subscriptions.filter((item) => {
    if (!item.expiresAt || item.expiresAt === "—") {
      return false;
    }

    const expiresAt = new Date(item.expiresAt).getTime();
    return (
      Number.isFinite(expiresAt) &&
      expiresAt >= Date.now() &&
      expiresAt <= sevenDaysFromNow
    );
  }).length;
}

export default function AdminDashboardPage() {
  const { psychologists, subscriptions, isLoading, error } = useAdminData();

  const stats = {
    psychologistsTotal: psychologists.length,
    activeAccounts: psychologists.filter(
      (item) => item.accountStatus === "active",
    ).length,
    activeSubscriptions: subscriptions.filter(
      (item) => item.status === "active",
    ).length,
    expiringSoon: getExpiringSoonCount(subscriptions),
  };

  return (
    <PageCard
      embedded
      title="Админ-панель"
      description="Контроль психологов, доступов и жизненного цикла подписок."
      links={[
        { to: ROUTES.adminPsychologists, label: "Психологи" },
        { to: ROUTES.adminSubscriptions, label: "Подписки" },
      ]}
    >
      {error ? (
        <p className="admin-form-message admin-form-message--error">
          {error.message || "Не удалось загрузить сводку администратора."}
        </p>
      ) : null}

      <section className="admin-stats">
        <article className="admin-stat-card">
          <p className="admin-stat-card__label">Всего психологов</p>
          <p className="admin-stat-card__value">
            {isLoading ? "..." : stats.psychologistsTotal}
          </p>
        </article>
        <article className="admin-stat-card">
          <p className="admin-stat-card__label">Активные аккаунты</p>
          <p className="admin-stat-card__value">
            {isLoading ? "..." : stats.activeAccounts}
          </p>
        </article>
        <article className="admin-stat-card">
          <p className="admin-stat-card__label">Активные подписки</p>
          <p className="admin-stat-card__value">
            {isLoading ? "..." : stats.activeSubscriptions}
          </p>
        </article>
        <article className="admin-stat-card">
          <p className="admin-stat-card__label">Истекают за 7 дней</p>
          <p className="admin-stat-card__value">
            {isLoading ? "..." : stats.expiringSoon}
          </p>
        </article>
      </section>

      <section className="admin-panels">
        <article className="admin-panel">
          <h2 className="admin-panel__title">Что проверить сегодня</h2>
          <ul className="admin-panel__list">
            <li>Истекающие подписки и просроченные продления.</li>
            <li>Заблокированные аккаунты.</li>
            <li>Психологов без активности за последние 7 дней.</li>
          </ul>
        </article>
        <article className="admin-panel">
          <h2 className="admin-panel__title">Быстрые действия</h2>
          <ul className="admin-panel__list">
            <li>Продлить подписку выбранному психологу.</li>
            <li>Открыть карточку психолога и изменить доступ.</li>
            <li>Перейти в реестр подписок.</li>
          </ul>
        </article>
      </section>
    </PageCard>
  );
}
