const STORAGE_KEY = "frontend-demo-subscriptions";
const DAY_IN_MS = 24 * 60 * 60 * 1000;

function readStore() {
  if (typeof window === "undefined") {
    return {};
  }

  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (!raw) {
      return {};
    }

    const parsed = JSON.parse(raw);
    return parsed && typeof parsed === "object" ? parsed : {};
  } catch {
    return {};
  }
}

function writeStore(store) {
  if (typeof window === "undefined") {
    return;
  }

  window.localStorage.setItem(STORAGE_KEY, JSON.stringify(store));
}

function buildUserKey(user) {
  if (!user || user.role !== "psychologist" || !user.id) {
    return "";
  }

  return `psychologist:${user.id}`;
}

function isActiveEntry(entry) {
  const expiresAt = new Date(entry?.expires_at || "").getTime();
  return Number.isFinite(expiresAt) && expiresAt > Date.now();
}

export function getMockSubscription(user) {
  const key = buildUserKey(user);
  if (!key) {
    return null;
  }

  const store = readStore();
  const entry = store[key];

  if (!entry) {
    return null;
  }

  if (!isActiveEntry(entry)) {
    delete store[key];
    writeStore(store);
    return null;
  }

  return entry;
}

export function applyMockSubscription(user) {
  const entry = getMockSubscription(user);
  if (!entry) {
    return user;
  }

  return {
    ...user,
    subscription_status: "active",
    portal_access_until: entry.expires_at,
    subscription_plan: entry.subscription_plan || user?.subscription_plan || "basic",
    mock_subscription: {
      purchased_at: entry.purchased_at,
      source: entry.source,
    },
  };
}

export function activateMockSubscription(user, { days = 30, subscriptionPlan } = {}) {
  const key = buildUserKey(user);
  if (!key) {
    return null;
  }

  const store = readStore();
  const previousEntry = store[key];
  const now = Date.now();
  const previousExpiry = new Date(previousEntry?.expires_at || "").getTime();
  const baseTimestamp = Number.isFinite(previousExpiry) && previousExpiry > now ? previousExpiry : now;
  const expiresAt = new Date(baseTimestamp + days * DAY_IN_MS).toISOString();

  const nextEntry = {
    status: "paid",
    purchased_at: new Date(now).toISOString(),
    expires_at: expiresAt,
    subscription_plan: subscriptionPlan || user?.subscription_plan || "basic",
    source: "frontend-demo-payment",
  };

  store[key] = nextEntry;
  writeStore(store);
  return nextEntry;
}
