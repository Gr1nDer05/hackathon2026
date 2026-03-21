const STORAGE_PREFIX = "public-test-session";

function getStorageKey(slug) {
  return `${STORAGE_PREFIX}:${String(slug || "").trim()}`;
}

export function readPublicTestSnapshot(slug) {
  if (typeof window === "undefined") {
    return null;
  }

  try {
    const raw = window.localStorage.getItem(getStorageKey(slug));
    return raw ? JSON.parse(raw) : null;
  } catch {
    return null;
  }
}

export function writePublicTestSnapshot(slug, payload) {
  if (typeof window === "undefined") {
    return;
  }

  try {
    window.localStorage.setItem(getStorageKey(slug), JSON.stringify(payload));
  } catch {
    // Ignore storage quota or privacy mode failures.
  }
}

export function clearPublicTestSnapshot(slug) {
  if (typeof window === "undefined") {
    return;
  }

  try {
    window.localStorage.removeItem(getStorageKey(slug));
  } catch {
    // Ignore storage cleanup failures.
  }
}

export function buildAnswersMap(answers = []) {
  return (Array.isArray(answers) ? answers : []).reduce((accumulator, item) => {
    const questionId = String(item?.question_id || "");

    if (!questionId) {
      return accumulator;
    }

    accumulator[questionId] = {
      answer_text: item?.answer_text || "",
      answer_value: item?.answer_value || "",
      answer_values: Array.isArray(item?.answer_values) ? item.answer_values : [],
    };

    return accumulator;
  }, {});
}

export function buildAnswersPayload(answersByQuestion = {}) {
  return Object.entries(answersByQuestion)
    .map(([questionId, value]) => {
      const answer = {
        question_id: Number(questionId),
      };

      if (value?.answer_text) {
        answer.answer_text = String(value.answer_text);
      }

      if (value?.answer_value !== undefined && value?.answer_value !== null && value?.answer_value !== "") {
        answer.answer_value = String(value.answer_value);
      }

      if (Array.isArray(value?.answer_values) && value.answer_values.length) {
        answer.answer_values = value.answer_values.map((item) => String(item));
      }

      return answer;
    })
    .filter((item) => {
      return (
        item.answer_text ||
        item.answer_value ||
        (Array.isArray(item.answer_values) && item.answer_values.length)
      );
    });
}
