export function formatRussianPhoneInput(value) {
  const digits = String(value || "").replace(/\D/g, "");
  const normalized = digits.startsWith("8") ? `7${digits.slice(1)}` : digits;
  const body = normalized.startsWith("7") ? normalized.slice(1, 11) : normalized.slice(0, 10);

  let result = "+7";

  if (body.length > 0) {
    result += ` ${body.slice(0, 3)}`;
  }

  if (body.length >= 4) {
    result += ` ${body.slice(3, 6)}`;
  }

  if (body.length >= 7) {
    result += `-${body.slice(6, 8)}`;
  }

  if (body.length >= 9) {
    result += `-${body.slice(8, 10)}`;
  }

  return result;
}

export function isRussianPhoneComplete(value) {
  return /^\+7 \d{3} \d{3}-\d{2}-\d{2}$/.test(String(value || "").trim());
}
