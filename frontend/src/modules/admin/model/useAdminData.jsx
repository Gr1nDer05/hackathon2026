import { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState } from "react";
import useSWR from "swr";
import {
  createPsychologistRequest,
  getPsychologistWorkspaceRequest,
  listPsychologistsRequest,
  updatePsychologistAccessRequest,
} from "../api/adminApi";
import { useAuth } from "../../auth/model/useAuth";

const AdminDataContext = createContext(null);

function isFutureDate(value) {
  if (!value) return false;

  const timestamp = new Date(value).getTime();
  return Number.isFinite(timestamp) && timestamp > Date.now();
}

function formatDate(value) {
  if (!value) return "—";

  const timestamp = new Date(value).getTime();
  if (!Number.isFinite(timestamp)) return "—";

  return new Date(value).toISOString().slice(0, 10);
}

function formatDateTime(value) {
  if (!value) return "—";

  const timestamp = new Date(value).getTime();
  if (!Number.isFinite(timestamp)) return "—";

  return new Date(value).toISOString().slice(0, 16).replace("T", " ");
}

function addDaysToAccessDate(currentValue, days) {
  const base = isFutureDate(currentValue) ? new Date(currentValue) : new Date();
  base.setDate(base.getDate() + days);
  return base.toISOString();
}

function getAccountStatus(user) {
  if (user?.account_status) {
    return user.account_status;
  }

  return user?.is_active === false || isFutureDate(user?.blocked_until) ? "blocked" : "active";
}

function getSubscriptionStatus(user) {
  if (user?.subscription_status) {
    return user.subscription_status;
  }

  if (getAccountStatus(user) === "blocked") {
    return "blocked";
  }

  if (!user?.portal_access_until) {
    return user?.is_active ? "active" : "blocked";
  }

  return isFutureDate(user.portal_access_until) ? "active" : "expired";
}

function getWorkspaceMetrics(workspace) {
  const tests = workspace?.tests || [];
  const testsCount = tests.length;
  const sessionsCount = tests.reduce(
    (total, item) => total + (item.completed_sessions_count || 0),
    0,
  );
  const completionRate = testsCount
    ? Math.round(
        (tests.filter((item) => (item.completed_sessions_count || 0) > 0).length / testsCount) *
          100,
      )
    : 0;

  return {
    testsCount,
    sessionsCount,
    completionRate,
  };
}

function mapPsychologist(user, workspace) {
  const metrics = getWorkspaceMetrics(workspace);
  const subscriptionPlan = user?.subscription_plan || "basic";

  return {
    id: String(user.id),
    backendId: user.id,
    name: user.full_name,
    email: user.email,
    phone: workspace?.card?.contact_phone || "Не указан",
    city: workspace?.profile?.city || "Не указан",
    specialization: workspace?.profile?.specialization || "Не указана",
    joinedAt: formatDate(user.created_at),
    lastActiveAt: formatDateTime(workspace?.profile?.updated_at || user.updated_at),
    accountStatus: getAccountStatus(user),
    subscriptionStatus: getSubscriptionStatus(user),
    subscriptionPlan,
    expiresAt: formatDate(user.portal_access_until),
    hasPassword: true,
    passwordUpdatedAt: formatDateTime(user.updated_at),
    raw: user,
    workspace,
    ...metrics,
  };
}

function mapSubscription(user) {
  const subscriptionPlan = user?.subscription_plan || "basic";

  return {
    id: `subscription-${user.id}`,
    psychologistId: String(user.id),
    psychologistName: user.full_name,
    plan: subscriptionPlan === "pro" ? "Pro" : "Basic",
    status: getSubscriptionStatus(user),
    startedAt: formatDate(user.created_at),
    expiresAt: formatDate(user.portal_access_until),
  };
}

