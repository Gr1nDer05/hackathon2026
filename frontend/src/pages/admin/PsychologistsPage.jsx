import { Ban, CalendarPlus, Save, SquareArrowOutUpRight, Undo2, UserPlus, X } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { useAdminData } from "../../modules/admin/model/useAdminData";
import { ROUTES } from "../../shared/config/routes";
import PageCard from "../../shared/ui/PageCard";

const EXTEND_OPTIONS = [7, 14, 30, 60, 90];
const EMPTY_FORM = {
  name: "",
  email: "",
  password: "",
  phone: "",
  city: "",
  specialization: "",
  subscriptionPlan: "basic",
  subscriptionDays: 30,
};

const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const RUSSIAN_PHONE_PATTERN = /^\+7 \d{3} \d{3}-\d{2}-\d{2}$/;
const CYRILLIC_TEXT_PATTERN = /^[А-ЯЁа-яё -]+$/;

function normalizeSpaces(value) {
  return value.trim().replace(/\s+/g, " ");
}

function isValidRussianFio(value) {
  const normalizedValue = normalizeSpaces(value);
  const parts = normalizedValue.split(" ");

  if (parts.length !== 3) {
    return false;
  }

  return parts.every((part) => /^[А-ЯЁа-яё]+(?:-[А-ЯЁа-яё]+)*$/.test(part));
}

function isValidRussianCity(value) {
  const normalizedValue = normalizeSpaces(value);

  if (normalizedValue.length < 2) {
    return false;
  }

  return /^[А-ЯЁа-яё]+(?:[ -][А-ЯЁа-яё]+)*$/.test(normalizedValue);
}

function getRussianPhoneDigits(value) {
  const digits = value.replace(/\D/g, "");

  if (!digits) return "";

  if (digits[0] === "7" || digits[0] === "8") {
    return digits.slice(1, 11);
  }

  return digits.slice(0, 10);
}

function formatRussianPhoneInput(value) {
  const digits = getRussianPhoneDigits(value);

  if (!digits) return "";

  let result = "+7";

  if (digits.length > 0) {
    result += ` ${digits.slice(0, 3)}`;
  }

  if (digits.length > 3) {
    result += ` ${digits.slice(3, 6)}`;
  }

  if (digits.length > 6) {
    result += `-${digits.slice(6, 8)}`;
  }

  if (digits.length > 8) {
    result += `-${digits.slice(8, 10)}`;
  }

  return result;
}

function normalizeRussianPhone(value) {
  const nationalNumber = getRussianPhoneDigits(value);

  if (!/^[3-9]\d{9}$/.test(nationalNumber)) {
    return "";
  }

  return `+7 ${nationalNumber.slice(0, 3)} ${nationalNumber.slice(3, 6)}-${nationalNumber.slice(6, 8)}-${nationalNumber.slice(8, 10)}`;
}

function validatePsychologistForm(form, psychologists) {
  const errors = {};
  const normalizedEmail = form.email.trim().toLowerCase();
  const normalizedDays = Number(form.subscriptionDays);
  const normalizedPhone = normalizeRussianPhone(form.phone.trim());
  const normalizedName = normalizeSpaces(form.name);
  const normalizedCity = normalizeSpaces(form.city);

  if (!normalizedName) {
    errors.name = "Укажите ФИО.";
  } else if (!CYRILLIC_TEXT_PATTERN.test(normalizedName)) {
    errors.name = "ФИО должно быть на русском.";
  } else if (!isValidRussianFio(normalizedName)) {
    errors.name = "Введите полное ФИО в формате: Иванов Иван Иванович.";
  }

  if (!normalizedEmail) {
    errors.email = "Укажите email.";
  } else if (!EMAIL_PATTERN.test(normalizedEmail)) {
    errors.email = "Введите корректный email.";
  } else if (psychologists.some((item) => item.email.toLowerCase() === normalizedEmail)) {
    errors.email = "Психолог с таким email уже существует.";
  }

  if (!form.password.trim()) {
    errors.password = "Укажите пароль.";
  } else if (form.password.length < 8) {
    errors.password = "Пароль должен быть не короче 8 символов.";
  } else if (!/[A-Za-zА-Яа-я]/.test(form.password) || !/\d/.test(form.password)) {
    errors.password = "В пароле нужны буквы и цифры.";
  }

  if (!form.phone.trim()) {
    errors.phone = "Укажите телефон.";
  } else if (!RUSSIAN_PHONE_PATTERN.test(form.phone.trim()) || !normalizedPhone) {
    errors.phone = "Введите номер в формате +7 999 123-45-67.";
  }

  if (!normalizedCity) {
    errors.city = "Укажите город.";
  } else if (!CYRILLIC_TEXT_PATTERN.test(normalizedCity)) {
    errors.city = "Город должен быть указан на русском.";
  } else if (!isValidRussianCity(normalizedCity)) {
    errors.city = "Введите корректное название города на русском.";
  }

  if (form.specialization.trim().length < 4) {
    errors.specialization = "Укажите специализацию подробнее.";
  }

  if (!Number.isInteger(normalizedDays) || normalizedDays < 1 || normalizedDays > 365) {
    errors.subscriptionDays = "Введите от 1 до 365 дней.";
  }

  return errors;
}

