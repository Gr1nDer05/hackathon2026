import { requestJson } from "../../../shared/api/http";

export async function listPsychologistsRequest() {
  const data = await requestJson("/admins/psychologists");
  return Array.isArray(data) ? data : [];
}

export async function getPsychologistWorkspaceRequest(id) {
  return requestJson(`/admins/psychologists/${id}/workspace`);
}

export async function updatePsychologistAccessRequest(id, payload) {
  return requestJson(`/admins/psychologists/${id}/access`, {
    method: "PUT",
    body: payload,
  });
}

export async function updatePsychologistProfileRequest(id, payload) {
  return requestJson(`/admins/psychologists/${id}/profile`, {
    method: "PUT",
    body: payload,
  });
}

export async function updatePsychologistCardRequest(id, payload) {
  return requestJson(`/admins/psychologists/${id}/card`, {
    method: "PUT",
    body: payload,
  });
}

export async function createPsychologistRequest(payload) {
  const workspace = await requestJson("/admins/psychologists", {
    method: "POST",
    body: {
      email: payload.email,
      password: payload.password,
      full_name: payload.name,
      is_active: true,
    },
  });

  const updates = [];
  const psychologistId = workspace?.user?.id;

  if (!psychologistId) {
    return workspace;
  }

  if (payload.specialization || payload.city) {
    updates.push(
      updatePsychologistProfileRequest(psychologistId, {
        specialization: payload.specialization || "",
        city: payload.city || "",
      }),
    );
  }

  if (payload.phone) {
    updates.push(
      updatePsychologistCardRequest(psychologistId, {
        contact_phone: payload.phone,
      }),
    );
  }

  if (payload.subscriptionDays) {
    updates.push(
      updatePsychologistAccessRequest(psychologistId, {
        is_active: true,
        blocked_until: "",
        subscription_plan: payload.subscriptionPlan || "basic",
        subscription_days: payload.subscriptionDays,
      }),
    );
  } else if (payload.subscriptionPlan) {
    updates.push(
      updatePsychologistAccessRequest(psychologistId, {
        subscription_plan: payload.subscriptionPlan,
      }),
    );
  }

  const results = await Promise.allSettled(updates);
  const hasRejectedUpdates = results.some((result) => result.status === "rejected");

  if (hasRejectedUpdates) {
    const error = new Error("Аккаунт создан, но часть дополнительных полей не сохранилась.");
    error.partial = true;
    throw error;
  }

  return getPsychologistWorkspaceRequest(psychologistId);
}
