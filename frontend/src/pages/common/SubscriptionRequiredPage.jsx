import { CreditCard, LoaderCircle, LogOut } from "lucide-react";
import { Navigate } from "react-router-dom";
import { useAuth } from "../../modules/auth/model/useAuth";
import { ROUTES } from "../../shared/config/routes";

function isFutureDate(value) {
  if (!value) {
    return false;
  }

  const timestamp = new Date(value).getTime();
  return Number.isFinite(timestamp) && timestamp > Date.now();
}

export default function SubscriptionRequiredPage() {
  const {
    user,
    signOut,
    hasActiveSubscription,
    canActivateDemoSubscription,
    activateDemoSubscription,
    isActivatingDemoSubscription,
    demoSubscriptionError,
  } = useAuth();

  if (hasActiveSubscription) {
    return <Navigate to={ROUTES.dashboard} replace />;
  }

  async function handleActivateSubscription() {
    try {
      await activateDemoSubscription();
    } catch {
      // User-facing error already handled in auth state.
    }
  }

  const isBlocked =
    user?.account_status === "blocked" ||
    user?.is_active === false ||
    isFutureDate(user?.blocked_until);

  return (
    <main className="screen">
      <section className="welcome-card welcome-card--subscription">
        <h1>{isBlocked ? "Доступ ограничен" : "Подписка неактивна"}</h1>
        <p>
          {isBlocked
            ? "Доступ к кабинету сейчас ограничен. Оплата подписки не снимет это ограничение."
            : "Для доступа к рабочим экранам нужна активная подписка."}
        </p>
        {demoSubscriptionError ? (
          <p className="welcome-card__error">{demoSubscriptionError}</p>
        ) : null}
        {!isBlocked ? (
          <p className="welcome-card__note">
            Для демо можно включить тестовую оплату. Доступ откроется сразу.
          </p>
        ) : null}
        <div className="welcome-card__actions">
          {!isBlocked ? (
            <button
              className="admin-primary-button"
              type="button"
              onClick={handleActivateSubscription}
              disabled={!canActivateDemoSubscription || isActivatingDemoSubscription}
            >
              {isActivatingDemoSubscription ? (
                <LoaderCircle className="icon-spin" size={16} strokeWidth={2.1} />
              ) : (
                <CreditCard size={16} strokeWidth={2.1} />
              )}
              <span>{isActivatingDemoSubscription ? "Проверяем оплату..." : "Оплатить подписку"}</span>
            </button>
          ) : null}
          <button className="logout-button" type="button" onClick={signOut}>
            <LogOut size={16} strokeWidth={2.1} />
            <span>Выйти</span>
          </button>
        </div>
        {!isBlocked ? (
          <p className="welcome-card__legal">
            Демонстрационный экран оплаты. Не является публичной офертой.
          </p>
        ) : null}
      </section>
    </main>
  );
}
