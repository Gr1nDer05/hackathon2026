import { Navigate } from "react-router-dom";
import { useAuth } from "../../modules/auth/model/useAuth";
import { ROUTES } from "../../shared/config/routes";

export default function RequireAuth({ children }) {
  const { user, isUserLoading } = useAuth();

  if (isUserLoading) {
    return <main className="screen">Загружаем сессию...</main>;
  }

  if (!user) {
    return <Navigate to={ROUTES.root} replace />;
  }

  return children;
}
