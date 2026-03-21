import { CheckCircle2, LoaderCircle, Save, Send } from "lucide-react";
import { useEffect, useMemo, useRef, useState } from "react";
import useSWR from "swr";
import { Link, useNavigate, useParams } from "react-router-dom";
import {
  getPublicTestRequest,
  savePublicTestProgressRequest,
  startPublicTestRequest,
  submitPublicTestRequest,
} from "../../modules/public-tests/api/publicTestsApi";
import {
  buildAnswersMap,
  buildAnswersPayload,
  readPublicTestSnapshot,
  writePublicTestSnapshot,
} from "../../modules/public-tests/lib/publicTestStorage";
import { normalizePublicTestAuthor } from "../../modules/public-tests/lib/publicAuthor";
import {
  buildClientAuthorPath,
  buildClientResultPath,
  ROUTES,
} from "../../shared/config/routes";
import {
  formatRussianPhoneInput,
  isRussianPhoneComplete,
} from "../../shared/lib/russianPhone";
import PageCard from "../../shared/ui/PageCard";

const EMPTY_START_FORM = {
  respondent_name: "",
  respondent_phone: "+7",
  respondent_email: "",
  respondent_age: "",
  respondent_gender: "",
  respondent_education: "",
};

function getSessionTest(fallbackTest, snapshotTest) {
  return snapshotTest || fallbackTest || null;
}

function validateStartForm(form, test) {
  const errors = {};
  const fullName = String(form.respondent_name || "").trim();
  const fullNameParts = fullName.split(/\s+/).filter(Boolean);
  const email = String(form.respondent_email || "").trim();
  const age = String(form.respondent_age || "").trim();

  if (fullNameParts.length < 2) {
    errors.respondent_name = "Укажите имя и фамилию.";
  }

  if (!isRussianPhoneComplete(form.respondent_phone)) {
    errors.respondent_phone = "Введите номер в формате +7 999 123-45-67.";
  }

  if (email && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
    errors.respondent_email = "Укажите корректный email.";
  }

  if (test?.collect_respondent_age) {
    const ageValue = Number(age);
    if (!Number.isInteger(ageValue) || ageValue < 1 || ageValue > 120) {
      errors.respondent_age = "Возраст должен быть числом от 1 до 120.";
    }
  }

  if (
    test?.collect_respondent_gender &&
    !String(form.respondent_gender || "").trim()
  ) {
    errors.respondent_gender = "Выберите пол.";
  }

  if (
    test?.collect_respondent_education &&
    !String(form.respondent_education || "").trim()
  ) {
    errors.respondent_education = "Укажите образование.";
  }

  return errors;
}

function isQuestionAnswered(question, answer) {
  if (!question?.is_required) {
    return true;
  }

  if (question.question_type === "text") {
    return Boolean(String(answer?.answer_text || "").trim());
  }

  if (question.question_type === "multiple_choice") {
    return (
      Array.isArray(answer?.answer_values) && answer.answer_values.length > 0
    );
  }

  return Boolean(String(answer?.answer_value || "").trim());
}

function validateSubmit(test, answersByQuestion) {
  const issues = [];

  for (const question of test?.questions || []) {
    if (!isQuestionAnswered(question, answersByQuestion[String(question.id)])) {
      issues.push(`Ответьте на обязательный вопрос: «${question.text}».`);
    }
  }

  return issues;
}

function createStartPayload(form, test) {
  const payload = {
    respondent_name: String(form.respondent_name || "").trim(),
    respondent_phone: String(form.respondent_phone || "").trim(),
  };

  const email = String(form.respondent_email || "").trim();
  const age = String(form.respondent_age || "").trim();
  const gender = String(form.respondent_gender || "").trim();
  const education = String(form.respondent_education || "").trim();

  if (email) {
    payload.respondent_email = email;
  }

  if (test?.collect_respondent_age && age) {
    payload.respondent_age = Number(age);
  }

  if (test?.collect_respondent_gender && gender) {
    payload.respondent_gender = gender;
  }

  if (test?.collect_respondent_education && education) {
    payload.respondent_education = education;
  }

  return payload;
}

