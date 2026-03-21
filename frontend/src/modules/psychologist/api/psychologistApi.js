import { requestJson } from "../../../shared/api/http";

export function updatePsychologistProfileRequest(payload) {
  return requestJson("/psychologists/me/profile", {
    method: "PUT",
    body: payload,
  });
}

export function updatePsychologistCardRequest(payload) {
  return requestJson("/psychologists/me/card", {
    method: "PUT",
    body: payload,
  });
}
