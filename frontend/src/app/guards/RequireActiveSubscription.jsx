import { Navigate } from "react-router-dom";
import { useAuth } from "../../modules/auth/model/useAuth";
import { ROUTES } from "../../shared/config/routes";

export default function RequireActiveSubscription({ children }) {
  const { role, hasActiveSubscription } = useAuth();

  if (role === "psychologist" && !hasActiveSubscription) {
    return <Navigate to={ROUTES.subscriptionRequired} replace />;
  }

  return children;
}
