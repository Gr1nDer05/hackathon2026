import { formatRussianPhoneInput } from "../../../shared/lib/russianPhone";

export function formatDate(value) {
  if (!value) {
    return "—";
  }

  const timestamp = new Date(value).getTime();
  if (!Number.isFinite(timestamp)) {
    return "—";
  }

  return new Date(value).toLocaleDateString("ru-RU", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
  });
}

export function isFutureDate(value) {
  if (!value) {
    return false;
  }

  const timestamp = new Date(value).getTime();
  return Number.isFinite(timestamp) && timestamp > Date.now();
}

export function getPsychologistDisplayName(user) {
  const profile = user?.profile || {};
  const candidates = [
    profile.full_name,
    profile.name,
    profile.fullName,
    user?.full_name,
    user?.name,
    user?.login,
    user?.email,
  ];

  return candidates.find((value) => String(value || "").trim()) || "Психолог";
}

export function getPsychologistInitials(user) {
  const displayName = getPsychologistDisplayName(user);
  const parts = String(displayName)
    .trim()
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2);

  if (!parts.length) {
    return "П";
  }

  return parts.map((part) => part[0]?.toUpperCase() || "").join("");
}

export function getPsychologistSpecialization(user) {
  return (
    user?.profile?.specialization ||
    user?.card?.specialization ||
    user?.specialization ||
    "Специализация не указана"
  );
}

export function getPsychologistPhone(user) {
  const phone =
    user?.card?.contact_phone ||
    user?.card?.contactPhone ||
    user?.profile?.contact_phone ||
    user?.profile?.contactPhone ||
    user?.profile?.phone ||
    user?.card?.phone ||
    user?.phone ||
    "";

  return String(phone || "").trim() ? formatRussianPhoneInput(phone) : "—";
}

export function getPsychologistCity(user) {
  return user?.profile?.city || user?.card?.city || user?.city || "—";
}

export function getPsychologistEducation(user) {
  return (
    user?.profile?.education ||
    user?.card?.education ||
    user?.education ||
    "Не указано"
  );
}

export function getPsychologistJoinedAt(user) {
  return (
    user?.profile?.created_at ||
    user?.workspace?.created_at ||
    user?.created_at ||
    user?.registered_at ||
    ""
  );
}

export function getPsychologistLastActivityAt(user, tests = []) {
  return (
    user?.workspace?.last_activity_at ||
    user?.profile?.last_activity_at ||
    user?.updated_at ||
    tests[0]?.updated_at ||
    tests[0]?.created_at ||
    ""
  );
}

export function getPsychologistAccessUntil(user) {
  return user?.portal_access_until || user?.subscription_expires_at || user?.expires_at || "";
}

export function getAccountStatus(user) {
  if (!user) {
    return "blocked";
  }

  if (user.account_status) {
    return user.account_status;
  }

  if (user.is_active === false) {
    return "blocked";
  }

  if (isFutureDate(user.blocked_until)) {
    return "blocked";
  }

  return "active";
}

export function getSubscriptionStatus(user) {
  if (!user) {
    return "expired";
  }

  if (user.subscription_status) {
    return user.subscription_status;
  }

  const accessUntil = getPsychologistAccessUntil(user);
  if (!accessUntil) {
    return "active";
  }

  return isFutureDate(accessUntil) ? "active" : "expired";
}

export function getStatusTone(status) {
  if (status === "active" || status === "published" || status === "completed") {
    return "active";
  }

  if (status === "trial" || status === "in_progress") {
    return "trial";
  }

  if (status === "draft") {
    return "draft";
  }

  if (status === "blocked") {
    return "blocked";
  }

  return "expired";
}

export function getAccountStatusLabel(status) {
  if (status === "active") {
    return "Аккаунт активен";
  }

  if (status === "blocked") {
    return "Аккаунт заблокирован";
  }

  return status || "Статус не определён";
}

export function getSubscriptionStatusLabel(status) {
  if (status === "active") {
    return "Подписка активна";
  }

  if (status === "trial") {
    return "Пробный период";
  }

  if (status === "expired") {
    return "Подписка истекла";
  }

  if (status === "blocked") {
    return "Доступ заблокирован";
  }

  return status || "Статус подписки неизвестен";
}

