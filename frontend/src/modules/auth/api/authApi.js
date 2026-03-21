import { requestJson } from "../../../shared/api/http";

function isEmailIdentifier(value) {
  return value.includes("@");
}

function normalizeAdminUser(user) {
  return {
    ...user,
    role: "admin",
  };
}

function normalizePsychologistWorkspace(workspace) {
  return {
    ...workspace.user,
    role: "psychologist",
    workspace,
    profile: workspace.profile,
    card: workspace.card,
    tests: workspace.tests,
  };
}

function isAuthError(error) {
  return [401, 403].includes(error?.status);
}

export async function loginRequest(payload) {
  const identifier = String(payload.identifier ?? payload.email ?? payload.login ?? "").trim();
  const password = String(payload.password ?? "");

  if (isEmailIdentifier(identifier)) {
    const data = await requestJson("/auth/psychologists/login", {
      method: "POST",
      body: {
        email: identifier,
        password,
      },
    });

    return normalizePsychologistWorkspace(data.workspace);
  }

  const data = await requestJson("/auth/admin/login", {
    method: "POST",
    body: {
      login: identifier,
      password,
    },
  });

  return normalizeAdminUser(data.user);
}

export async function meRequest() {
  try {
    const admin = await requestJson("/admins/me");
    return normalizeAdminUser(admin);
  } catch (error) {
    if (!isAuthError(error)) {
      throw error;
    }
  }

  try {
    const workspace = await requestJson("/psychologists/me");
    return normalizePsychologistWorkspace(workspace);
  } catch (error) {
    if (isAuthError(error)) {
      return null;
    }

    throw error;
  }
}

export async function logoutRequest(role) {
  const path = role === "admin" ? "/auth/admin/logout" : "/auth/psychologists/logout";

  try {
    return await requestJson(path, {
      method: "POST",
    });
  } catch (error) {
    if (isAuthError(error)) {
      return { status: "ok" };
    }

    throw error;
  }
}

export async function updateAdminMeRequest(payload) {
  return requestJson("/admins/me", {
    method: "PUT",
    body: {
      email: String(payload?.email ?? "").trim(),
    },
  });
}

export async function resendAdminEmailVerificationRequest() {
  return requestJson("/admins/me/email/verification-code", {
    method: "POST",
  });
}

export async function confirmAdminEmailRequest(payload) {
  return requestJson("/admins/me/email/confirm", {
    method: "POST",
    body: {
      code: String(payload?.code ?? "").trim(),
    },
  });
}
