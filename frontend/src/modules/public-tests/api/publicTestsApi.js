import { requestJson } from "../../../shared/api/http";

function encodeSlug(slug) {
  return encodeURIComponent(String(slug || "").trim());
}

export function getPublicTestRequest(slug) {
  return requestJson(`/public/tests/${encodeSlug(slug)}`);
}

export function startPublicTestRequest(slug, payload) {
  return requestJson(`/public/tests/${encodeSlug(slug)}/start`, {
    method: "POST",
    body: payload,
  });
}

export function savePublicTestProgressRequest(slug, payload) {
  return requestJson(`/public/tests/${encodeSlug(slug)}/progress`, {
    method: "POST",
    body: payload,
  });
}

export function submitPublicTestRequest(slug, payload) {
  return requestJson(`/public/tests/${encodeSlug(slug)}/submit`, {
    method: "POST",
    body: payload,
  });
}
