import { useEffect } from "react";

const APP_TITLE = "ПрофДНК";

export function buildDocumentTitle(title) {
  const normalizedTitle = String(title || "").trim();

  if (!normalizedTitle) {
    return APP_TITLE;
  }

  return `${normalizedTitle} — ${APP_TITLE}`;
}

export default function useDocumentTitle(title) {
  useEffect(() => {
    document.title = buildDocumentTitle(title);
  }, [title]);
}
