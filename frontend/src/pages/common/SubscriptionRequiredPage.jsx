import { CreditCard, LoaderCircle, LogOut, RefreshCw } from "lucide-react";
import { useEffect, useState } from "react";
import { Navigate } from "react-router-dom";
import { useAuth } from "../../modules/auth/model/useAuth";
import {
  getSubscriptionPlan,
  getSubscriptionPlanLabel,
} from "../../modules/psychologist/lib/psychologistUi";
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
    createSubscriptionPurchaseRequest,
    isCreatingSubscriptionPurchaseRequest,
    refreshSession,
  } = useAuth();
  const [requestedPlan, setRequestedPlan] = useState("basic");
  const [requestInfo, setRequestInfo] = useState(null);
  const [actionError, setActionError] = useState("");
  const [isRefreshingStatus, setIsRefreshingStatus] = useState(false);

  useEffect(() => {
    setRequestedPlan(getSubscriptionPlan(user));
  }, [user]);

  if (hasActiveSubscription) {
    return <Navigate to={ROUTES.dashboard} replace />;
  }

  async function handleCreatePurchaseRequest() {
    try {
      setActionError("");
      const response = await createSubscriptionPurchaseRequest({
        subscriptionPlan: requestedPlan,
      });
      setRequestInfo(response);
    } catch {
      setActionError("Не удалось отправить заявку. Попробуйте ещё раз.");
    }
  }

  async function handleRefreshStatus() {
    setIsRefreshingStatus(true);

    try {
      await refreshSession();
    } finally {
      setIsRefreshingStatus(false);
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
            ? "Доступ к кабинету сейчас ограничен. Для разбора ситуации свяжитесь с администратором."
            : "Чтобы вернуться в кабинет, отправьте заявку на нужный план. Администратор обработает её вручную."}
        </p>
        {actionError ? (
          <p className="welcome-card__error">{actionError}</p>
        ) : null}
        {!isBlocked ? (
          <p className="welcome-card__note">
            После подтверждения заявки доступ откроется на 30 дней.
          </p>
        ) : null}
        {!isBlocked ? (
          <label className="admin-form-field welcome-card__field">
            <span>План подписки</span>
            <select
              className="admin-form-control"
              value={requestedPlan}
              onChange={(event) => setRequestedPlan(event.target.value)}
              disabled={isCreatingSubscriptionPurchaseRequest}
            >
              <option value="basic">Basic</option>
              <option value="pro">Pro</option>
            </select>
          </label>
        ) : null}
        {requestInfo ? (
          <p className="welcome-card__note">
            Заявка на план <strong>{getSubscriptionPlanLabel(requestInfo.subscription_plan || requestedPlan)}</strong> отправлена.
            Срок доступа после подтверждения: {requestInfo.duration_days || 30} дней.
          </p>
        ) : null}
        <div className="welcome-card__actions">
          {!isBlocked ? (
            <button
              className="admin-primary-button"
              type="button"
              onClick={handleCreatePurchaseRequest}
              disabled={isCreatingSubscriptionPurchaseRequest}
            >
              {isCreatingSubscriptionPurchaseRequest ? (
                <LoaderCircle className="icon-spin" size={16} strokeWidth={2.1} />
              ) : (
                <CreditCard size={16} strokeWidth={2.1} />
              )}
              <span>{isCreatingSubscriptionPurchaseRequest ? "Отправляем..." : "Оставить заявку"}</span>
            </button>
          ) : null}
          {!isBlocked ? (
            <button
              className="logout-button"
              type="button"
              onClick={handleRefreshStatus}
              disabled={isRefreshingStatus}
            >
              {isRefreshingStatus ? (
                <LoaderCircle className="icon-spin" size={16} strokeWidth={2.1} />
              ) : (
                <RefreshCw size={16} strokeWidth={2.1} />
              )}
              <span>Проверить статус</span>
            </button>
          ) : null}
          <button className="logout-button" type="button" onClick={signOut}>
            <LogOut size={16} strokeWidth={2.1} />
            <span>Выйти</span>
          </button>
        </div>
        {!isBlocked ? (
          <p className="welcome-card__legal">
            Экран оформления заявки. Не является публичной офертой.
          </p>
        ) : null}
      </section>
    </main>
  );
}
