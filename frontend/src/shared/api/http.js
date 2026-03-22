const FALLBACK_API_URL = "http://localhost:8080";

function resolveApiBaseUrl() {
  if (import.meta.env.DEV) {
    return "/api";
  }

  return import.meta.env.VITE_API_URL || FALLBACK_API_URL;
}

export const API_BASE_URL = resolveApiBaseUrl();

function getCookie(name) {
  const target = `${name}=`;
  const parts = document.cookie.split("; ");
  const matched = parts.find((part) => part.startsWith(target));
  return matched ? decodeURIComponent(matched.slice(target.length)) : "";
}

function buildRequestHeaders(method, hasBody, extraHeaders) {
  const headers = new Headers(extraHeaders || {});

  if (hasBody && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  if (method !== "GET" && method !== "HEAD") {
    const csrfToken = getCookie("csrf_token");

    if (csrfToken) {
      headers.set("X-CSRF-Token", csrfToken);
    }
  }

  return headers;
}

function isTechnicalMessage(message) {
  return /backend|endpoint|swagger|payload|json|server|openai|openrouter|http\s*\d+|cors|report endpoint|legacy/i.test(
    String(message || ""),
  );
}

const FIELD_LABELS = {
  email: "email",
  contact_email: "email",
  respondent_email: "email",
  password: "пароль",
  login: "логин",
  full_name: "ФИО",
  respondent_name: "имя и фамилия",
  respondent_phone: "номер телефона",
  contact_phone: "номер телефона",
  phone: "номер телефона",
  title: "название",
  name: "название",
  description: "описание",
  about: "описание",
  specialization: "специализация",
  city: "город",
  code: "код",
  test_id: "тест",
  report_template_id: "шаблон отчёта",
  question_type: "тип вопроса",
  recommended_duration: "длительность",
  max_participants: "лимит участников",
  prompt: "описание шаблона",
};

function looksLikeDeveloperMessage(message) {
  const text = String(message || "").trim();

  if (!text) {
    return true;
  }

  if (isTechnicalMessage(text)) {
    return true;
  }

  if (/[a-z]/i.test(text) && !/[а-яё]/i.test(text)) {
    return true;
  }

  return /minlength|maxlength|pattern|required|invalid|must|failed|rfc|uuid|regex|format|parse|json/i.test(text);
}

function getFieldLabel(field) {
  return FIELD_LABELS[field] || "это поле";
}

function buildFriendlyFieldMessage(field, value) {
  const raw = String(value || "").trim();
  const label = getFieldLabel(field);

  if (!raw) {
    return "";
  }

  if (/already exists|already used|duplicate|taken/i.test(raw)) {
    if (field === "email" || field === "contact_email" || field === "respondent_email") {
      return "Этот email уже используется.";
    }

    return `Поле «${label}» уже используется.`;
  }

  if (!looksLikeDeveloperMessage(raw)) {
    return raw;
  }

  if (field === "email" || field === "contact_email" || field === "respondent_email") {
    return "Проверьте адрес электронной почты.";
  }

  if (field === "password") {
    return "Проверьте пароль. Он должен соответствовать требованиям формы.";
  }

  if (field === "login") {
    return "Проверьте логин.";
  }

  if (field === "full_name" || field === "respondent_name") {
    return "Проверьте имя и фамилию.";
  }

  if (field === "respondent_phone" || field === "contact_phone" || field === "phone") {
    return "Проверьте номер телефона.";
  }

  if (field === "title" || field === "name") {
    return "Заполните название.";
  }

  if (field === "description" || field === "about") {
    return "Проверьте описание.";
  }

  if (field === "specialization") {
    return "Проверьте специализацию.";
  }

  if (field === "city") {
    return "Проверьте город.";
  }

  if (field === "code") {
    return "Проверьте код.";
  }

  if (field === "test_id") {
    return "Выберите тест.";
  }

  if (field === "report_template_id") {
    return "Выберите шаблон отчёта.";
  }

  if (field === "prompt") {
    return "Опишите, каким должен быть шаблон отчёта.";
  }

  if (field === "recommended_duration") {
    return "Проверьте длительность теста.";
  }

  if (field === "max_participants") {
    return "Проверьте лимит участников.";
  }

  return `Проверьте поле «${label}».`;
}

function buildUserMessage({ status, body, fieldErrors, fallbackMessage, path }) {
  const firstFieldEntry = body?.field_errors && typeof body.field_errors === "object"
    ? Object.entries(body.field_errors).find(([, value]) => String(value || "").trim())
    : null;
  const rawMessage = body?.error || body?.message || "";

  if (path.includes("/psychologists/report-templates/generate")) {
    return "Проблемы с OpenRouter, попробуйте ещё раз.";
  }

  if (firstFieldEntry) {
    return buildFriendlyFieldMessage(firstFieldEntry[0], firstFieldEntry[1]);
  }

  if (rawMessage && !looksLikeDeveloperMessage(rawMessage)) {
    return String(rawMessage);
  }

  if (status === 400) {
    return "Проверьте заполненные данные и попробуйте ещё раз.";
  }

  if (status === 401) {
    return "Сессия завершилась. Войдите ещё раз и повторите попытку.";
  }

  if (status === 403) {
    return "Сейчас это действие недоступно.";
  }

  if (status === 404) {
    return "Нужные данные не найдены.";
  }

  if (status === 409) {
    return "Сейчас это действие выполнить нельзя. Попробуйте ещё раз чуть позже.";
  }

  if (status === 429) {
    return "Слишком много попыток. Подождите немного и попробуйте снова.";
  }

  if (status >= 500) {
    return "Сервис временно недоступен. Попробуйте ещё раз.";
  }

  return fallbackMessage;
}

function createHttpError(response, body, path = "") {
  const fallbackMessage = response.status
    ? `HTTP ${response.status}${response.statusText ? ` ${response.statusText}` : ""}`
    : "Не удалось выполнить запрос";
  const fieldErrors = body?.field_errors && typeof body.field_errors === "object"
    ? Object.entries(body.field_errors)
        .filter(([, value]) => String(value || "").trim())
        .map(([field, value]) => `${field}: ${value}`)
    : [];
  const rawMessage =
    body?.error ||
    body?.message ||
    (fieldErrors.length ? fieldErrors[0] : "") ||
    fallbackMessage;
  const userMessage = buildUserMessage({
    status: response.status,
    body,
    fieldErrors,
    fallbackMessage: "Не удалось выполнить запрос. Попробуйте ещё раз.",
    path,
  });
  const error = new Error(userMessage);
  error.status = response.status;
  error.body = body;
  error.fieldErrors = body?.field_errors || null;
  error.details = fieldErrors;
  error.rawMessage = rawMessage;
  return error;
}

async function parseResponse(response, path) {
  const text = await response.text();
  let body = {};

  if (text) {
    try {
      body = JSON.parse(text);
    } catch {
      body = {};
    }
  }

  if (!response.ok) {
    throw createHttpError(response, body, path);
  }

  return body;
}

function parseFilenameFromDisposition(disposition) {
  const raw = String(disposition || "");

  const utf8Match = raw.match(/filename\*=UTF-8''([^;]+)/i);
  if (utf8Match?.[1]) {
    return decodeURIComponent(utf8Match[1]);
  }

  const simpleMatch = raw.match(/filename="?([^"]+)"?/i);
  return simpleMatch?.[1] || "";
}

export async function requestJson(path, { method = "GET", body, headers } = {}) {
  const hasBody = body !== undefined;
  const response = await fetch(`${API_BASE_URL}${path}`, {
    method,
    credentials: "include",
    headers: buildRequestHeaders(method, hasBody, headers),
    body: hasBody ? JSON.stringify(body) : undefined,
  });

  return parseResponse(response, path);
}

export async function requestFile(path, { method = "GET", headers } = {}) {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    method,
    credentials: "include",
    headers: buildRequestHeaders(method, false, headers),
  });

  if (!response.ok) {
    let body = {};

    try {
      body = await response.json();
    } catch {
      body = {};
    }

    throw createHttpError(response, body, path);
  }

  return {
    blob: await response.blob(),
    contentType: response.headers.get("content-type") || "",
    filename: parseFilenameFromDisposition(response.headers.get("content-disposition")),
  };
}
