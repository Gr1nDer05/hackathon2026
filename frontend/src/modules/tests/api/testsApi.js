import { requestFile, requestJson } from "../../../shared/api/http";
import { normalizeResultsCollection } from "../../psychologist/lib/psychologistUi";

export async function listPsychologistTestsRequest() {
  const data = await requestJson("/psychologists/tests");
  return Array.isArray(data) ? data : [];
}

export async function createPsychologistTestRequest(payload) {
  return requestJson("/psychologists/tests", {
    method: "POST",
    body: payload,
  });
}

export async function getPsychologistTestRequest(id) {
  return requestJson(`/psychologists/tests/${id}`);
}

export async function updatePsychologistTestRequest(id, payload) {
  return requestJson(`/psychologists/tests/${id}`, {
    method: "PUT",
    body: payload,
  });
}

export async function publishPsychologistTestRequest(id) {
  return requestJson(`/psychologists/tests/${id}/publish`, {
    method: "POST",
  });
}

export async function listPsychologistQuestionsRequest(testId) {
  const data = await requestJson(`/psychologists/tests/${testId}/questions`);
  return Array.isArray(data) ? data : [];
}

export async function createPsychologistQuestionRequest(testId, payload) {
  return requestJson(`/psychologists/tests/${testId}/questions`, {
    method: "POST",
    body: payload,
  });
}

export async function updatePsychologistQuestionRequest(testId, questionId, payload) {
  return requestJson(`/psychologists/tests/${testId}/questions/${questionId}`, {
    method: "PUT",
    body: payload,
  });
}

export async function deletePsychologistQuestionRequest(testId, questionId) {
  return requestJson(`/psychologists/tests/${testId}/questions/${questionId}`, {
    method: "DELETE",
  });
}

export async function listPsychologistTestResultsRequest(testId) {
  const data = await requestJson(`/psychologists/tests/${testId}/results`);
  return normalizeResultsCollection(data);
}

export async function getPsychologistTestSubmissionRequest(testId, sessionId) {
  return requestJson(`/psychologists/tests/${testId}/results/${sessionId}`);
}

export async function listFormulaRulesRequest(testId) {
  const data = await requestJson(`/psychologists/tests/${testId}/formulas`);
  return Array.isArray(data) ? data : [];
}

export async function createFormulaRuleRequest(testId, payload) {
  return requestJson(`/psychologists/tests/${testId}/formulas`, {
    method: "POST",
    body: payload,
  });
}

export async function updateFormulaRuleRequest(testId, ruleId, payload) {
  return requestJson(`/psychologists/tests/${testId}/formulas/${ruleId}`, {
    method: "PUT",
    body: payload,
  });
}

export async function deleteFormulaRuleRequest(testId, ruleId) {
  return requestJson(`/psychologists/tests/${testId}/formulas/${ruleId}`, {
    method: "DELETE",
  });
}

export async function calculateFormulaPreviewRequest(testId, payload) {
  return requestJson(`/psychologists/tests/${testId}/formulas/calculate`, {
    method: "POST",
    body: payload,
  });
}

export async function getPsychologistSubmissionReportRequest(sessionId, { format = "html", audience = "psychologist" } = {}) {
  const query = new URLSearchParams({
    format,
    audience,
  });

  const file = await requestFile(`/psychologists/results/${sessionId}/report?${query.toString()}`);
  const extension = format === "docx" ? "docx" : "html";

  return {
    ...file,
    filename: file.filename || `report-${audience}-${sessionId}.${extension}`,
  };
}

export async function listPsychologistReportTemplatesRequest() {
  const data = await requestJson("/psychologists/report-templates");
  return Array.isArray(data) ? data : [];
}

export async function createPsychologistReportTemplateRequest(payload) {
  return requestJson("/psychologists/report-templates", {
    method: "POST",
    body: payload,
  });
}

export async function updatePsychologistReportTemplateRequest(templateId, payload) {
  return requestJson(`/psychologists/report-templates/${templateId}`, {
    method: "PUT",
    body: payload,
  });
}

export async function deletePsychologistReportTemplateRequest(templateId) {
  return requestJson(`/psychologists/report-templates/${templateId}`, {
    method: "DELETE",
  });
}