function getAnsweredCount(test, answersByQuestion) {
  return (test?.questions || []).filter((question) =>
    isQuestionAnswered(
      { ...question, is_required: true },
      answersByQuestion[String(question.id)],
    ),
  ).length;
}

function buildSnapshot(slug, current) {
  const snapshot = {
    slug,
    saved_at: new Date().toISOString(),
    ...current,
  };
  writePublicTestSnapshot(slug, snapshot);
}

function buildAutoSaveKey(accessToken, answersByQuestion) {
  const answers = buildAnswersPayload(answersByQuestion);

  return JSON.stringify({
    access_token: String(accessToken || ""),
    answers,
  });
}

export default function ClientSessionPage() {
  const { slug = "" } = useParams();
  const navigate = useNavigate();
  const publicTestQuery = useSWR(
    slug ? ["public-test", slug] : null,
    () => getPublicTestRequest(slug),
    {
      revalidateOnFocus: false,
      shouldRetryOnError: false,
    },
  );

  const [startForm, setStartForm] = useState(EMPTY_START_FORM);
  const [startErrors, setStartErrors] = useState({});
  const [session, setSession] = useState(null);
  const [sessionTest, setSessionTest] = useState(null);
  const [answersByQuestion, setAnswersByQuestion] = useState({});
  const [feedbackError, setFeedbackError] = useState("");
  const [feedbackMessage, setFeedbackMessage] = useState("");
  const [isStarting, setIsStarting] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [autoSaveState, setAutoSaveState] = useState("");
  const autoSaveKeyRef = useRef("");
  const autoSaveInFlightRef = useRef(false);

  useEffect(() => {
    const snapshot = readPublicTestSnapshot(slug);
    if (!snapshot) {
      setStartForm(EMPTY_START_FORM);
      setSession(null);
      setSessionTest(null);
      setAnswersByQuestion({});
      setAutoSaveState("");
      autoSaveKeyRef.current = "";
      return;
    }

    setStartForm({
      ...EMPTY_START_FORM,
      ...snapshot.startForm,
      respondent_phone: formatRussianPhoneInput(
        snapshot.startForm?.respondent_phone || "+7",
      ),
    });
    setSession(snapshot.session || null);
    setSessionTest(snapshot.test || null);
    setAnswersByQuestion(
      snapshot.answersByQuestion ||
        buildAnswersMap(snapshot.result?.answers || snapshot.answers),
    );
    autoSaveKeyRef.current = buildAutoSaveKey(
      snapshot.session?.access_token || "",
      snapshot.answersByQuestion ||
        buildAnswersMap(snapshot.result?.answers || snapshot.answers),
    );
  }, [slug]);

  useEffect(() => {
    if (!feedbackMessage) {
      return undefined;
    }

    const timeoutId = window.setTimeout(() => setFeedbackMessage(""), 2600);
    return () => window.clearTimeout(timeoutId);
  }, [feedbackMessage]);

  useEffect(() => {
    if (!autoSaveState) {
      return undefined;
    }

    const timeoutId = window.setTimeout(() => setAutoSaveState(""), 2600);
    return () => window.clearTimeout(timeoutId);
  }, [autoSaveState]);

  const activeTest = getSessionTest(publicTestQuery.data, sessionTest);
  const answeredCount = useMemo(
    () => getAnsweredCount(activeTest, answersByQuestion),
    [activeTest, answersByQuestion],
  );
  const publicAuthor = useMemo(
    () => normalizePublicTestAuthor(activeTest),
    [activeTest],
  );
  const questionCount = activeTest?.questions?.length || 0;
  const completionPercent = questionCount
    ? Math.round((answeredCount / questionCount) * 100)
    : 0;
  const hasRespondentEmail = Boolean(
    String(startForm.respondent_email || "").trim(),
  );

  async function saveProgress({ silent = false } = {}) {
    if (!session?.access_token) {
      if (!silent) {
        setFeedbackError("Сначала начните тест.");
      }
      return false;
    }

    const answers = buildAnswersPayload(answersByQuestion);
    if (!answers.length) {
      if (!silent) {
        setFeedbackError("Добавьте хотя бы один ответ перед сохранением.");
      }
      return false;
    }

    const currentKey = buildAutoSaveKey(
      session.access_token,
      answersByQuestion,
    );
    if (silent && currentKey === autoSaveKeyRef.current) {
      return true;
    }

    if (silent) {
      autoSaveInFlightRef.current = true;
    } else {
      setIsSaving(true);
    }
    setFeedbackError("");

    try {
      const response = await savePublicTestProgressRequest(slug, {
        access_token: session.access_token,
        answers,
      });
      const nextAnswers = buildAnswersMap(response.answers);
      setAnswersByQuestion(nextAnswers);
      autoSaveKeyRef.current = buildAutoSaveKey(
        session.access_token,
        nextAnswers,
      );
      buildSnapshot(slug, {
        session: {
          ...session,
          id: response.session_id,
          status: response.status,
        },
        test: activeTest,
        startForm,
        answersByQuestion: nextAnswers,
        answers: response.answers || [],
      });

      if (silent) {
        setAutoSaveState("Черновик сохранён автоматически.");
      } else {
        setFeedbackMessage("Прогресс сохранён.");
      }

      return true;
    } catch (error) {
      if (silent) {
        setAutoSaveState(
          "Автосохранение не удалось. Используйте кнопку «Сохранить прогресс».",
        );
      } else {
        setFeedbackError(error?.message || "Не удалось сохранить прогресс.");
      }
      return false;
    } finally {
      if (silent) {
        autoSaveInFlightRef.current = false;
      } else {
        setIsSaving(false);
      }
    }
  }

  useEffect(() => {
    if (!session?.access_token || !activeTest || isSubmitting || isSaving) {
      return undefined;
    }

    const answers = buildAnswersPayload(answersByQuestion);
    if (!answers.length) {
      return undefined;
    }

    const pendingKey = buildAutoSaveKey(
      session.access_token,
      answersByQuestion,
    );
    if (pendingKey === autoSaveKeyRef.current) {
      return undefined;
    }

    const timeoutId = window.setTimeout(() => {
      if (autoSaveInFlightRef.current) {
        return;
      }
      void saveProgress({ silent: true });
    }, 15000);

    return () => window.clearTimeout(timeoutId);
  }, [
    activeTest,
    answersByQuestion,
    isSaving,
    isSubmitting,
    session?.access_token,
    slug,
    startForm,
  ]);

  function handleStartFieldChange(field, value) {
    const nextValue =
      field === "respondent_phone" ? formatRussianPhoneInput(value) : value;
    setStartForm((prev) => ({ ...prev, [field]: nextValue }));
    setFeedbackError("");
    setStartErrors((prev) => {
      if (!prev[field]) {
        return prev;
      }

      const next = { ...prev };
      delete next[field];
      return next;
    });
  }

  function updateAnswer(questionId, patch) {
    setAnswersByQuestion((prev) => {
      const next = {
        ...prev,
        [String(questionId)]: {
          answer_text: "",
          answer_value: "",
          answer_values: [],
          ...prev[String(questionId)],
          ...patch,
        },
      };

      buildSnapshot(slug, {
        session,
        test: activeTest,
        startForm,
        answersByQuestion: next,
      });

      return next;
    });
    setFeedbackError("");
  }

  async function handleStart(event) {
    event.preventDefault();

    const errors = validateStartForm(startForm, publicTestQuery.data);
    setStartErrors(errors);
    if (Object.keys(errors).length) {
      return;
    }

    setIsStarting(true);
    setFeedbackError("");

    try {
      const payload = createStartPayload(startForm, publicTestQuery.data);
      const response = await startPublicTestRequest(slug, payload);
      const nextAnswers = buildAnswersMap(response.answers);

      setSession(response.session);
      setSessionTest(response.test);
      setAnswersByQuestion(nextAnswers);
      autoSaveKeyRef.current = buildAutoSaveKey(
        response.session?.access_token || "",
        nextAnswers,
      );
      setFeedbackMessage(
        response.resumed
          ? "Найдена незавершённая сессия. Продолжайте с сохранённого места."
          : "Тест запущен. Можно отвечать на вопросы.",
      );
      buildSnapshot(slug, {
        session: response.session,
        test: response.test,
        startForm,
        answersByQuestion: nextAnswers,
        answers: response.answers || [],
      });
    } catch (error) {
      setFeedbackError(error?.message || "Не удалось начать тест.");
    } finally {
      setIsStarting(false);
    }
  }

  async function handleSaveProgress() {
    await saveProgress();
  }

  async function handleSubmit() {
    if (!session?.access_token) {
      setFeedbackError("Сначала начните тест.");
      return;
    }

    const issues = validateSubmit(activeTest, answersByQuestion);
    if (issues.length) {
      setFeedbackError(issues[0]);
      return;
    }

    const answers = buildAnswersPayload(answersByQuestion);
    if (!answers.length) {
      setFeedbackError("Заполните хотя бы один ответ.");
      return;
    }

    setIsSubmitting(true);
    setFeedbackError("");

    try {
      const response = await submitPublicTestRequest(slug, {
        access_token: session.access_token,
        answers,
      });

      buildSnapshot(slug, {
        session: {
          ...session,
          id: response.session_id,
          status: response.status,
        },
        test: activeTest,
        startForm,
        answersByQuestion: buildAnswersMap(response.answers),
        result: response,
      });

      navigate(buildClientResultPath(slug));
    } catch (error) {
      setFeedbackError(error?.message || "Не удалось отправить ответы.");
    } finally {
      setIsSubmitting(false);
    }
  }

  function renderQuestion(question) {
    const answer = answersByQuestion[String(question.id)] || {
      answer_text: "",
      answer_value: "",
      answer_values: [],
    };

    if (question.question_type === "text") {
      return (
        <textarea
          className="admin-form-control client-answer-control client-answer-control--textarea"
          rows={4}
          value={answer.answer_text}
          onChange={(event) =>
            updateAnswer(question.id, { answer_text: event.target.value })
          }
          placeholder="Введите ваш ответ"
        />
      );
    }

    if (question.question_type === "number") {
      return (
        <input
          className="admin-form-control client-answer-control"
          type="number"
          value={answer.answer_value}
          onChange={(event) =>
            updateAnswer(question.id, { answer_value: event.target.value })
          }
          placeholder="Введите число"
        />
      );
    }

    if (question.question_type === "multiple_choice") {
      return (
        <div className="client-answer-options">
          {(question.options || []).map((option) => {
            const isChecked = answer.answer_values.includes(option.value);
            return (
              <label
                className={`client-answer-option client-answer-option--multi ${isChecked ? "client-answer-option--checked" : ""}`}
                key={option.id || option.value}
              >
                <input
                  className="client-answer-option__native"
                  checked={isChecked}
                  type="checkbox"
                  onChange={(event) => {
                    const nextValues = event.target.checked
                      ? [...answer.answer_values, option.value]
                      : answer.answer_values.filter(
                          (item) => item !== option.value,
                        );

                    updateAnswer(question.id, { answer_values: nextValues });
                  }}
                />
                <span
                  className="client-answer-option__marker"
                  aria-hidden="true"
                />
                <span className="client-answer-option__text">
                  {option.label}
                </span>
              </label>
            );
          })}
        </div>
      );
    }

    return (
      <div className="client-answer-options">
        {(question.options || []).map((option) => {
          const isScale = question.question_type === "scale";
          const checked = answer.answer_value === option.value;
          return (
            <label
              className={`client-answer-option ${isScale ? "client-answer-option--scale" : "client-answer-option--single"} ${checked ? "client-answer-option--checked" : ""}`}
              key={option.id || option.value}
            >
              <input
                className="client-answer-option__native"
                checked={checked}
                type="radio"
                name={`question-${question.id}`}
                onChange={() =>
                  updateAnswer(question.id, { answer_value: option.value })
                }
              />
              {!isScale ? (
                <span
                  className="client-answer-option__marker"
                  aria-hidden="true"
                />
              ) : null}
              <span className="client-answer-option__text">{option.label}</span>
            </label>
          );
        })}
      </div>
    );
  }

  return (
    <PageCard
      wide
      title={activeTest?.title || "Публичное прохождение теста"}
      description={
        activeTest?.description ||
        "Клиентский сценарий прохождения по уникальной ссылке без регистрации."
      }
      links={[
        { to: buildClientAuthorPath(slug), label: "Об авторе" },
        { to: ROUTES.root, label: "На главную" },
      ]}
    >
      {publicTestQuery.error ? (
        <p className="admin-form-message admin-form-message--error">
          {publicTestQuery.error.message ||
            "Не удалось загрузить тест по публичной ссылке."}
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

      {publicTestQuery.isLoading && !activeTest ? (
        <div className="builder-empty">Загружаем данные теста...</div>
      ) : null}

      {activeTest && !session ? (
        <section className="client-flow-grid">
          <article className="builder-panel builder-panel--hero">
            <div>
              <p className="builder-panel__eyebrow">Старт прохождения</p>
              <h2 className="builder-panel__title">{activeTest.title}</h2>
            </div>
            <div className="client-start-meta">
              <span>
                {activeTest.recommended_duration
                  ? `${activeTest.recommended_duration} мин`
                  : "Без лимита по времени"}
              </span>
              <span>{activeTest.questions?.length || 0} вопросов</span>
            </div>
          </article>

          <article className="builder-panel">
            <div className="builder-section-head">
              <div>
                <p className="builder-section-head__eyebrow">
                  Анкета респондента
                </p>
                <h3 className="builder-section-head__title">
                  Введите данные для начала
                </h3>
              </div>
            </div>

            <div className="client-info-banner">
              <strong>Важно</strong>
              <p>
                Результаты теста и итоговые рекомендации придут на почту,
                поэтому email лучше указать до начала прохождения.
              </p>
            </div>

            <div className="workflow-note">
              <p>
                Если тест уже запускался с тем же номером телефона, backend
                может вернуть незавершённую сессию и восстановить ваши ответы.
              </p>
            </div>

            <div className="client-author-access">
              <div>
                <p className="client-author-access__eyebrow">Автор теста</p>
                <strong className="client-author-access__title">
                  {publicAuthor
                    ? `Открыть карточку ${publicAuthor.full_name || "специалиста"}`
                    : "Открыть карточку автора теста"}
                </strong>
                <p className="client-author-access__text">
                  Перед стартом можно посмотреть карточку специалиста и понять,
                  кто проводит эту методику.
                </p>
              </div>
              <Link
                className="table-action-link"
                to={buildClientAuthorPath(slug)}
              >
                Карточка автора
              </Link>
            </div>

            <form className="admin-form-grid" onSubmit={handleStart}>
              <label className="admin-form-field">
                <span>ФИО</span>
                <input
                  aria-invalid={Boolean(startErrors.respondent_name)}
                  className={
                    startErrors.respondent_name
                      ? "admin-form-control admin-form-control--invalid"
                      : "admin-form-control"
                  }
                  value={startForm.respondent_name}
                  onChange={(event) =>
                    handleStartFieldChange(
                      "respondent_name",
                      event.target.value,
                    )
                  }
                  placeholder="Иванов Иван"
                />
                {startErrors.respondent_name ? (
                  <small className="admin-form-error admin-form-error--inline">
                    {startErrors.respondent_name}
                  </small>
                ) : null}
              </label>

              <label className="admin-form-field">
                <span>Телефон</span>
                <input
                  aria-invalid={Boolean(startErrors.respondent_phone)}
                  className={
                    startErrors.respondent_phone
                      ? "admin-form-control admin-form-control--invalid"
                      : "admin-form-control"
                  }
                  inputMode="tel"
                  maxLength={16}
                  value={startForm.respondent_phone}
                  onChange={(event) =>
                    handleStartFieldChange(
                      "respondent_phone",
                      event.target.value,
                    )
                  }
                  placeholder="+7 999 123-45-67"
                />
                {startErrors.respondent_phone ? (
                  <small className="admin-form-error admin-form-error--inline">
                    {startErrors.respondent_phone}
                  </small>
                ) : null}
              </label>

              <label className="admin-form-field admin-form-field--wide">
                <span>Email для результатов</span>
                <input
                  aria-invalid={Boolean(startErrors.respondent_email)}
                  className={
                    startErrors.respondent_email
                      ? "admin-form-control admin-form-control--invalid"
                      : "admin-form-control"
                  }
                  type="email"
                  value={startForm.respondent_email}
                  onChange={(event) =>
                    handleStartFieldChange(
                      "respondent_email",
                      event.target.value,
                    )
                  }
                  placeholder="Укажите почту для получения результатов"
                />
                {startErrors.respondent_email ? (
                  <small className="admin-form-error admin-form-error--inline">
                    {startErrors.respondent_email}
                  </small>
                ) : null}
              </label>

              {activeTest.collect_respondent_age ? (
                <label className="admin-form-field">
                  <span>Возраст</span>
                  <input
                    aria-invalid={Boolean(startErrors.respondent_age)}
                    className={
                      startErrors.respondent_age
                        ? "admin-form-control admin-form-control--invalid"
                        : "admin-form-control"
                    }
                    inputMode="numeric"
                    type="number"
                    value={startForm.respondent_age}
                    onChange={(event) =>
                      handleStartFieldChange(
                        "respondent_age",
                        event.target.value,
                      )
                    }
                    placeholder="18"
                  />
                  {startErrors.respondent_age ? (
                    <small className="admin-form-error admin-form-error--inline">
                      {startErrors.respondent_age}
                    </small>
                  ) : null}
                </label>
              ) : null}

              {activeTest.collect_respondent_gender ? (
                <label className="admin-form-field">
                  <span>Пол</span>
                  <select
                    aria-invalid={Boolean(startErrors.respondent_gender)}
                    className={
                      startErrors.respondent_gender
                        ? "admin-form-control admin-form-control--invalid"
                        : "admin-form-control"
                    }
                    value={startForm.respondent_gender}
                    onChange={(event) =>
                      handleStartFieldChange(
                        "respondent_gender",
                        event.target.value,
                      )
                    }
                  >
                    <option value="">Выберите</option>
                    <option value="male">Мужской</option>
                    <option value="female">Женский</option>
                  </select>
                  {startErrors.respondent_gender ? (
                    <small className="admin-form-error admin-form-error--inline">
                      {startErrors.respondent_gender}
                    </small>
                  ) : null}
                </label>
              ) : null}

              {activeTest.collect_respondent_education ? (
                <label className="admin-form-field">
                  <span>Образование</span>
                  <input
                    aria-invalid={Boolean(startErrors.respondent_education)}
                    className={
                      startErrors.respondent_education
                        ? "admin-form-control admin-form-control--invalid"
                        : "admin-form-control"
                    }
                    value={startForm.respondent_education}
                    onChange={(event) =>
                      handleStartFieldChange(
                        "respondent_education",
                        event.target.value,
                      )
                    }
                    placeholder="Среднее, бакалавриат, магистратура..."
                  />
                  {startErrors.respondent_education ? (
                    <small className="admin-form-error admin-form-error--inline">
                      {startErrors.respondent_education}
                    </small>
                  ) : null}
                </label>
              ) : null}

              <div className="admin-form-actions">
                <button
                  className="admin-primary-button"
                  disabled={isStarting}
                  type="submit"
                >
                  {isStarting ? (
                    <LoaderCircle
                      className="icon-spin"
                      size={16}
                      strokeWidth={2.1}
                    />
                  ) : (
                    <Send size={16} strokeWidth={2.1} />
                  )}
                  <span>{isStarting ? "Запуск..." : "Начать тест"}</span>
                </button>
              </div>
            </form>
          </article>
        </section>
      ) : null}

      {activeTest && session ? (
        <section className="client-flow-grid">
          <article className="builder-panel builder-panel--hero">
            <div>
              <p className="builder-panel__eyebrow">Прохождение теста</p>
              <h2 className="builder-panel__title">{activeTest.title}</h2>
            </div>
            <div className="client-start-meta">
              <span>
                {answeredCount} из {questionCount} заполнено
              </span>
              <span>{completionPercent}% прогресса</span>
            </div>
          </article>

          <article className="builder-panel">
            <div className="client-progress-bar">
              <div
                className="client-progress-bar__fill"
                style={{ width: `${completionPercent}%` }}
              />
            </div>

            <div
              className={`workflow-note ${hasRespondentEmail ? "workflow-note--success" : "workflow-note--warning"}`}
            >
              <p>
                {hasRespondentEmail
                  ? `Результаты будут связаны с почтой ${startForm.respondent_email}.`
                  : "Почта для результатов не указана. Если нужен итог на email, лучше вернуться и заполнить её до старта новой сессии."}{" "}
                Прогресс можно сохранять и продолжать позже. Черновик также
                сохраняется автоматически после паузы.
              </p>
            </div>

            {autoSaveState ? (
              <p className="admin-form-message">{autoSaveState}</p>
            ) : null}

            <div className="client-author-access client-author-access--compact">
              <div>
                <p className="client-author-access__eyebrow">Автор теста</p>
                <strong className="client-author-access__title">
                  {publicAuthor?.full_name || "Карточка специалиста"}
                </strong>
              </div>
              <Link
                className="table-action-link"
                to={buildClientAuthorPath(slug)}
              >
                Открыть карточку
              </Link>
            </div>

            <div className="client-question-list">
              {activeTest.questions.map((question, index) => (
                <section className="client-question-card" key={question.id}>
                  <div className="client-question-card__head">
                    <strong>Вопрос {index + 1}</strong>
                    {question.is_required ? (
                      <span className="status-badge status-badge--draft">
                        Обязательный
                      </span>
                    ) : null}
                  </div>
                  <p className="client-question-card__title">{question.text}</p>
                  {renderQuestion(question)}
                </section>
              ))}
            </div>

            <div className="client-flow-actions">
              <button
                className="table-action-button"
                disabled={isSaving || isSubmitting}
                type="button"
                onClick={handleSaveProgress}
              >
                {isSaving ? (
                  <LoaderCircle
                    className="icon-spin"
                    size={15}
                    strokeWidth={2.1}
                  />
                ) : (
                  <Save size={15} strokeWidth={2.1} />
                )}
                <span>{isSaving ? "Сохранение..." : "Сохранить прогресс"}</span>
              </button>
              <button
                className="admin-primary-button"
                disabled={isSubmitting}
                type="button"
                onClick={handleSubmit}
              >
                {isSubmitting ? (
                  <LoaderCircle
                    className="icon-spin"
                    size={16}
                    strokeWidth={2.1}
                  />
                ) : (
                  <CheckCircle2 size={16} strokeWidth={2.1} />
                )}
                <span>{isSubmitting ? "Отправка..." : "Завершить тест"}</span>
              </button>
            </div>
          </article>
        </section>
      ) : null}

      {!publicTestQuery.isLoading && !activeTest && !publicTestQuery.error ? (
        <div className="builder-empty">
          По этой ссылке тест не найден.{" "}
          <Link to={ROUTES.root}>Вернуться на главную</Link>
        </div>
      ) : null}
    </PageCard>
  );
}