export function AdminDataProvider({ children }) {
  const { user, role } = useAuth();
  const [workspaceById, setWorkspaceById] = useState({});
  const [workspaceLoadingById, setWorkspaceLoadingById] = useState({});
  const [workspaceErrorById, setWorkspaceErrorById] = useState({});
  const workspaceByIdRef = useRef({});

  const shouldFetchAdminData = role === "admin" && Boolean(user);
  const psychologistsQuery = useSWR(
    shouldFetchAdminData ? "admin-psychologists" : null,
    listPsychologistsRequest,
  );

  const rawPsychologists = psychologistsQuery.data || [];

  useEffect(() => {
    workspaceByIdRef.current = workspaceById;
  }, [workspaceById]);

  const ensureWorkspace = useCallback(
    async (id, { force = false } = {}) => {
      const key = String(id);

      if (!force && workspaceByIdRef.current[key]) {
        return workspaceByIdRef.current[key];
      }

      setWorkspaceLoadingById((prev) => ({ ...prev, [key]: true }));
      setWorkspaceErrorById((prev) => {
        const next = { ...prev };
        delete next[key];
        return next;
      });

      try {
        const workspace = await getPsychologistWorkspaceRequest(id);
        setWorkspaceById((prev) => ({ ...prev, [key]: workspace }));
        return workspace;
      } catch (error) {
        setWorkspaceErrorById((prev) => ({ ...prev, [key]: error }));
        throw error;
      } finally {
        setWorkspaceLoadingById((prev) => ({ ...prev, [key]: false }));
      }
    },
    [],
  );

  useEffect(() => {
    if (!shouldFetchAdminData || rawPsychologists.length === 0) {
      return;
    }

    rawPsychologists.forEach((item) => {
      const key = String(item.id);

      if (!workspaceById[key] && !workspaceLoadingById[key]) {
        ensureWorkspace(item.id).catch(() => {});
      }
    });
  }, [
    ensureWorkspace,
    rawPsychologists,
    shouldFetchAdminData,
    workspaceById,
    workspaceLoadingById,
  ]);

  const psychologists = useMemo(
    () =>
      rawPsychologists.map((item) =>
        mapPsychologist(item, workspaceById[String(item.id)]),
      ),
    [rawPsychologists, workspaceById],
  );

  const subscriptions = useMemo(
    () => rawPsychologists.map(mapSubscription),
    [rawPsychologists],
  );

  async function refreshPsychologists() {
    await psychologistsQuery.mutate();
  }

  async function addPsychologist(payload) {
    try {
      const workspace = await createPsychologistRequest(payload);

      if (workspace?.user?.id) {
        setWorkspaceById((prev) => ({
          ...prev,
          [String(workspace.user.id)]: workspace,
        }));
      }

      await refreshPsychologists();
      return workspace;
    } catch (error) {
      await refreshPsychologists();
      throw error;
    }
  }

  async function togglePsychologistStatus(id) {
    const current = rawPsychologists.find((item) => String(item.id) === String(id));

    if (!current) {
      return;
    }

    const isBlocked = getAccountStatus(current) === "blocked";

    await updatePsychologistAccessRequest(current.id, {
      is_active: isBlocked,
      blocked_until: isBlocked ? "" : addDaysToAccessDate(null, 30),
    });

    await refreshPsychologists();
  }

  async function extendPsychologistSubscription(id, days) {
    const current = rawPsychologists.find((item) => String(item.id) === String(id));

    if (!current) {
      return;
    }

    await updatePsychologistAccessRequest(current.id, {
      is_active: true,
      blocked_until: "",
      subscription_days: days,
    });

    await refreshPsychologists();
  }

  async function setPsychologistSubscriptionPlan(id, plan) {
    const current = rawPsychologists.find((item) => String(item.id) === String(id));

    if (!current) {
      return;
    }

    await updatePsychologistAccessRequest(current.id, {
      subscription_plan: plan,
    });

    await refreshPsychologists();
  }

  async function extendSubscription(subscriptionId, days) {
    const subscription = subscriptions.find((item) => item.id === subscriptionId);

    if (!subscription) {
      return;
    }

    await extendPsychologistSubscription(subscription.psychologistId, days);
  }

  return (
    <AdminDataContext.Provider
      value={{
        psychologists,
        subscriptions,
        addPsychologist,
        togglePsychologistStatus,
        extendPsychologistSubscription,
        setPsychologistSubscriptionPlan,
        extendSubscription,
        ensureWorkspace,
        workspaceById,
        workspaceLoadingById,
        workspaceErrorById,
        isLoading: shouldFetchAdminData && psychologistsQuery.isLoading,
        error: psychologistsQuery.error,
        refreshPsychologists,
      }}
    >
      {children}
    </AdminDataContext.Provider>
  );
}

export function useAdminData() {
  const value = useContext(AdminDataContext);

  if (!value) {
    throw new Error("useAdminData must be used within AdminDataProvider");
  }

  return value;
}
