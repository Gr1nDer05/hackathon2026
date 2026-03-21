import { Navigate } from "react-router-dom";
import LoginForm from "../../modules/auth/ui/LoginForm";
import { resolveHomeRoute } from "../../modules/auth/model/access";
import { useAuth } from "../../modules/auth/model/useAuth";
import { ROUTES } from "../../shared/config/routes";

export default function AuthPage() {
  const {
    user,
    role,
    hasActiveSubscription,
    isUserLoading,
    signIn,
    isSigningIn,
    signInError,
    meError,
  } = useAuth();

  async function handleSignIn(payload) {
    await signIn(payload);
  }

  if (isUserLoading) {
    return <main className="screen auth-screen">Загружаем сессию...</main>;
  }

  if (user) {
    const nextRoute = resolveHomeRoute({
      role,
      hasActiveSubscription,
    });
    return <Navigate to={nextRoute || ROUTES.root} replace />;
  }

  return (
    <main className="screen auth-screen">
      <LoginForm
        onSignIn={handleSignIn}
        isSigningIn={isSigningIn}
        signInError={signInError || meError}
      />
    </main>
  );
}
