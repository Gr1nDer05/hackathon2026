import { Award, BarChart3, RotateCcw } from "lucide-react";
import { useMemo } from "react";
import { Link, useParams } from "react-router-dom";
import { normalizePublicTestAuthor } from "../../modules/public-tests/lib/publicAuthor";
import {
  formatDate,
  getResultMetricEntries,
  getResultProfessionEntries,
  getResultTopMetricEntries,
} from "../../modules/psychologist/lib/psychologistUi";
import { readPublicTestSnapshot } from "../../modules/public-tests/lib/publicTestStorage";
import {
  buildClientAuthorPath,
  buildClientSessionPath,
  ROUTES,
} from "../../shared/config/routes";
import PageCard from "../../shared/ui/PageCard";

export default function ClientResultPage() {
  const { slug = "" } = useParams();
  const snapshot = useMemo(() => readPublicTestSnapshot(slug), [slug]);
  const resultPayload = snapshot?.result || null;
  const metricEntries = useMemo(() => getResultMetricEntries(resultPayload), [resultPayload]);
  const topMetrics = useMemo(() => getResultTopMetricEntries(resultPayload), [resultPayload]);
  const topProfessions = useMemo(() => getResultProfessionEntries(resultPayload), [resultPayload]);
  const respondentEmail = String(snapshot?.startForm?.respondent_email || "").trim();
  const publicAuthor = useMemo(() => normalizePublicTestAuthor(snapshot?.test), [snapshot?.test]);

  if (!resultPayload) {
    return (
      <PageCard
        wide
        title="Результат пока недоступен"
        description="Итоговая интерпретация появляется после завершения теста в этом браузере."
        links={[
          { to: buildClientSessionPath(slug), label: "Вернуться к тесту" },
          { to: ROUTES.root, label: "На главную" },
        ]}
      >
        <div className="builder-empty">
          Данные результата не найдены. Завершите тест заново по публичной ссылке, чтобы сформировать клиентский отчёт.
        </div>
      </PageCard>
    );
  }

  return (
    <PageCard
      wide
      title={snapshot?.test?.title ? `Результат: ${snapshot.test.title}` : "Результат теста"}
      description="Клиентский итог по рассчитанным метрикам теста и рекомендуемым направлениям."
      links={[
        { to: buildClientAuthorPath(slug), label: "Об авторе" },
        { to: buildClientSessionPath(slug), label: "Вернуться к тесту" },
        { to: ROUTES.root, label: "На главную" },
      ]}
    >
      <div className={`workflow-note ${respondentEmail ? "workflow-note--success" : "workflow-note--warning"}`}>
        <p>
          {respondentEmail
            ? `Результат связан с почтой ${respondentEmail}. Если у специалиста настроена отправка, итоговые материалы уйдут на этот адрес.`
            : "Почта для результатов не была указана. Итог сохранён в этом браузере, но отправка на email может быть недоступна."}
        </p>
      </div>

      <div className="client-author-access client-author-access--compact">
        <div>
          <p className="client-author-access__eyebrow">Автор теста</p>
          <strong className="client-author-access__title">
            {publicAuthor?.full_name || "Карточка специалиста"}
          </strong>
          <p className="client-author-access__text">
            После результата можно вернуться к карточке специалиста и уточнить дальнейшие шаги.
          </p>
        </div>
        <Link className="table-action-link" to={buildClientAuthorPath(slug)}>
          Открыть карточку
        </Link>
      </div>

      <section className="client-result-hero">
        <article className="client-result-card">
          <BarChart3 size={18} strokeWidth={2.1} />
          <div>
            <p className="client-result-card__label">Респондент</p>
            <strong>{snapshot?.startForm?.respondent_name || "Не указан"}</strong>
          </div>
        </article>
        <article className="client-result-card">
          <Award size={18} strokeWidth={2.1} />
          <div>
            <p className="client-result-card__label">Лучшие метрики</p>
            <strong>{topMetrics.slice(0, 2).map((item) => item.label).join(", ") || "Не рассчитано"}</strong>
          </div>
        </article>
        <article className="client-result-card">
          <RotateCcw size={18} strokeWidth={2.1} />
          <div>
            <p className="client-result-card__label">Статус сессии</p>
            <strong>{snapshot?.result?.status === "completed" ? "Тест завершён" : "Есть незавершённые ответы"}</strong>
          </div>
        </article>
        <article className="client-result-card">
          <Award size={18} strokeWidth={2.1} />
          <div>
            <p className="client-result-card__label">Сохранено</p>
            <strong>{formatDate(snapshot?.saved_at)}</strong>
          </div>
        </article>
      </section>

      <section className="client-result-grid">
        <article className="builder-panel">
          <div className="builder-section-head">
            <div>
              <p className="builder-section-head__eyebrow">Метрики теста</p>
              <h3 className="builder-section-head__title">Итоговое распределение</h3>
            </div>
          </div>

          <div className="client-scale-list">
            {metricEntries.length ? (
              metricEntries.map((item) => (
              <div className="client-scale-row" key={item.key}>
                <div className="client-scale-row__head">
                  <strong>{item.label}</strong>
                  <span>{item.displayValue}</span>
                </div>
                <div className="client-progress-bar">
                  <div className="client-progress-bar__fill" style={{ width: `${item.progress}%` }} />
                </div>
                <p className="client-scale-row__meta">{item.meta}</p>
              </div>
            ))
            ) : (
              <p className="psychologist-empty-state">Сервер пока не вернул рассчитанные метрики для этой сессии.</p>
            )}
          </div>
        </article>

        <article className="builder-panel">
          <div className="builder-section-head">
            <div>
              <p className="builder-section-head__eyebrow">Рекомендации</p>
              <h3 className="builder-section-head__title">Наиболее подходящие направления</h3>
            </div>
          </div>

          <div className="client-profession-list">
            {topProfessions.length ? (
              topProfessions.map((item, index) => (
                <div className="client-profession-card" key={`${item.profession}-${index}`}>
                  <span className="client-profession-card__index">#{index + 1}</span>
                  <div>
                    <strong>{item.profession}</strong>
                    <p>{item.score} баллов</p>
                  </div>
                </div>
              ))
            ) : (
              <p className="psychologist-empty-state">Профессиональные рекомендации для этой сессии ещё не рассчитаны.</p>
            )}
          </div>
        </article>
      </section>

      <section className="psychologist-summary-grid">
        <div className="psychologist-summary-card">
          <span>Почта</span>
          <strong>{respondentEmail || "Не указана"}</strong>
          <p>Адрес, на который можно отправлять итоговый отчёт и рекомендации.</p>
        </div>
        <div className="psychologist-summary-card">
          <span>Топ-метрики</span>
          <strong>{topMetrics.slice(0, 2).map((item) => item.label).join(", ") || "Нет данных"}</strong>
          <p>Наиболее выраженные направления по рассчитанному результату.</p>
        </div>
        <div className="psychologist-summary-card">
          <span>Рекомендаций</span>
          <strong>{topProfessions.length}</strong>
          <p>Количество направлений, которые backend вернул как наиболее подходящие.</p>
        </div>
      </section>
    </PageCard>
  );
}
