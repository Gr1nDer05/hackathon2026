import { ROUTES } from "../../../shared/config/routes";

export function resolveHomeRoute({ role, hasActiveSubscription }) {
  if (role === "admin") {
    return ROUTES.adminDashboard;
  }

  if (role === "psychologist") {
    return hasActiveSubscription ? ROUTES.dashboard : ROUTES.subscriptionRequired;
  }

  return ROUTES.root;
}
