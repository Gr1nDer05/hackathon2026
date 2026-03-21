import useSWR from "swr";
import useSWRMutation from "swr/mutation";
import { loginRequest, logoutRequest, meRequest } from "../api/authApi";

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

export function useAuth() {
  const session = useSWR("auth-session", meRequest, {
    revalidateOnFocus: false,
    shouldRetryOnError: false,
  });

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

  const role = session.data?.role || "";
  const hasActiveSubscription = isSubscriptionActive(session.data);

  return {
    user: session.data,
    role,
    hasActiveSubscription,
    isUserLoading: session.isLoading,
    meError: session.error,
    signIn: login.trigger,
    signInError: login.error,
    isSigningIn: login.isMutating,
    signOut,
    isSigningOut: logout.isMutating,
  };
}
