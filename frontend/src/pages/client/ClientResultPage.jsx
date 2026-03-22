import { Award, BarChart3, FileDown, FileText, LoaderCircle, RotateCcw } from "lucide-react";
import { AnimatePresence, motion, useReducedMotion } from "motion/react";
import { useMemo, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { getPublicTestReportRequest } from "../../modules/public-tests/api/publicTestsApi";
import { normalizePublicTestAuthor } from "../../modules/public-tests/lib/publicAuthor";
import {
  formatDate,
  getResultMetricEntries,
  getResultProfessionEntries,
  getResultTopMetricEntries,
} from "../../modules/psychologist/lib/psychologistUi";
import { readPublicTestSnapshot } from "../../modules/public-tests/lib/publicTestStorage";
import { API_BASE_URL } from "../../shared/api/http";
import {
  buildClientAuthorPath,
  buildClientSessionPath,
  ROUTES,
} from "../../shared/config/routes";
import {
  createFadeMove,
  createRevealContainer,
} from "../../shared/lib/motion";
import PageCard from "../../shared/ui/PageCard";

export default function ClientResultPage() {
  const { slug = "" } = useParams();
  const reducedMotion = useReducedMotion();
  const [reportError, setReportError] = useState("");
  const [reportLoadingFormat, setReportLoadingFormat] = useState("");
  const sectionVariants = createRevealContainer(reducedMotion, {
    staggerChildren: 0.08,
    delayChildren: 0.04,
  });
  const blockVariants = createFadeMove(reducedMotion, {
    axis: "y",
    distance: 18,
    scale: 0.992,
  });
  const heroVariants = createRevealContainer(reducedMotion, {
    staggerChildren: 0.06,
    delayChildren: 0.03,
  });
  const cardVariants = createFadeMove(reducedMotion, {
    axis: "y",
    distance: 14,
    scale: 0.996,
  });
  const snapshot = useMemo(() => readPublicTestSnapshot(slug), [slug]);
  const resultPayload = snapshot?.result || null;
  const metricEntries = useMemo(() => getResultMetricEntries(resultPayload), [resultPayload]);
  const topMetrics = useMemo(() => getResultTopMetricEntries(resultPayload), [resultPayload]);
  const topProfessions = useMemo(() => getResultProfessionEntries(resultPayload), [resultPayload]);
  const respondentEmail = String(snapshot?.startForm?.respondent_email || "").trim();
  const publicAuthor = useMemo(() => normalizePublicTestAuthor(snapshot?.test), [snapshot?.test]);
  const hasCompletedResult = snapshot?.result?.status === "completed";
  const clientReportUrl = useMemo(() => {
    const rawUrl = String(snapshot?.result?.client_report_url || "").trim();

    if (!rawUrl) {
      return "";
    }

    if (/^https?:\/\//i.test(rawUrl)) {
      return rawUrl;
    }

    if (rawUrl.startsWith("/")) {
      return `${API_BASE_URL}${rawUrl}`;
    }

    return `${API_BASE_URL}/${rawUrl}`;
  }, [snapshot?.result?.client_report_url]);
  const clientReportAccessToken = useMemo(() => {
    const directToken = String(snapshot?.session?.access_token || "").trim();

    if (directToken) {
      return directToken;
    }

    if (!clientReportUrl) {
      return "";
    }

    try {
      return new URL(clientReportUrl, window.location.origin).searchParams.get("access_token") || "";
    } catch {
      return "";
    }
  }, [clientReportUrl, snapshot?.session?.access_token]);
  const clientReportAvailable = Boolean(
    hasCompletedResult &&
      snapshot?.result?.client_report_available === true &&
      clientReportAccessToken,
  );

  async function handleOpenClientReport(format = "html") {
    const reportTab =
      format === "html"
        ? window.open("about:blank", "_blank", "noopener,noreferrer")
        : null;

    setReportLoadingFormat(format);
    setReportError("");

    try {
      const file = await getPublicTestReportRequest(slug, {
        accessToken: clientReportAccessToken,
        format,
      });
      const objectUrl = window.URL.createObjectURL(file.blob);

      if (format === "html") {
        if (!reportTab) {
          window.location.assign(objectUrl);
          return;
        }

        reportTab.location.href = objectUrl;
        window.setTimeout(() => window.URL.revokeObjectURL(objectUrl), 60_000);
      } else {
        const link = document.createElement("a");
        link.href = objectUrl;
        link.download = file.filename || `client-report-${slug}.docx`;
        document.body.append(link);
        link.click();
        link.remove();
        window.setTimeout(() => window.URL.revokeObjectURL(objectUrl), 1_000);
      }
    } catch (error) {
      if (reportTab) {
        reportTab.close();
      }

      if (error?.status === 409) {
        setReportError(
          "Отчёт пока недоступен. Попробуйте открыть его ещё раз через несколько секунд.",
        );
      } else {
        setReportError(error?.message || "Не удалось открыть клиентский отчёт.");
      }
    } finally {
      setReportLoadingFormat("");
    }
  }

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
      <motion.div
        animate="visible"
        initial="hidden"
        variants={sectionVariants}
      >
        <AnimatePresence>
          {reportError ? (
            <motion.p
              key={`report-error-${reportError}`}
              animate={{ opacity: 1, y: 0 }}
              className="admin-form-message admin-form-message--error"
              exit={{ opacity: 0, y: reducedMotion ? 0 : -10 }}
              initial={{ opacity: 0, y: reducedMotion ? 0 : -10 }}
              transition={{ duration: reducedMotion ? 0.01 : 0.22 }}
            >
              {reportError}
            </motion.p>
          ) : null}
        </AnimatePresence>

        <motion.div
          className={`workflow-note ${respondentEmail ? "workflow-note--success" : "workflow-note--warning"}`}
          variants={blockVariants}
        >
          <p>
            {respondentEmail
              ? `Результат связан с почтой ${respondentEmail}. Если у специалиста настроена отправка, итоговые материалы уйдут на этот адрес.`
              : "Почта для результатов не была указана. Итог сохранён в этом браузере, но отправка на email может быть недоступна."}
          </p>
        </motion.div>

        <motion.div
          className={`workflow-note ${clientReportAvailable ? "workflow-note--success" : "workflow-note--warning"}`}
          variants={blockVariants}
        >
          <p>
            {clientReportAvailable
              ? "Клиентский отчёт уже доступен. Его можно открыть в HTML или сразу скачать в DOCX."
              : "Отчёт появится немного позже или станет доступен после повторной проверки."}
          </p>
          {clientReportAvailable ? (
            <div className="workflow-note__actions">
              <motion.button
                className="table-action-button"
                disabled={Boolean(reportLoadingFormat)}
                type="button"
                whileHover={reducedMotion ? undefined : { y: -1, scale: 1.01 }}
                whileTap={reducedMotion ? undefined : { scale: 0.992 }}
                onClick={() => handleOpenClientReport("html")}
              >
                {reportLoadingFormat === "html" ? (
                  <LoaderCircle className="icon-spin" size={15} strokeWidth={2.1} />
                ) : (
                  <FileText size={15} strokeWidth={2.1} />
                )}
                <span>HTML отчёт</span>
              </motion.button>
              <motion.button
                className="table-action-button"
                disabled={Boolean(reportLoadingFormat)}
                type="button"
                whileHover={reducedMotion ? undefined : { y: -1, scale: 1.01 }}
                whileTap={reducedMotion ? undefined : { scale: 0.992 }}
                onClick={() => handleOpenClientReport("docx")}
              >
                {reportLoadingFormat === "docx" ? (
                  <LoaderCircle className="icon-spin" size={15} strokeWidth={2.1} />
                ) : (
                  <FileDown size={15} strokeWidth={2.1} />
                )}
                <span>DOCX отчёт</span>
              </motion.button>
            </div>
          ) : null}
        </motion.div>

        <motion.div
          className="client-author-access client-author-access--compact"
          variants={blockVariants}
        >
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
        </motion.div>

        <motion.section
          className="client-result-hero"
          variants={heroVariants}
        >
          <motion.article className="client-result-card" variants={cardVariants}>
            <BarChart3 size={18} strokeWidth={2.1} />
            <div>
              <p className="client-result-card__label">Респондент</p>
              <strong>{snapshot?.startForm?.respondent_name || "Не указан"}</strong>
            </div>
          </motion.article>
          <motion.article className="client-result-card" variants={cardVariants}>
            <Award size={18} strokeWidth={2.1} />
            <div>
              <p className="client-result-card__label">Лучшие метрики</p>
              <strong>{topMetrics.slice(0, 2).map((item) => item.label).join(", ") || "Не рассчитано"}</strong>
            </div>
          </motion.article>
          <motion.article className="client-result-card" variants={cardVariants}>
            <RotateCcw size={18} strokeWidth={2.1} />
            <div>
              <p className="client-result-card__label">Статус сессии</p>
              <strong>{snapshot?.result?.status === "completed" ? "Тест завершён" : "Есть незавершённые ответы"}</strong>
            </div>
          </motion.article>
          <motion.article className="client-result-card" variants={cardVariants}>
            <Award size={18} strokeWidth={2.1} />
            <div>
              <p className="client-result-card__label">Сохранено</p>
              <strong>{formatDate(snapshot?.saved_at)}</strong>
            </div>
          </motion.article>
        </motion.section>

        <motion.section
          className="client-result-grid"
          variants={heroVariants}
        >
          <motion.article className="builder-panel" variants={blockVariants}>
            <div className="builder-section-head">
              <div>
                <p className="builder-section-head__eyebrow">Метрики теста</p>
                <h3 className="builder-section-head__title">Итоговое распределение</h3>
              </div>
            </div>

            <motion.div className="client-scale-list" variants={heroVariants}>
              {metricEntries.length ? (
                metricEntries.map((item) => (
                  <motion.div className="client-scale-row" key={item.key} variants={cardVariants}>
                    <div className="client-scale-row__head">
                      <strong>{item.label}</strong>
                      <span>{item.displayValue}</span>
                    </div>
                    <div className="client-progress-bar">
                      <motion.div
                        className="client-progress-bar__fill"
                        initial={{ width: 0 }}
                        animate={{ width: `${item.progress}%` }}
                        transition={{ duration: reducedMotion ? 0.01 : 0.5 }}
                      />
                    </div>
                    <p className="client-scale-row__meta">{item.meta}</p>
                  </motion.div>
                ))
              ) : (
                <p className="psychologist-empty-state">Сервер пока не вернул рассчитанные метрики для этой сессии.</p>
              )}
            </motion.div>
          </motion.article>

          <motion.article className="builder-panel" variants={blockVariants}>
            <div className="builder-section-head">
              <div>
                <p className="builder-section-head__eyebrow">Рекомендации</p>
                <h3 className="builder-section-head__title">Наиболее подходящие направления</h3>
              </div>
            </div>

            <motion.div className="client-profession-list" variants={heroVariants}>
              {topProfessions.length ? (
                topProfessions.map((item, index) => (
                  <motion.div className="client-profession-card" key={`${item.profession}-${index}`} variants={cardVariants}>
                    <span className="client-profession-card__index">#{index + 1}</span>
                    <div>
                      <strong>{item.profession}</strong>
                      <p>{item.score} баллов</p>
                    </div>
                  </motion.div>
                ))
              ) : (
                <p className="psychologist-empty-state">Профессиональные рекомендации для этой сессии ещё не рассчитаны.</p>
              )}
            </motion.div>
          </motion.article>
        </motion.section>

        <motion.section
          className="psychologist-summary-grid"
          variants={heroVariants}
        >
          <motion.div className="psychologist-summary-card" variants={cardVariants}>
            <span>Почта</span>
            <strong>{respondentEmail || "Не указана"}</strong>
            <p>Адрес, на который можно отправлять итоговый отчёт и рекомендации.</p>
          </motion.div>
          <motion.div className="psychologist-summary-card" variants={cardVariants}>
            <span>Топ-метрики</span>
            <strong>{topMetrics.slice(0, 2).map((item) => item.label).join(", ") || "Нет данных"}</strong>
            <p>Наиболее выраженные направления по рассчитанному результату.</p>
          </motion.div>
          <motion.div className="psychologist-summary-card" variants={cardVariants}>
            <span>Рекомендаций</span>
            <strong>{topProfessions.length}</strong>
            <p>Количество направлений, которые были подобраны по результатам теста.</p>
          </motion.div>
        </motion.section>
      </motion.div>
    </PageCard>
  );
}
