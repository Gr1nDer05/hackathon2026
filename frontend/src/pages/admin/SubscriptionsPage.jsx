import { CalendarPlus, IdCard } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { useAdminData } from "../../modules/admin/model/useAdminData";
import { ROUTES } from "../../shared/config/routes";
import PageCard from "../../shared/ui/PageCard";

const EXTEND_OPTIONS = [7, 14, 30, 60, 90];

function toPsychologistPath(id) {
  return ROUTES.adminPsychologistById.replace(":id", id);
}

export default function SubscriptionsPage() {
  const {
    subscriptions,
    purchaseRequests,
    extendSubscription,
    processPurchaseRequest,
    isLoading,
    error,
    purchaseRequestsError,
  } = useAdminData();
  const [statusFilter, setStatusFilter] = useState("all");
  const [actionError, setActionError] = useState("");
  const [actionMessage, setActionMessage] = useState("");
  const [processingRequestId, setProcessingRequestId] = useState("");
  const [extendDaysById, setExtendDaysById] = useState(
    Object.fromEntries(subscriptions.map((item) => [item.id, 30])),
  );

  useEffect(() => {
    setExtendDaysById((prev) => {
      const next = { ...prev };

      subscriptions.forEach((item) => {
        if (!next[item.id]) {
          next[item.id] = 30;
        }
      });

      return next;
    });
  }, [subscriptions]);

  const filteredItems = useMemo(
    () =>
      subscriptions.filter(
        (item) => statusFilter === "all" || item.status === statusFilter,
      ),
    [subscriptions, statusFilter],
  );

  async function handleExtend(id) {
    const days = extendDaysById[id] || 30;
    setActionError("");
    setActionMessage("");

    try {
      await extendSubscription(id, days);
      setActionMessage("Доступ продлён.");
    } catch (requestError) {
      setActionError(
        requestError?.message || "Не удалось продлить доступ. Повторите попытку.",
      );
    }
  }

  async function handleProcessPurchaseRequest(requestId) {
    setActionError("");
    setActionMessage("");
    setProcessingRequestId(String(requestId));

    try {
      await processPurchaseRequest(requestId);
      setActionMessage("Заявка обработана. Доступ обновлён.");
    } catch (requestError) {
      setActionError(
        requestError?.message || "Не удалось обработать заявку. Повторите попытку.",
      );
    } finally {
      setProcessingRequestId("");
    }
  }

  return (
    <PageCard
      embedded
      title="Подписки"
      description="Реестр подписок и управление статусами."
      links={[
        { to: ROUTES.adminDashboard, label: "Назад в админ-панель" },
        { to: ROUTES.adminPsychologists, label: "Психологи" },
      ]}
    >
      {error ? (
        <p className="admin-form-message admin-form-message--error">
          {error.message || "Не удалось загрузить реестр подписок."}
        </p>
      ) : null}

      {actionError ? (
        <p className="admin-form-message admin-form-message--error">{actionError}</p>
      ) : null}
      {actionMessage ? <p className="admin-form-message">{actionMessage}</p> : null}
      {purchaseRequestsError ? (
        <p className="admin-form-message admin-form-message--error">
          {purchaseRequestsError.message || "Не удалось загрузить заявки на подписку."}
        </p>
      ) : null}

      {purchaseRequests.length ? (
        <>
          <div className="workflow-note workflow-note--warning">
            <p>
              Есть новые заявки на подписку: {purchaseRequests.length}. Сначала обработайте их, потом переходите к общему реестру.
            </p>
          </div>

          <div className="admin-table-wrap">
            <table className="admin-table">
              <thead>
                <tr>
                  <th>Психолог</th>
                  <th>План</th>
                  <th>Срок</th>
                  <th>Статус</th>
                  <th>Действия</th>
                </tr>
              </thead>
              <tbody>
                {purchaseRequests.map((request) => (
                  <tr key={`purchase-request-row-${request.id}`}>
                    <td>
                      <p className="admin-table__primary">{request.psychologistName}</p>
                      <p className="admin-table__secondary">{request.psychologistEmail}</p>
                    </td>
                    <td>{request.plan}</td>
                    <td>{request.durationDays} дн.</td>
                    <td>
                      <span className={`status-badge status-badge--${request.status === "pending" ? "warning" : "active"}`}>
                        {request.status === "pending" ? "Ожидает" : "Обработана"}
                      </span>
                    </td>
                    <td>
                      <div className="admin-table__actions">
                        <Link
                          className="table-action-link"
                          to={toPsychologistPath(request.psychologistId)}
                        >
                          <IdCard size={15} strokeWidth={2.1} />
                          <span>Карточка</span>
                        </Link>
                        <button
                          className="table-action-button"
                          type="button"
                          onClick={() => handleProcessPurchaseRequest(request.id)}
                          disabled={processingRequestId === request.id}
                        >
                          <CalendarPlus size={15} strokeWidth={2.1} />
                          <span>{processingRequestId === request.id ? "Обрабатываем..." : "Выдать доступ"}</span>
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </>
      ) : null}

      <section className="admin-tools admin-tools--short">
        <select
          className="admin-tools__select"
          value={statusFilter}
          onChange={(event) => setStatusFilter(event.target.value)}
        >
          <option value="all">Все статусы</option>
          <option value="active">Активные</option>
          <option value="expired">Просроченные</option>
          <option value="blocked">Заблокированные</option>
        </select>
      </section>

      <div className="admin-table-wrap">
        <table className="admin-table">
          <thead>
            <tr>
              <th>Психолог</th>
              <th>Тариф</th>
              <th>Статус</th>
              <th>Период</th>
              <th>Действия</th>
            </tr>
          </thead>
          <tbody>
            {isLoading ? (
              <tr>
                <td className="admin-table__empty" colSpan={5}>
                  Загружаем подписки...
                </td>
              </tr>
            ) : filteredItems.length ? (
              filteredItems.map((item) => (
                <tr key={item.id}>
                  <td>{item.psychologistName}</td>
                  <td>{item.plan}</td>
                  <td>
                    <span className={`status-badge status-badge--${item.status}`}>
                      {item.status}
                    </span>
                  </td>
                  <td>
                    {item.startedAt} - {item.expiresAt}
                  </td>
                  <td>
                    <div className="admin-table__actions">
                      <Link
                        className="table-action-link"
                        to={toPsychologistPath(item.psychologistId)}
                      >
                        <IdCard size={15} strokeWidth={2.1} />
                        <span>Карточка</span>
                      </Link>
                      <select
                        className="table-action-select"
                        value={extendDaysById[item.id] || 30}
                        onChange={(event) =>
                          setExtendDaysById((prev) => ({
                            ...prev,
                            [item.id]: Number(event.target.value),
                          }))
                        }
                      >
                        {EXTEND_OPTIONS.map((days) => (
                          <option key={days} value={days}>
                            +{days} дн.
                          </option>
                        ))}
                      </select>
                      <button
                        className="table-action-button"
                        type="button"
                        onClick={() => handleExtend(item.id)}
                      >
                        <CalendarPlus size={15} strokeWidth={2.1} />
                        <span>Продлить</span>
                      </button>
                    </div>
                  </td>
                </tr>
              ))
            ) : (
              <tr>
                <td className="admin-table__empty" colSpan={5}>
                  Подписок по текущему фильтру нет.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </PageCard>
  );
}
