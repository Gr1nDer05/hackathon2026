import { Mail, MapPin, Phone, UserCircle2 } from "lucide-react";
import { useMemo } from "react";
import useSWR from "swr";
import { Link, useParams } from "react-router-dom";
import { getPublicTestRequest } from "../../modules/public-tests/api/publicTestsApi";
import { normalizePublicTestAuthor } from "../../modules/public-tests/lib/publicAuthor";
import { readPublicTestSnapshot } from "../../modules/public-tests/lib/publicTestStorage";
import {
  formatDate,
  getPsychologistCity,
  getPsychologistDisplayName,
  getPsychologistEducation,
  getPsychologistInitials,
  getPsychologistPhone,
  getPsychologistSpecialization,
} from "../../modules/psychologist/lib/psychologistUi";
import {
  buildClientSessionPath,
  ROUTES,
} from "../../shared/config/routes";
import PageCard from "../../shared/ui/PageCard";

function buildTelegramHref(value) {
  const raw = String(value || "").trim();
  if (!raw || raw === "—") {
    return "";
  }

  if (/^https?:\/\//i.test(raw)) {
    return raw;
  }

  return `https://t.me/${raw.replace(/^@/, "")}`;
}

export default function ClientAuthorPage() {
  const { slug = "" } = useParams();
  const snapshot = useMemo(() => readPublicTestSnapshot(slug), [slug]);
  const publicTestQuery = useSWR(
    slug ? ["public-test", slug] : null,
    () => getPublicTestRequest(slug),
    {
      revalidateOnFocus: false,
      shouldRetryOnError: false,
    },
  );

  const activeTest = publicTestQuery.data || snapshot?.test || null;
  const author = useMemo(() => normalizePublicTestAuthor(activeTest), [activeTest]);

  if (publicTestQuery.isLoading && !activeTest) {
    return (
      <PageCard
        title="Карточка автора теста"
        description="Загружаем данные специалиста."
        links={[{ to: ROUTES.root, label: "На главную" }]}
      >
        <div className="builder-empty">Загружаем карточку автора...</div>
      </PageCard>
    );
  }

  if (publicTestQuery.error && !activeTest) {
    return (
      <PageCard
        title="Карточка автора теста"
        description="Не удалось получить данные по публичной ссылке."
        links={[{ to: ROUTES.root, label: "На главную" }]}
      >
        <p className="admin-form-message admin-form-message--error">
          {publicTestQuery.error.message || "Не удалось загрузить карточку автора."}
        </p>
      </PageCard>
    );
  }

  if (!activeTest || !author) {
    return (
      <PageCard
        title="Карточка автора теста"
        description="Для этой методики публичная карточка автора пока недоступна."
        links={[{ to: buildClientSessionPath(slug), label: "К тесту" }]}
      >
        <div className="builder-empty">
          Сервер пока не передал карточку автора для этого теста.{" "}
          <Link to={buildClientSessionPath(slug)}>Вернуться к прохождению</Link>
        </div>
      </PageCard>
    );
  }

  const displayName = getPsychologistDisplayName(author);
  const initials = getPsychologistInitials(author);
  const specialization = getPsychologistSpecialization(author);
  const city = getPsychologistCity(author);
  const phone = getPsychologistPhone(author);
  const education = getPsychologistEducation(author);
  const headline = author?.card?.headline || "Карточка специалиста";
  const shortBio = author?.card?.short_bio || "";
  const about = author?.profile?.about || "";
  const email = author?.card?.contact_email || author?.email || "—";
  const telegram = author?.card?.telegram || "—";
  const telegramHref = buildTelegramHref(author?.card?.telegram);
  const experienceYears = author?.profile?.experience_years;
  const workFormat =
    [
      author?.card?.online_available ? "Онлайн" : "",
      author?.card?.offline_available ? "Офлайн" : "",
    ]
      .filter(Boolean)
      .join(" / ") || "Формат не указан";
  const updatedAt = formatDate(
    author?.card?.updated_at || author?.profile?.updated_at || author?.updated_at,
  );

  return (
    <PageCard
      wide
      title={`Автор теста: ${displayName}`}
      description={
        activeTest?.title
          ? `Перед прохождением «${activeTest.title}» можно ознакомиться с карточкой специалиста.`
          : "Публичная карточка автора теста."
      }
      links={[
        { to: buildClientSessionPath(slug), label: "К тесту" },
        { to: ROUTES.root, label: "На главную" },
      ]}
    >
      <section className="psychologist-profile-hero">
        <div className="psychologist-profile-hero__identity">
          <div className="psychologist-profile-hero__avatar" aria-hidden="true">
            {initials}
          </div>
          <div>
            <p className="psychologist-profile-hero__eyebrow">Карточка автора</p>
            <h2 className="psychologist-profile-hero__name">{displayName}</h2>
            <p className="psychologist-profile-hero__role">{specialization}</p>
          </div>
        </div>
        <div className="psychologist-profile-hero__badges">
          <span className="status-badge status-badge--active">{workFormat}</span>
          {updatedAt !== "—" ? (
            <span className="status-badge status-badge--draft">Обновлено {updatedAt}</span>
          ) : null}
        </div>
      </section>

      <section className="psychologist-kpis">
        <article className="psychologist-kpi">
          <p className="psychologist-kpi__label">Город</p>
          <p className="psychologist-kpi__value psychologist-kpi__value--small">
            {city}
          </p>
        </article>
        <article className="psychologist-kpi">
          <p className="psychologist-kpi__label">Формат</p>
          <p className="psychologist-kpi__value psychologist-kpi__value--small">
            {workFormat}
          </p>
        </article>
        <article className="psychologist-kpi">
          <p className="psychologist-kpi__label">Стаж</p>
          <p className="psychologist-kpi__value psychologist-kpi__value--small">
            {experienceYears === null || experienceYears === undefined
              ? "—"
              : `${experienceYears} лет`}
          </p>
        </article>
        <article className="psychologist-kpi">
          <p className="psychologist-kpi__label">Тест</p>
          <p className="psychologist-kpi__value psychologist-kpi__value--small">
            {activeTest?.title || "Текущая методика"}
          </p>
        </article>
      </section>

      <section className="psychologist-profile-grid">
        <article className="admin-panel">
          <h3 className="admin-panel__title">Контакты</h3>
          <dl className="profile-meta-list">
            <div>
              <dt>
                <Mail size={15} strokeWidth={2.1} /> Email
              </dt>
              <dd>{email}</dd>
            </div>
            <div>
              <dt>
                <Phone size={15} strokeWidth={2.1} /> Телефон
              </dt>
              <dd>{phone}</dd>
            </div>
            <div>
              <dt>
                <MapPin size={15} strokeWidth={2.1} /> Город
              </dt>
              <dd>{city}</dd>
            </div>
            <div>
              <dt>
                <UserCircle2 size={15} strokeWidth={2.1} /> Telegram
              </dt>
              <dd>{telegram}</dd>
            </div>
          </dl>
        </article>

        <article className="admin-panel">
          <h3 className="admin-panel__title">Профессиональный профиль</h3>
          <dl className="profile-meta-list">
            <div>
              <dt>Специализация</dt>
              <dd>{specialization}</dd>
            </div>
            <div>
              <dt>Образование</dt>
              <dd>{education}</dd>
            </div>
            <div>
              <dt>Формат работы</dt>
              <dd>{workFormat}</dd>
            </div>
          </dl>
        </article>

        <article className="admin-panel psychologist-profile-panel--about">
          <h3 className="admin-panel__title">{headline}</h3>
          <div className="psychologist-profile-text psychologist-profile-text--panel">
            <p className="psychologist-profile-text__body">
              {about || shortBio || "Автор теста пока не заполнил описание своей карточки."}
            </p>
          </div>
        </article>
      </section>

      <div className="workflow-note workflow-note--success">
        <p>
          Если специалист вам подходит, можно вернуться к тесту и продолжить прохождение.
        </p>
        <div className="workflow-note__actions">
          {email !== "—" ? (
            <a className="table-action-link" href={`mailto:${email}`}>
              Написать на почту
            </a>
          ) : null}
          {phone !== "—" ? (
            <a className="table-action-link" href={`tel:${String(phone).replace(/[^\d+]/g, "")}`}>
              Позвонить
            </a>
          ) : null}
          {telegramHref ? (
            <a className="table-action-link" href={telegramHref} target="_blank" rel="noreferrer">
              Telegram
            </a>
          ) : null}
          <Link className="table-action-link" to={buildClientSessionPath(slug)}>
            Вернуться к тесту
          </Link>
        </div>
      </div>
    </PageCard>
  );
}
