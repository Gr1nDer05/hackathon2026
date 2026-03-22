import {
  CreditCard,
  LoaderCircle,
  ListChecks,
  LogOut,
  Mail,
  MapPin,
  Phone,
  Save,
  Sparkles,
  UserCircle2,
} from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import useSWR from "swr";
import { useSWRConfig } from "swr";
import { Link } from "react-router-dom";
import { useAuth } from "../../modules/auth/model/useAuth";
import {
  updatePsychologistCardRequest,
  updatePsychologistProfileRequest,
} from "../../modules/psychologist/api/psychologistApi";
import {
  formatDate,
  getAccountStatus,
  getAccountStatusLabel,
  getPsychologistAccessUntil,
  getPsychologistCity,
  getPsychologistDisplayName,
  getPsychologistEducation,
  getPsychologistInitials,
  getPsychologistJoinedAt,
  getPsychologistLastActivityAt,
  getPsychologistPhone,
  getPsychologistSpecialization,
  getStatusTone,
  getSubscriptionPlan,
  getSubscriptionPlanLabel,
  getSubscriptionStatus,
  getSubscriptionStatusLabel,
  hasAiTemplateAccess,
  sortTestsByRecent,
  summarizeTests,
} from "../../modules/psychologist/lib/psychologistUi";
import { listPsychologistTestsRequest } from "../../modules/tests/api/testsApi";
import { ROUTES } from "../../shared/config/routes";
import {
  formatRussianPhoneInput,
  isRussianPhoneComplete,
} from "../../shared/lib/russianPhone";
import PageCard from "../../shared/ui/PageCard";

function toResultsPath(id) {
  return ROUTES.testResults.replace(":id", String(id));
}

function createProfileForm(profile = {}) {
  return {
    about: String(profile.about || ""),
    specialization: String(profile.specialization || ""),
    experience_years:
      profile.experience_years === null ||
      profile.experience_years === undefined
        ? ""
        : String(profile.experience_years),
    education: String(profile.education || ""),
    methods: String(profile.methods || ""),
    city: String(profile.city || ""),
    timezone: String(profile.timezone || ""),
    is_public: Boolean(profile.is_public),
  };
}

function createCardForm(card = {}, user = {}) {
  return {
    headline: String(card.headline || ""),
    short_bio: String(card.short_bio || ""),
    contact_email: String(card.contact_email || user.email || ""),
    contact_phone: card.contact_phone
      ? formatRussianPhoneInput(card.contact_phone)
      : "",
    telegram: String(card.telegram || ""),
    online_available: Boolean(card.online_available),
    offline_available: Boolean(card.offline_available),
  };
}

function validateProfileForm(form) {
  const errors = {};

  if (form.specialization && String(form.specialization).trim().length < 4) {
    errors.specialization = "Специализация должна содержать минимум 4 символа.";
  }

  if (form.city && !/^[А-ЯЁа-яё -]+$/.test(String(form.city).trim())) {
    errors.city = "Город должен быть указан на русском.";
  }

  if (String(form.experience_years || "").trim()) {
    const years = Number(form.experience_years);
    if (!Number.isInteger(years) || years < 0) {
      errors.experience_years = "Стаж должен быть целым числом от 0.";
    }
  }

  return errors;
}

function validateCardForm(form) {
  const errors = {};
  const email = String(form.contact_email || "").trim();
  const phone = String(form.contact_phone || "").trim();

  if (email && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
    errors.contact_email = "Укажите корректный email.";
  }

  if (phone && !isRussianPhoneComplete(phone)) {
    errors.contact_phone = "Телефон должен быть в формате +7 XXX XXX-XX-XX.";
  }

  return errors;
}

function buildProfilePayload(form) {
  const payload = {
    about: String(form.about || "").trim(),
    specialization: String(form.specialization || "").trim(),
    education: String(form.education || "").trim(),
    methods: String(form.methods || "").trim(),
    city: String(form.city || "").trim(),
    timezone: String(form.timezone || "").trim(),
    is_public: Boolean(form.is_public),
  };

  const years = String(form.experience_years || "").trim();
  if (years) {
    payload.experience_years = Number(years);
  }

  return payload;
}

