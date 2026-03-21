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

function createHttpError(response, body) {
  const fallbackMessage = response.status
    ? `HTTP ${response.status}${response.statusText ? ` ${response.statusText}` : ""}`
    : "Не удалось выполнить запрос";
  const fieldErrors = body?.field_errors && typeof body.field_errors === "object"
    ? Object.entries(body.field_errors)
        .filter(([, value]) => String(value || "").trim())
        .map(([field, value]) => `${field}: ${value}`)
    : [];
  const primaryMessage =
    body?.error ||
    body?.message ||
    (fieldErrors.length ? fieldErrors[0] : "") ||
    fallbackMessage;
  const error = new Error(primaryMessage);
  error.status = response.status;
  error.body = body;
  error.fieldErrors = body?.field_errors || null;
  error.details = fieldErrors;
  return error;
}

async function parseResponse(response) {
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
    throw createHttpError(response, body);
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

  return parseResponse(response);
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

    throw createHttpError(response, body);
  }

  return {
    blob: await response.blob(),
    contentType: response.headers.get("content-type") || "",
    filename: parseFilenameFromDisposition(response.headers.get("content-disposition")),
  };
}
