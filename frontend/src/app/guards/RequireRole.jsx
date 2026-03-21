import { Navigate } from "react-router-dom";
import { useAuth } from "../../modules/auth/model/useAuth";
import { ROUTES } from "../../shared/config/routes";

export default function RequireRole({ roles, children }) {
  const { role } = useAuth();

  if (!roles.includes(role)) {
    return <Navigate to={ROUTES.forbidden} replace />;
  }

  return children;
}