function buildCardPayload(form) {
  const payload = {
    headline: String(form.headline || "").trim(),
    short_bio: String(form.short_bio || "").trim(),
    contact_email: String(form.contact_email || "").trim(),
    telegram: String(form.telegram || "").trim(),
    online_available: Boolean(form.online_available),
    offline_available: Boolean(form.offline_available),
  };

  const phone = String(form.contact_phone || "").trim();
  if (phone) {
    payload.contact_phone = phone;
  }

  return payload;
}

export default function ProfilePage() {
  const {
    user,
    signOut,
    isSigningOut,
    createSubscriptionPurchaseRequest,
    isCreatingSubscriptionPurchaseRequest,
    refreshSession,
  } = useAuth();
  const { mutate } = useSWRConfig();
  const testsQuery = useSWR(
    "psychologist-tests",
    listPsychologistTestsRequest,
    {
      revalidateOnFocus: false,
      shouldRetryOnError: false,
    },
  );

  const [profileForm, setProfileForm] = useState(createProfileForm());
  const [cardForm, setCardForm] = useState(createCardForm());
  const [profileErrors, setProfileErrors] = useState({});
  const [cardErrors, setCardErrors] = useState({});
  const [feedbackMessage, setFeedbackMessage] = useState("");
  const [feedbackError, setFeedbackError] = useState("");
  const [isSavingProfile, setIsSavingProfile] = useState(false);
  const [isSavingCard, setIsSavingCard] = useState(false);
  const [requestedPlan, setRequestedPlan] = useState("basic");
  const [isRefreshingAccess, setIsRefreshingAccess] = useState(false);

  useEffect(() => {
    setProfileForm(createProfileForm(user?.profile));
    setCardForm(createCardForm(user?.card, user));
  }, [user]);

  useEffect(() => {
    setRequestedPlan(getSubscriptionPlan(user));
  }, [user]);

  useEffect(() => {
    if (!feedbackMessage) {
      return undefined;
    }

    const timeoutId = window.setTimeout(() => setFeedbackMessage(""), 2600);
    return () => window.clearTimeout(timeoutId);
  }, [feedbackMessage]);

  const tests = testsQuery.data || user?.tests || user?.workspace?.tests || [];
  const recentTests = useMemo(() => sortTestsByRecent(tests), [tests]);
  const stats = useMemo(() => summarizeTests(tests), [tests]);

  const displayName = getPsychologistDisplayName(user);
  const initials = getPsychologistInitials(user);
  const specialization = getPsychologistSpecialization(user);
  const phone = getPsychologistPhone(user);
  const city = getPsychologistCity(user);
  const education = getPsychologistEducation(user);
  const joinedAt = formatDate(getPsychologistJoinedAt(user));
  const lastActivity = formatDate(
    getPsychologistLastActivityAt(user, recentTests),
  );
  const accessUntil = formatDate(getPsychologistAccessUntil(user));
  const accountStatus = getAccountStatus(user);
  const subscriptionStatus = getSubscriptionStatus(user);
  const subscriptionPlan = getSubscriptionPlan(user);
  const experienceYears = user?.profile?.experience_years ?? null;
  const about = user?.profile?.about || "";
  const aboutLength = profileForm.about.trim().length;
  const timezone = user?.profile?.timezone || "Не указано";
  const publicVisibility = user?.profile?.is_public ? "Публичный" : "Скрыт";
  const cardHeadline = user?.card?.headline || "Заголовок карточки не заполнен";
  const workFormat =
    [
      user?.card?.online_available ? "Онлайн" : "",
      user?.card?.offline_available ? "Офлайн" : "",
    ]
      .filter(Boolean)
      .join(" / ") || "Не указан";
  const recentTestTitle = recentTests[0]?.title || "Пока нет созданных тестов";
  const latestResultsPath = recentTests[0]?.id
    ? toResultsPath(recentTests[0].id)
    : ROUTES.tests;

  function patchSession(patch) {
    mutate(
      "auth-session",
      (current) => {
        if (!current) {
          return current;
        }

        return {
          ...current,
          ...patch,
          profile: patch.profile || current.profile,
          card: patch.card || current.card,
          workspace: current.workspace
            ? {
                ...current.workspace,
                profile: patch.profile || current.workspace.profile,
                card: patch.card || current.workspace.card,
              }
            : current.workspace,
        };
      },
      false,
    );
  }

  function handleProfileChange(field, value) {
    setProfileForm((prev) => ({ ...prev, [field]: value }));
    setFeedbackError("");
    setProfileErrors((prev) => {
      if (!prev[field]) {
        return prev;
      }

      const next = { ...prev };
      delete next[field];
      return next;
    });
  }

  function handleCardChange(field, value) {
    const nextValue =
      field === "contact_phone" ? formatRussianPhoneInput(value) : value;
    setCardForm((prev) => ({ ...prev, [field]: nextValue }));
    setFeedbackError("");
    setCardErrors((prev) => {
      if (!prev[field]) {
        return prev;
      }

      const next = { ...prev };
      delete next[field];
      return next;
    });
  }

  async function handleSaveProfile(event) {
    event.preventDefault();
    const errors = validateProfileForm(profileForm);
    setProfileErrors(errors);

    if (Object.keys(errors).length) {
      return;
    }

    setIsSavingProfile(true);
    setFeedbackError("");

    try {
      const updatedProfile = await updatePsychologistProfileRequest(
        buildProfilePayload(profileForm),
      );
      patchSession({ profile: updatedProfile });
      setFeedbackMessage("Профиль сохранён.");
    } catch (error) {
      setFeedbackError(error?.message || "Не удалось сохранить профиль.");
    } finally {
      setIsSavingProfile(false);
    }
  }

  async function handleSaveCard(event) {
    event.preventDefault();
    const errors = validateCardForm(cardForm);
    setCardErrors(errors);

    if (Object.keys(errors).length) {
      return;
    }

    setIsSavingCard(true);
    setFeedbackError("");

    try {
      const updatedCard = await updatePsychologistCardRequest(
        buildCardPayload(cardForm),
      );
      patchSession({ card: updatedCard });
      setFeedbackMessage("Карточка специалиста сохранена.");
    } catch (error) {
      setFeedbackError(error?.message || "Не удалось сохранить карточку.");
    } finally {
      setIsSavingCard(false);
    }
  }

  return (
    <PageCard
      wide
      title="Профиль психолога"
      description="Редактирование рабочего профиля, карточки специалиста и параметров отображения."
      links={[
        { to: ROUTES.dashboard, label: "В кабинет" },
        { to: ROUTES.tests, label: "Мои тесты" },
      ]}
    >
      <section className="psychologist-profile-hero">
        <div className="psychologist-profile-hero__identity">
          <div className="psychologist-profile-hero__avatar">{initials}</div>
          <div>
            <p className="psychologist-profile-hero__eyebrow">
              Профиль специалиста
            </p>
            <h2 className="psychologist-profile-hero__name">{displayName}</h2>
            <p className="psychologist-profile-hero__role">{specialization}</p>
          </div>
        </div>

        <div className="psychologist-profile-hero__badges">
          <span
            className={`status-badge status-badge--${getStatusTone(accountStatus)}`}
          >
            {getAccountStatusLabel(accountStatus)}
          </span>
          <span
            className={`status-badge status-badge--${getStatusTone(subscriptionStatus)}`}
          >
            {getSubscriptionStatusLabel(subscriptionStatus)}
          </span>
          <span
            className={`status-badge status-badge--${subscriptionPlan === "pro" ? "active" : "draft"}`}
          >
            {getSubscriptionPlanLabel(subscriptionPlan)}
          </span>
        </div>
      </section>

      {testsQuery.error ? (
        <p className="admin-form-message admin-form-message--error">
          {testsQuery.error.message ||
            "Не удалось загрузить данные по методикам."}
        </p>
      ) : null}
      {feedbackError ? (
        <p className="admin-form-message admin-form-message--error">
          {feedbackError}
        </p>
      ) : null}
      {feedbackMessage ? (
        <p className="admin-form-message">{feedbackMessage}</p>
      ) : null}

      <div className={`workflow-note ${hasAiTemplateAccess(user) ? "workflow-note--success" : "workflow-note--warning"}`}>
        <p>
          {hasAiTemplateAccess(user)
            ? "На текущем плане доступно автоматическое заполнение шаблонов отчётов."
            : "Автоматическое заполнение шаблонов сейчас недоступно на вашем плане."}
        </p>
        <div className="workflow-note__actions">
          <Link className="table-action-link" to={ROUTES.reportTemplates}>
            <Sparkles size={15} strokeWidth={2.1} />
            <span>Шаблоны отчётов</span>
          </Link>
        </div>
      </div>

      <div className={`workflow-note ${subscriptionStatus === "active" ? "workflow-note--success" : "workflow-note--warning"}`}>
        <p>
          Текущий план: <strong>{getSubscriptionPlanLabel(subscriptionPlan)}</strong>. Можно отправить заявку на продление или смену плана.
        </p>
        <div className="workflow-note__actions">
          <select
            className="table-action-select"
            value={requestedPlan}
            onChange={(event) => setRequestedPlan(event.target.value)}
            disabled={isCreatingSubscriptionPurchaseRequest}
          >
            <option value="basic">Basic</option>
            <option value="pro">Pro</option>
          </select>
          <button
            className="table-action-button"
            type="button"
            disabled={isCreatingSubscriptionPurchaseRequest}
            onClick={async () => {
              setFeedbackError("");

              try {
                await createSubscriptionPurchaseRequest({
                  subscriptionPlan: requestedPlan,
                });
                setFeedbackMessage(
                  `Заявка на план ${getSubscriptionPlanLabel(requestedPlan)} отправлена администратору.`,
                );
              } catch (error) {
                setFeedbackError(
                  error?.message || "Не удалось отправить заявку на подписку.",
                );
              }
            }}
          >
            {isCreatingSubscriptionPurchaseRequest ? (
              <LoaderCircle className="icon-spin" size={15} strokeWidth={2.1} />
            ) : (
              <CreditCard size={15} strokeWidth={2.1} />
            )}
            <span>{isCreatingSubscriptionPurchaseRequest ? "Отправляем..." : "Оставить заявку"}</span>
          </button>
          <button
            className="table-action-button"
            type="button"
            disabled={isRefreshingAccess}
            onClick={async () => {
              setIsRefreshingAccess(true);

              try {
                await refreshSession();
              } finally {
                setIsRefreshingAccess(false);
              }
            }}
          >
            {isRefreshingAccess ? (
              <LoaderCircle className="icon-spin" size={15} strokeWidth={2.1} />
            ) : (
              <Sparkles size={15} strokeWidth={2.1} />
            )}
            <span>Обновить статус</span>
          </button>
        </div>
      </div>

      <section className="psychologist-kpis">
        <article className="psychologist-kpi">
          <p className="psychologist-kpi__label">Опросников</p>
          <p className="psychologist-kpi__value">{stats.totalCount}</p>
        </article>
        <article className="psychologist-kpi">
          <p className="psychologist-kpi__label">Опубликовано</p>
          <p className="psychologist-kpi__value">{stats.publishedCount}</p>
        </article>
        <article className="psychologist-kpi">
          <p className="psychologist-kpi__label">Черновиков</p>
          <p className="psychologist-kpi__value">{stats.draftCount}</p>
        </article>
        <article className="psychologist-kpi">
          <p className="psychologist-kpi__label">Прохождений</p>
          <p className="psychologist-kpi__value">{stats.completedSessions}</p>
        </article>
      </section>

      <section className="psychologist-profile-grid">
        <article className="admin-panel">
          <h3 className="admin-panel__title">Контакты</h3>
          <dl className="profile-meta-list">
            <div>
              <dt>Email</dt>
              <dd className="psychologist-meta-row">
                <Mail size={15} strokeWidth={2.1} />{" "}
                <span>{user?.email || "—"}</span>
              </dd>
            </div>
            <div>
              <dt>Телефон</dt>
              <dd className="psychologist-meta-row">
                <Phone size={15} strokeWidth={2.1} /> <span>{phone}</span>
              </dd>
            </div>
            <div>
              <dt>Город</dt>
              <dd className="psychologist-meta-row">
                <MapPin size={15} strokeWidth={2.1} /> <span>{city}</span>
              </dd>
            </div>
            <div>
              <dt>Таймзона</dt>
              <dd>{timezone}</dd>
            </div>
          </dl>
        </article>

        <article className="admin-panel">
          <h3 className="admin-panel__title">Профессиональный профиль</h3>
          <dl className="profile-meta-list">
            <div>
              <dt>Стаж</dt>
              <dd>
                {experienceYears ?? "Не указан"}{" "}
                {experienceYears !== null ? "лет" : ""}
              </dd>
            </div>
            <div>
              <dt>Образование</dt>
              <dd>{education}</dd>
            </div>
            <div>
              <dt>Видимость профиля</dt>
              <dd>{publicVisibility}</dd>
            </div>
            <div>
              <dt>План</dt>
              <dd>{getSubscriptionPlanLabel(subscriptionPlan)}</dd>
            </div>
          </dl>
        </article>

        <article className="admin-panel">
          <h3 className="admin-panel__title">Карточка для клиента</h3>
          <dl className="profile-meta-list">
            <div>
              <dt>Заголовок</dt>
              <dd>{cardHeadline}</dd>
            </div>
            <div>
              <dt>Формат работы</dt>
              <dd>{workFormat}</dd>
            </div>
            <div>
              <dt>Telegram</dt>
              <dd>{user?.card?.telegram || "Не указан"}</dd>
            </div>
            <div>
              <dt>Портал доступен до</dt>
              <dd>{accessUntil}</dd>
            </div>
          </dl>
        </article>

        <article className="admin-panel">
          <h3 className="admin-panel__title">Активность</h3>
          <dl className="profile-meta-list">
            <div>
              <dt>Дата подключения</dt>
              <dd>{joinedAt}</dd>
            </div>
            <div>
              <dt>Последняя активность</dt>
              <dd>{lastActivity}</dd>
            </div>
            <div>
              <dt>Последняя методика</dt>
              <dd>{recentTestTitle}</dd>
            </div>
          </dl>
        </article>

        <article className="admin-panel psychologist-profile-panel psychologist-profile-panel--about">
          <h3 className="admin-panel__title">О специалисте</h3>
          <div className="psychologist-profile-text psychologist-profile-text--panel">
            <p className="psychologist-profile-text__body">
              {about || "Описание специалиста пока не заполнено."}
            </p>
          </div>
        </article>
      </section>

      <section className="psychologist-edit-grid">
        <article className="admin-panel admin-panel--spaced">
          <h3 className="admin-panel__title">Редактирование профиля</h3>
          <form className="admin-form-grid" onSubmit={handleSaveProfile}>
            <div className="admin-form-field admin-form-field--wide psychologist-bio-field">
              <div className="psychologist-bio-field__head">
                <div>
                  <span>О специалисте</span>
                  <p
                    className="psychologist-bio-field__hint"
                    id="profile-about-hint"
                  >
                    Коротко опиши подход, с кем работаешь и какой результат
                    получает клиент после консультации.
                  </p>
                </div>
                <span
                  className={`psychologist-bio-field__counter ${
                    aboutLength >= 180 && aboutLength <= 450
                      ? "psychologist-bio-field__counter--good"
                      : ""
                  }`}
                  id="profile-about-count"
                >
                  {aboutLength} симв.
                </span>
              </div>
              <textarea
                aria-describedby="profile-about-hint profile-about-count"
                className="admin-form-control builder-textarea psychologist-bio-field__textarea"
                rows={6}
                value={profileForm.about}
                onChange={(event) =>
                  handleProfileChange("about", event.target.value)
                }
              />
              <div className="psychologist-bio-field__tips" aria-hidden="true">
                <span>подход</span>
                <span>целевая аудитория</span>
                <span>формат работы</span>
              </div>
            </div>

            <label className="admin-form-field">
              <span>Специализация</span>
              <input
                aria-invalid={Boolean(profileErrors.specialization)}
                className={
                  profileErrors.specialization
                    ? "admin-form-control admin-form-control--invalid"
                    : "admin-form-control"
                }
                value={profileForm.specialization}
                onChange={(event) =>
                  handleProfileChange("specialization", event.target.value)
                }
              />
              {profileErrors.specialization ? (
                <small className="admin-form-error admin-form-error--inline">
                  {profileErrors.specialization}
                </small>
              ) : null}
            </label>

            <label className="admin-form-field">
              <span>Стаж, лет</span>
              <input
                aria-invalid={Boolean(profileErrors.experience_years)}
                className={
                  profileErrors.experience_years
                    ? "admin-form-control admin-form-control--invalid"
                    : "admin-form-control"
                }
                inputMode="numeric"
                type="number"
                value={profileForm.experience_years}
                onChange={(event) =>
                  handleProfileChange("experience_years", event.target.value)
                }
              />
              {profileErrors.experience_years ? (
                <small className="admin-form-error admin-form-error--inline">
                  {profileErrors.experience_years}
                </small>
              ) : null}
            </label>

            <label className="admin-form-field">
              <span>Образование</span>
              <input
                className="admin-form-control"
                value={profileForm.education}
                onChange={(event) =>
                  handleProfileChange("education", event.target.value)
                }
              />
            </label>

            <label className="admin-form-field">
              <span>Город</span>
              <input
                aria-invalid={Boolean(profileErrors.city)}
                className={
                  profileErrors.city
                    ? "admin-form-control admin-form-control--invalid"
                    : "admin-form-control"
                }
                value={profileForm.city}
                onChange={(event) =>
                  handleProfileChange("city", event.target.value)
                }
                placeholder="Москва"
              />
              {profileErrors.city ? (
                <small className="admin-form-error admin-form-error--inline">
                  {profileErrors.city}
                </small>
              ) : null}
            </label>

            <label className="admin-form-field">
              <span>Таймзона</span>
              <input
                className="admin-form-control"
                value={profileForm.timezone}
                onChange={(event) =>
                  handleProfileChange("timezone", event.target.value)
                }
                placeholder="Europe/Moscow"
              />
            </label>

            <label className="admin-form-field admin-form-field--wide">
              <span>Методы работы</span>
              <textarea
                className="admin-form-control builder-textarea"
                rows={3}
                value={profileForm.methods}
                onChange={(event) =>
                  handleProfileChange("methods", event.target.value)
                }
                placeholder="Например: профориентация, диагностика интересов, карьерные консультации."
              />
            </label>

            <label className="builder-toggle">
              <span>Показывать профиль публично</span>
              <input
                checked={profileForm.is_public}
                type="checkbox"
                onChange={(event) =>
                  handleProfileChange("is_public", event.target.checked)
                }
              />
            </label>

            <div className="admin-form-actions">
              <button
                className="admin-primary-button"
                disabled={isSavingProfile}
                type="submit"
              >
                {isSavingProfile ? (
                  <LoaderCircle
                    className="icon-spin"
                    size={16}
                    strokeWidth={2.1}
                  />
                ) : (
                  <Save size={16} strokeWidth={2.1} />
                )}
                <span>
                  {isSavingProfile ? "Сохранение..." : "Сохранить профиль"}
                </span>
              </button>
            </div>
          </form>
        </article>

        <article className="admin-panel admin-panel--spaced">
          <h3 className="admin-panel__title">Карточка специалиста</h3>
          <form className="admin-form-grid" onSubmit={handleSaveCard}>
            <label className="admin-form-field">
              <span>Заголовок карточки</span>
              <input
                className="admin-form-control"
                value={cardForm.headline}
                onChange={(event) =>
                  handleCardChange("headline", event.target.value)
                }
                placeholder="Психолог-профориентолог"
              />
            </label>

            <label className="admin-form-field">
              <span>Контактный email</span>
              <input
                aria-invalid={Boolean(cardErrors.contact_email)}
                className={
                  cardErrors.contact_email
                    ? "admin-form-control admin-form-control--invalid"
                    : "admin-form-control"
                }
                type="email"
                value={cardForm.contact_email}
                onChange={(event) =>
                  handleCardChange("contact_email", event.target.value)
                }
              />
              {cardErrors.contact_email ? (
                <small className="admin-form-error admin-form-error--inline">
                  {cardErrors.contact_email}
                </small>
              ) : null}
            </label>

            <label className="admin-form-field">
              <span>Контактный телефон</span>
              <input
                aria-invalid={Boolean(cardErrors.contact_phone)}
                className={
                  cardErrors.contact_phone
                    ? "admin-form-control admin-form-control--invalid"
                    : "admin-form-control"
                }
                inputMode="tel"
                maxLength={16}
                value={cardForm.contact_phone}
                onChange={(event) =>
                  handleCardChange("contact_phone", event.target.value)
                }
                placeholder="+7 999 123-45-67"
              />
              {cardErrors.contact_phone ? (
                <small className="admin-form-error admin-form-error--inline">
                  {cardErrors.contact_phone}
                </small>
              ) : null}
            </label>

            <label className="admin-form-field">
              <span>Telegram</span>
              <input
                className="admin-form-control"
                value={cardForm.telegram}
                onChange={(event) =>
                  handleCardChange("telegram", event.target.value)
                }
                placeholder="@username"
              />
            </label>

            <label className="admin-form-field admin-form-field--wide">
              <span>Краткое описание</span>
              <textarea
                className="admin-form-control builder-textarea"
                rows={4}
                value={cardForm.short_bio}
                onChange={(event) =>
                  handleCardChange("short_bio", event.target.value)
                }
                placeholder="Короткий текст для карточки клиента."
              />
            </label>

            <div className="builder-toggle-grid admin-form-field--wide">
              <label className="builder-toggle">
                <span>Доступно онлайн</span>
                <input
                  checked={cardForm.online_available}
                  type="checkbox"
                  onChange={(event) =>
                    handleCardChange("online_available", event.target.checked)
                  }
                />
              </label>
              <label className="builder-toggle">
                <span>Доступно офлайн</span>
                <input
                  checked={cardForm.offline_available}
                  type="checkbox"
                  onChange={(event) =>
                    handleCardChange("offline_available", event.target.checked)
                  }
                />
              </label>
            </div>

            <div className="admin-form-actions">
              <button
                className="admin-primary-button"
                disabled={isSavingCard}
                type="submit"
              >
                {isSavingCard ? (
                  <LoaderCircle
                    className="icon-spin"
                    size={16}
                    strokeWidth={2.1}
                  />
                ) : (
                  <Save size={16} strokeWidth={2.1} />
                )}
                <span>
                  {isSavingCard ? "Сохранение..." : "Сохранить карточку"}
                </span>
              </button>
            </div>
          </form>
        </article>
      </section>

      <section className="admin-panel admin-panel--spaced">
        <h3 className="admin-panel__title">Быстрые действия</h3>
        <div className="psychologist-quick-grid">
          <Link className="psychologist-quick-card" to={ROUTES.dashboard}>
            <UserCircle2 size={18} strokeWidth={2.1} />
            <strong>Рабочий кабинет</strong>
            <span>Сводка по активности, последним тестам и доступу.</span>
          </Link>
          <Link className="psychologist-quick-card" to={ROUTES.tests}>
            <Sparkles size={18} strokeWidth={2.1} />
            <strong>Управление тестами</strong>
            <span>Создание, публикация и редактирование методик.</span>
          </Link>
          <Link className="psychologist-quick-card" to={latestResultsPath}>
            <ListChecks size={18} strokeWidth={2.1} />
            <strong>Последние результаты</strong>
            <span>
              {recentTests[0]
                ? `Посмотреть прохождения по «${recentTests[0].title}».`
                : "Результаты появятся после первых прохождений тестов."}
            </span>
          </Link>
          <button
            className="psychologist-quick-card psychologist-quick-card--button"
            type="button"
            onClick={signOut}
            disabled={isSigningOut}
          >
            <LogOut size={18} strokeWidth={2.1} />
            <strong>{isSigningOut ? "Выход..." : "Выйти из аккаунта"}</strong>
            <span>Завершить текущую сессию психолога.</span>
          </button>
        </div>
      </section>
    </PageCard>
  );
}