export function sortTestsByRecent(tests = []) {
  return [...tests].sort((left, right) => {
    const leftTimestamp = new Date(getTestActivityAt(left) || 0).getTime();
    const rightTimestamp = new Date(getTestActivityAt(right) || 0).getTime();
    return rightTimestamp - leftTimestamp;
  });
}

export function getTestActivityAt(test) {
  return (
    test?.last_activity_at ||
    test?.last_completed_at ||
    test?.last_started_at ||
    test?.updated_at ||
    test?.created_at ||
    ""
  );
}

export function summarizeTests(tests = []) {
  const normalizedTests = Array.isArray(tests) ? tests : [];
  const publishedCount = normalizedTests.filter((item) => item.status === "published").length;
  const draftCount = normalizedTests.filter((item) => item.status !== "published").length;
  const startedSessions = normalizedTests.reduce(
    (total, item) => total + Number(item?.started_sessions_count || 0),
    0,
  );
  const inProgressSessions = normalizedTests.reduce(
    (total, item) => total + Number(item?.in_progress_sessions_count || 0),
    0,
  );
  const completedSessions = normalizedTests.reduce(
    (total, item) => total + Number(item?.completed_sessions_count || 0),
    0,
  );
  const lastActivityAt = normalizedTests
    .map((item) => getTestActivityAt(item))
    .filter(Boolean)
    .sort((left, right) => new Date(right).getTime() - new Date(left).getTime())[0] || "";

  return {
    totalCount: normalizedTests.length,
    publishedCount,
    draftCount,
    startedSessions,
    inProgressSessions,
    completedSessions,
    lastActivityAt,
  };
}

export function normalizeResultsCollection(data) {
  if (Array.isArray(data)) {
    return data;
  }

  if (Array.isArray(data?.items)) {
    return data.items;
  }

  if (Array.isArray(data?.results)) {
    return data.results;
  }

  if (Array.isArray(data?.sessions)) {
    return data.sessions;
  }

  return [];
}

const LEGACY_METRIC_LABELS = {
  analytic: "Аналитика",
  creative: "Креативность",
  social: "Социальность",
  organizer: "Организация",
  practical: "Практичность",
};

function toFiniteNumber(value) {
  const numeric = Number(value);
  return Number.isFinite(numeric) ? numeric : null;
}

function formatMetricValue(value) {
  const numeric = toFiniteNumber(value);

  if (numeric === null) {
    return "—";
  }

  return Number.isInteger(numeric) ? String(numeric) : numeric.toFixed(1);
}

export function getMetricLabel(key) {
  const normalizedKey = String(key || "").trim();

  if (!normalizedKey) {
    return "Метрика";
  }

  if (LEGACY_METRIC_LABELS[normalizedKey]) {
    return LEGACY_METRIC_LABELS[normalizedKey];
  }

  return normalizedKey
    .replace(/([a-z])([A-Z])/g, "$1 $2")
    .replace(/[_-]+/g, " ")
    .replace(/\s+/g, " ")
    .trim()
    .replace(/^\p{L}/u, (letter) => letter.toUpperCase());
}