function toDetailsPath(id) {
  return ROUTES.adminPsychologistById.replace(":id", id);
}

export default function PsychologistsPage() {
  const {
    psychologists,
    addPsychologist,
    togglePsychologistStatus,
    extendPsychologistSubscription,
    isLoading,
    error,
  } = useAdminData();
  const [query, setQuery] = useState("");
  const [accountFilter, setAccountFilter] = useState("all");
  const [subscriptionFilter, setSubscriptionFilter] = useState("all");
  const [extendDaysById, setExtendDaysById] = useState({});
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [form, setForm] = useState(EMPTY_FORM);
  const [formErrors, setFormErrors] = useState({});
  const [submitError, setSubmitError] = useState("");
  const [actionError, setActionError] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    setExtendDaysById((prev) => {
      const next = { ...prev };

      psychologists.forEach((item) => {
        if (!next[item.id]) {
          next[item.id] = 30;
        }
      });

      return next;
    });
  }, [psychologists]);

  const filteredItems = useMemo(() => {
    const normalizedQuery = query.trim().toLowerCase();

    return psychologists.filter((item) => {
      const matchesText =
        !normalizedQuery ||
        item.name.toLowerCase().includes(normalizedQuery) ||
        item.email.toLowerCase().includes(normalizedQuery);

      const matchesAccount = accountFilter === "all" || item.accountStatus === accountFilter;
      const matchesSubscription =
        subscriptionFilter === "all" || item.subscriptionStatus === subscriptionFilter;

      return matchesText && matchesAccount && matchesSubscription;
    });
  }, [psychologists, query, accountFilter, subscriptionFilter]);

  function handleFormChange(field, value) {
    const nextValue = field === "phone" ? formatRussianPhoneInput(value) : value;

    setForm((prev) => ({ ...prev, [field]: nextValue }));
    setSubmitError("");
    setFormErrors((prev) => {
      if (!prev[field]) return prev;

      const next = { ...prev };
      delete next[field];
      return next;
    });
  }

  async function handleCreatePsychologist(event) {
    event.preventDefault();
    const nextErrors = validatePsychologistForm(form, psychologists);

    if (Object.keys(nextErrors).length > 0) {
      setFormErrors(nextErrors);
      return;
    }

    setIsSubmitting(true);
    setSubmitError("");

    try {
        await addPsychologist({
          ...form,
          name: normalizeSpaces(form.name),
          email: form.email.trim().toLowerCase(),
          password: form.password.trim(),
          phone: normalizeRussianPhone(form.phone.trim()),
          city: normalizeSpaces(form.city),
          specialization: form.specialization.trim(),
          subscriptionPlan: form.subscriptionPlan,
          subscriptionDays: Number(form.subscriptionDays) || 30,
        });

      setForm(EMPTY_FORM);
      setFormErrors({});
      setIsCreateOpen(false);
    } catch (requestError) {
      setSubmitError(
        requestError?.message ||
          "Не удалось создать психолога. Проверьте данные и повторите попытку.",
      );
    } finally {
      setIsSubmitting(false);
    }
  }

  async function handleToggleBlock(id) {
    setActionError("");

    try {
      await togglePsychologistStatus(id);
    } catch (requestError) {
      setActionError(
        requestError?.message || "Не удалось обновить статус аккаунта. Повторите попытку.",
      );
    }
  }

  async function handleExtendSubscription(id) {
    const days = extendDaysById[id] || 30;
    setActionError("");

    try {
      await extendPsychologistSubscription(id, days);
    } catch (requestError) {
      setActionError(
        requestError?.message || "Не удалось продлить доступ. Повторите попытку.",
      );
    }
  }

  return (
    <PageCard
      embedded
      title="Психологи"
      description="Список психологов с быстрыми действиями администратора."
      links={[
        { to: ROUTES.adminDashboard, label: "Назад в админ-панель" },
        { to: ROUTES.adminSubscriptions, label: "Подписки" },
      ]}
    >
      <section className="admin-page-actions">
        <button
          className="admin-primary-button"
          type="button"
          onClick={() => setIsCreateOpen((prev) => !prev)}
        >
          {isCreateOpen ? <X size={16} strokeWidth={2.1} /> : <UserPlus size={16} strokeWidth={2.1} />}
          <span>{isCreateOpen ? "Скрыть форму" : "Добавить психолога"}</span>
        </button>
      </section>

      {isCreateOpen ? (
        <section className="admin-create-panel">
          <div className="admin-create-panel__header">
            <div>
              <p className="admin-create-panel__eyebrow">Создание аккаунта</p>
              <h2 className="admin-create-panel__title">Новый психолог</h2>
            </div>
            <p className="admin-create-panel__meta">
              Администратор создаёт аккаунт и сразу задаёт стартовые параметры подписки.
            </p>
          </div>

          <form className="admin-form-grid" onSubmit={handleCreatePsychologist}>
            <label className="admin-form-field">
              <span>ФИО</span>
              <input
                aria-invalid={Boolean(formErrors.name)}
                className={formErrors.name ? "admin-form-control admin-form-control--invalid" : "admin-form-control"}
                value={form.name}
                onChange={(event) => handleFormChange("name", event.target.value)}
                placeholder="Иванов Иван Иванович"
              />
              {formErrors.name ? <small className="admin-form-error admin-form-error--inline">{formErrors.name}</small> : null}
            </label>

            <label className="admin-form-field">
              <span>Email</span>
              <input
                type="email"
                aria-invalid={Boolean(formErrors.email)}
                className={formErrors.email ? "admin-form-control admin-form-control--invalid" : "admin-form-control"}
                value={form.email}
                onChange={(event) => handleFormChange("email", event.target.value)}
                placeholder="anna@example.com"
              />
              {formErrors.email ? <small className="admin-form-error admin-form-error--inline">{formErrors.email}</small> : null}
            </label>

            <label className="admin-form-field">
              <span>Пароль</span>
              <input
                type="password"
                aria-invalid={Boolean(formErrors.password)}
                className={formErrors.password ? "admin-form-control admin-form-control--invalid" : "admin-form-control"}
                value={form.password}
                onChange={(event) => handleFormChange("password", event.target.value)}
                placeholder="Минимум 8 символов"
              />
              {formErrors.password ? <small className="admin-form-error admin-form-error--inline">{formErrors.password}</small> : null}
            </label>

            <label className="admin-form-field">
              <span>Телефон</span>
              <input
                type="tel"
                inputMode="tel"
                maxLength={16}
                aria-invalid={Boolean(formErrors.phone)}
                className={formErrors.phone ? "admin-form-control admin-form-control--invalid" : "admin-form-control"}
                value={form.phone}
                onChange={(event) => handleFormChange("phone", event.target.value)}
                placeholder="+7 999 000-00-00"
              />
              {formErrors.phone ? <small className="admin-form-error admin-form-error--inline">{formErrors.phone}</small> : null}
            </label>

            <label className="admin-form-field">
              <span>Город</span>
              <input
                aria-invalid={Boolean(formErrors.city)}
                className={formErrors.city ? "admin-form-control admin-form-control--invalid" : "admin-form-control"}
                value={form.city}
                onChange={(event) => handleFormChange("city", event.target.value)}
                placeholder="Екатеринбург"
              />
              {formErrors.city ? <small className="admin-form-error admin-form-error--inline">{formErrors.city}</small> : null}
            </label>

            <label className="admin-form-field admin-form-field--wide">
              <span>Специализация</span>
              <input
                aria-invalid={Boolean(formErrors.specialization)}
                className={formErrors.specialization ? "admin-form-control admin-form-control--invalid" : "admin-form-control"}
                value={form.specialization}
                onChange={(event) => handleFormChange("specialization", event.target.value)}
                placeholder="Профориентация старшеклассников"
              />
              {formErrors.specialization ? (
                <small className="admin-form-error admin-form-error--inline">
                  {formErrors.specialization}
                </small>
              ) : null}
            </label>

            <label className="admin-form-field">
              <span>План</span>
              <select
                className="admin-form-control"
                value={form.subscriptionPlan}
                onChange={(event) => handleFormChange("subscriptionPlan", event.target.value)}
              >
                <option value="basic">Basic</option>
                <option value="pro">Pro</option>
              </select>
            </label>

            <label className="admin-form-field">
              <span>Дней подписки</span>
              <input
                type="number"
                min="1"
                max="365"
                aria-invalid={Boolean(formErrors.subscriptionDays)}
                className={
                  formErrors.subscriptionDays
                    ? "admin-form-control admin-form-control--invalid"
                    : "admin-form-control"
                }
                value={form.subscriptionDays}
                onChange={(event) => handleFormChange("subscriptionDays", event.target.value)}
              />
              {formErrors.subscriptionDays ? (
                <small className="admin-form-error admin-form-error--inline">
                  {formErrors.subscriptionDays}
                </small>
              ) : null}
            </label>

            <div className="admin-form-actions">
              <button className="admin-primary-button" disabled={isSubmitting} type="submit">
                <Save size={16} strokeWidth={2.1} />
                <span>{isSubmitting ? "Создание..." : "Создать психолога"}</span>
              </button>
            </div>
            {submitError ? (
              <p className="admin-form-message admin-form-message--error">{submitError}</p>
            ) : null}
          </form>
        </section>
      ) : null}

      {error ? (
        <p className="admin-form-message admin-form-message--error">
          {error.message || "Не удалось загрузить список психологов."}
        </p>
      ) : null}

      {actionError ? (
        <p className="admin-form-message admin-form-message--error">{actionError}</p>
      ) : null}

      <section className="admin-tools admin-tools--compact">
        <input
          className="admin-tools__input"
          type="text"
          value={query}
          onChange={(event) => setQuery(event.target.value)}
          placeholder="Поиск по имени или email"
        />

        <select
          className="admin-tools__select"
          value={accountFilter}
          onChange={(event) => setAccountFilter(event.target.value)}
        >
          <option value="all">Все аккаунты</option>
          <option value="active">Активные</option>
          <option value="blocked">Заблокированные</option>
        </select>

        <select
          className="admin-tools__select"
          value={subscriptionFilter}
          onChange={(event) => setSubscriptionFilter(event.target.value)}
        >
          <option value="all">Все подписки</option>
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
              <th>Фокус</th>
              <th>Аккаунт</th>
              <th>Подписка</th>
              <th>Истекает</th>
              <th>Действия</th>
            </tr>
          </thead>
          <tbody>
            {isLoading ? (
              <tr>
                <td className="admin-table__empty" colSpan={6}>
                  Загружаем список психологов...
                </td>
              </tr>
            ) : filteredItems.length ? (
              filteredItems.map((item) => (
                <tr key={item.id}>
                  <td>
                    <p className="admin-table__primary">{item.name}</p>
                    <p className="admin-table__secondary">{item.email}</p>
                  </td>
                  <td>
                    <p className="admin-table__primary admin-table__primary--small">
                      {item.specialization}
                    </p>
                    <p className="admin-table__secondary">
                      {item.city} · {item.testsCount} тестов
                    </p>
                  </td>
                  <td>
                    <span className={`status-badge status-badge--${item.accountStatus}`}>
                      {item.accountStatus}
                    </span>
                  </td>
                  <td>
                    <span className={`status-badge status-badge--${item.subscriptionStatus}`}>
                      {item.subscriptionStatus}
                    </span>
                  </td>
                  <td>{item.expiresAt}</td>
                  <td>
                    <div className="admin-table__actions">
                      <Link className="table-action-link" to={toDetailsPath(item.id)}>
                        <SquareArrowOutUpRight size={15} strokeWidth={2.1} />
                        <span>Открыть</span>
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
                        onClick={() => handleToggleBlock(item.id)}
                      >
                        {item.accountStatus === "blocked" ? (
                          <Undo2 size={15} strokeWidth={2.1} />
                        ) : (
                          <Ban size={15} strokeWidth={2.1} />
                        )}
                        <span>
                          {item.accountStatus === "blocked" ? "Разблокировать" : "Блокировать"}
                        </span>
                      </button>
                      <button
                        className="table-action-button"
                        type="button"
                        onClick={() => handleExtendSubscription(item.id)}
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
                <td className="admin-table__empty" colSpan={6}>
                  {psychologists.length
                    ? "Ничего не найдено. Измени фильтры или скорректируй запрос."
                    : "Психологов пока нет. Создай первый аккаунт через форму выше."}
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </PageCard>
  );
}
