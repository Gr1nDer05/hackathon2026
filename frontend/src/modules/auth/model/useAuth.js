import { useState } from "react";
import useSWR from "swr";
import useSWRMutation from "swr/mutation";
import { loginRequest, logoutRequest, meRequest } from "../api/authApi";
import {
  activateMockSubscription as persistMockSubscription,
  applyMockSubscription,
} from "../lib/mockSubscription";

function isFutureDate(value) {
  if (!value) return false;

  const timestamp = new Date(value).getTime();
  return Number.isFinite(timestamp) && timestamp > Date.now();
}

function isSubscriptionActive(user) {
  if (!user || user.role !== "psychologist") {
    return true;
  }

  if (user.account_status === "blocked") {
    return false;
  }

  if (user.subscription_status) {
    return user.subscription_status === "active";
  }

  if (user.is_active === false) {
    return false;
  }

  if (isFutureDate(user.blocked_until)) {
    return false;
  }

  if (!user.portal_access_until) {
    return true;
  }

  return isFutureDate(user.portal_access_until);
}

function canActivateSubscriptionStub(user) {
  if (!user || user.role !== "psychologist") {
    return false;
  }

  if (user.account_status === "blocked" || user.is_active === false) {
    return false;
  }

  if (isFutureDate(user.blocked_until)) {
    return false;
  }

  return true;
}

export function useAuth() {
  const session = useSWR("auth-session", meRequest, {
    revalidateOnFocus: false,
    shouldRetryOnError: false,
  });
  const [isActivatingDemoSubscription, setIsActivatingDemoSubscription] = useState(false);
  const [demoSubscriptionError, setDemoSubscriptionError] = useState("");

  const login = useSWRMutation(
    "auth-login",
    async (_, { arg }) => loginRequest(arg),
    {
      onSuccess(data) {
        session.mutate(data, false);
      },
    },
  );

  const logout = useSWRMutation(
    "auth-logout",
    async (_, { arg }) => logoutRequest(arg?.role),
    {
      onSuccess() {
        session.mutate(null, false);
      },
    },
  );

  async function signOut() {
    await logout.trigger({ role: session.data?.role });
  }

  async function activateDemoSubscription() {
    if (!session.data || session.data.role !== "psychologist") {
      return null;
    }

    if (!canActivateSubscriptionStub(session.data)) {
      const error = new Error("Сейчас подписку нельзя активировать из этого экрана.");
      setDemoSubscriptionError(error.message);
      throw error;
    }

    setIsActivatingDemoSubscription(true);
    setDemoSubscriptionError("");

    try {
      const subscription = persistMockSubscription(session.data, {
        days: 30,
        subscriptionPlan: session.data.subscription_plan || "basic",
      });

      const nextUser = applyMockSubscription({
        ...session.data,
        subscription_status: "active",
        portal_access_until: subscription?.expires_at || session.data.portal_access_until,
      });

      session.mutate(nextUser, false);
      return nextUser;
    } catch (error) {
      setDemoSubscriptionError("Не удалось активировать подписку. Попробуйте ещё раз.");
      throw error;
    } finally {
      setIsActivatingDemoSubscription(false);
    }
  }

  const user = applyMockSubscription(session.data);
  const role = user?.role || "";
  const hasActiveSubscription = isSubscriptionActive(user);

  return {
    user,
    role,
    hasActiveSubscription,
    canActivateDemoSubscription: canActivateSubscriptionStub(user),
    isUserLoading: session.isLoading,
    meError: session.error,
    signIn: login.trigger,
    signInError: login.error,
    isSigningIn: login.isMutating,
    signOut,
    isSigningOut: logout.isMutating,
    activateDemoSubscription,
    isActivatingDemoSubscription,
    demoSubscriptionError,
  };
}