export function getResultMetricEntries(resultPayload) {
  const directMetrics = resultPayload?.metrics;

  if (directMetrics && typeof directMetrics === "object" && !Array.isArray(directMetrics)) {
    const entries = Object.entries(directMetrics)
      .map(([key, value]) => ({
        key,
        numericValue: toFiniteNumber(value),
      }))
      .filter((item) => item.numericValue !== null);

    const maxValue = entries.reduce((max, item) => Math.max(max, Math.abs(item.numericValue)), 0);

    return entries
      .sort((left, right) => Math.abs(right.numericValue) - Math.abs(left.numericValue))
      .map((item) => ({
        key: item.key,
        label: getMetricLabel(item.key),
        value: item.numericValue,
        displayValue: formatMetricValue(item.numericValue),
        progress: maxValue ? Math.round((Math.abs(item.numericValue) / maxValue) * 100) : 0,
        meta: "Рассчитано по rules",
      }));
  }

  const scales = Array.isArray(resultPayload?.career_result?.scales)
    ? resultPayload.career_result.scales
    : Array.isArray(resultPayload?.scales)
      ? resultPayload.scales
      : [];

  return scales
    .map((item) => ({
      key: item.scale,
      label: getMetricLabel(item.scale),
      value: toFiniteNumber(item.raw_score ?? item.percentage ?? 0) ?? 0,
      displayValue:
        item.percentage !== null && item.percentage !== undefined
          ? `${Math.round(Number(item.percentage) || 0)}%`
          : formatMetricValue(item.raw_score),
      progress:
        item.percentage !== null && item.percentage !== undefined
          ? Math.max(0, Math.min(100, Math.round(Number(item.percentage) || 0)))
          : 0,
      meta:
        item.raw_score !== null &&
        item.raw_score !== undefined &&
        item.max_score !== null &&
        item.max_score !== undefined
          ? `${formatMetricValue(item.raw_score)} из ${formatMetricValue(item.max_score)} баллов`
          : "Рассчитано backend",
    }))
    .sort((left, right) => right.progress - left.progress);
}

export function getResultTopMetricEntries(resultPayload) {
  const topMetrics = Array.isArray(resultPayload?.top_metrics)
    ? resultPayload.top_metrics
    : Array.isArray(resultPayload?.career_result?.top_scales)
      ? resultPayload.career_result.top_scales
      : [];

  return topMetrics.map((item, index) => {
    const key = item.key || item.metric || item.result_key || item.scale || `metric-${index + 1}`;
    const numericValue =
      toFiniteNumber(item.value) ??
      toFiniteNumber(item.score) ??
      toFiniteNumber(item.raw_score) ??
      toFiniteNumber(item.percentage) ??
      0;

    return {
      key,
      label: getMetricLabel(key),
      value: numericValue,
      displayValue:
        item.percentage !== null && item.percentage !== undefined
          ? `${Math.round(Number(item.percentage) || 0)}%`
          : formatMetricValue(numericValue),
    };
  });
}

export function getResultProfessionEntries(resultPayload) {
  const professions = Array.isArray(resultPayload?.career_result?.top_professions)
    ? resultPayload.career_result.top_professions
    : Array.isArray(resultPayload?.top_professions)
      ? resultPayload.top_professions
      : [];

  return professions.map((item) => ({
    profession: item.profession || item.label || item.name || "Без названия",
    score: formatMetricValue(item.score),
  }));
}

export function summarizeMetricEntries(entries = [], limit = 2) {
  return (Array.isArray(entries) ? entries : [])
    .slice(0, limit)
    .map((item) => item.label || getMetricLabel(item.key))
    .filter(Boolean)
    .join(", ");
}

export function normalizeResultItem(item, index) {
  const completedAt = item?.completed_at || item?.submitted_at || item?.finished_at || "";
  const startedAt = item?.started_at || item?.created_at || item?.startedAt || "";
  const rawStatus =
    item?.status ||
    item?.session_status ||
    item?.state ||
    (completedAt ? "completed" : "in_progress");
  const progress = Number(
    item?.progress_percentage ??
      item?.progress ??
      item?.completion_percentage ??
      (completedAt ? 100 : 0),
  );
  const metrics = getResultMetricEntries(item);
  const topMetrics = getResultTopMetricEntries(item);

  return {
    id: item?.id || item?.session_id || item?.result_id || `result-${index + 1}`,
    respondent:
      item?.respondent_name ||
      item?.respondent?.full_name ||
      item?.respondent?.name ||
      item?.email ||
      item?.respondent_email ||
      `Респондент ${index + 1}`,
    email: item?.respondent_email || item?.email || item?.respondent?.email || "",
    status: rawStatus,
    progress: Number.isFinite(progress) ? Math.max(0, Math.min(100, Math.round(progress))) : 0,
    startedAt,
    completedAt,
    metrics,
    topMetrics,
    topMetricSummary: summarizeMetricEntries(topMetrics.length ? topMetrics : metrics),
  };
}
