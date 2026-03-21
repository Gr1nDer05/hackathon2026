export const ROUTES = {
  root: "/",

  // Public client flow
  clientSession: "/session/:slug",
  clientAuthor: "/session/:slug/author",
  clientResult: "/result/:slug",

  // Psychologist
  dashboard: "/dashboard",
  tests: "/tests",
  testBuilder: "/builder/:id",
  testResults: "/tests/:id/results",
  testSubmission: "/tests/:id/results/:sessionId",
  reportTemplates: "/report-templates",
  profile: "/profile",

  // Admin
  adminDashboard: "/admin",
  adminPsychologists: "/admin/psychologists",
  adminPsychologistById: "/admin/psychologists/:id",
  adminSubscriptions: "/admin/subscriptions",

  // Common
  subscriptionRequired: "/subscription-required",
  forbidden: "/forbidden",
  notFound: "*",
};

export function buildClientSessionPath(slug) {
  return ROUTES.clientSession.replace(":slug", encodeURIComponent(String(slug)));
}

export function buildClientResultPath(slug) {
  return ROUTES.clientResult.replace(":slug", encodeURIComponent(String(slug)));
}

export function buildClientAuthorPath(slug) {
  return ROUTES.clientAuthor.replace(":slug", encodeURIComponent(String(slug)));
}

export function buildTestSubmissionPath(id, sessionId) {
  return ROUTES.testSubmission
    .replace(":id", encodeURIComponent(String(id)))
    .replace(":sessionId", encodeURIComponent(String(sessionId)));
}
