import { Suspense, lazy } from "react";
import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import RequireActiveSubscription from "./guards/RequireActiveSubscription";
import RequireAuth from "./guards/RequireAuth";
import RequireRole from "./guards/RequireRole";
import { useAuth } from "../modules/auth/model/useAuth";
import { resolveHomeRoute } from "../modules/auth/model/access";
import { ROUTES } from "../shared/config/routes";

const AdminDashboardPage = lazy(() => import("../pages/admin/AdminDashboardPage"));
const AdminLayout = lazy(() => import("../pages/admin/AdminLayout"));
const PsychologistDetailsPage = lazy(() => import("../pages/admin/PsychologistDetailsPage"));
const PsychologistsPage = lazy(() => import("../pages/admin/PsychologistsPage"));
const SubscriptionsPage = lazy(() => import("../pages/admin/SubscriptionsPage"));
const AuthPage = lazy(() => import("../pages/auth/AuthPage"));
const ClientAuthorPage = lazy(() => import("../pages/client/ClientAuthorPage"));
const ClientResultPage = lazy(() => import("../pages/client/ClientResultPage"));
const ClientSessionPage = lazy(() => import("../pages/client/ClientSessionPage"));
const ForbiddenPage = lazy(() => import("../pages/common/ForbiddenPage"));
const NotFoundPage = lazy(() => import("../pages/common/NotFoundPage"));
const SubscriptionRequiredPage = lazy(() => import("../pages/common/SubscriptionRequiredPage"));
const BuilderPage = lazy(() => import("../pages/psychologist/BuilderPage"));
const DashboardPage = lazy(() => import("../pages/psychologist/DashboardPage"));
const ProfilePage = lazy(() => import("../pages/psychologist/ProfilePage"));
const ReportTemplatesPage = lazy(() => import("../pages/psychologist/ReportTemplatesPage"));
const TestSubmissionPage = lazy(() => import("../pages/psychologist/TestSubmissionPage"));
const TestResultsPage = lazy(() => import("../pages/psychologist/TestResultsPage"));
const TestsPage = lazy(() => import("../pages/psychologist/TestsPage"));

function RoleHomeRedirect() {
  const { user, role, hasActiveSubscription, isUserLoading } = useAuth();

  if (isUserLoading) {
    return <main className="screen">Загружаем сессию...</main>;
  }

  if (!user) {
    return <AuthPage />;
  }

  const targetRoute = resolveHomeRoute({
    role,
    hasActiveSubscription,
  });
  return <Navigate to={targetRoute} replace />;
}

function PsychologistRoute({ children }) {
  return (
    <RequireAuth>
      <RequireRole roles={["psychologist"]}>
        <RequireActiveSubscription>{children}</RequireActiveSubscription>
      </RequireRole>
    </RequireAuth>
  );
}

function AdminRoute({ children }) {
  return (
    <RequireAuth>
      <RequireRole roles={["admin"]}>{children}</RequireRole>
    </RequireAuth>
  );
}

function RouteLoader() {
  return <main className="screen">Загрузка...</main>;
}

export default function AppRouter() {
  return (
    <BrowserRouter>
      <Suspense fallback={<RouteLoader />}>
        <Routes>
          <Route path={ROUTES.root} element={<RoleHomeRedirect />} />

          <Route path={ROUTES.clientSession} element={<ClientSessionPage />} />
          <Route path={ROUTES.clientAuthor} element={<ClientAuthorPage />} />
          <Route path={ROUTES.clientResult} element={<ClientResultPage />} />

          <Route
            path={ROUTES.dashboard}
            element={
              <PsychologistRoute>
                <DashboardPage />
              </PsychologistRoute>
            }
          />
          <Route
            path={ROUTES.tests}
            element={
              <PsychologistRoute>
                <TestsPage />
              </PsychologistRoute>
            }
          />
          <Route
            path={ROUTES.testBuilder}
            element={
              <PsychologistRoute>
                <BuilderPage />
              </PsychologistRoute>
            }
          />
          <Route
            path={ROUTES.testResults}
            element={
              <PsychologistRoute>
                <TestResultsPage />
              </PsychologistRoute>
            }
          />
          <Route
            path={ROUTES.testSubmission}
            element={
              <PsychologistRoute>
                <TestSubmissionPage />
              </PsychologistRoute>
            }
          />
          <Route
            path={ROUTES.reportTemplates}
            element={
              <PsychologistRoute>
                <ReportTemplatesPage />
              </PsychologistRoute>
            }
          />
          <Route
            path={ROUTES.profile}
            element={
              <PsychologistRoute>
                <ProfilePage />
              </PsychologistRoute>
            }
          />

          <Route
            path="/admin"
            element={
              <AdminRoute>
                <AdminLayout />
              </AdminRoute>
            }
          >
            <Route index element={<AdminDashboardPage />} />
            <Route path="psychologists" element={<PsychologistsPage />} />
            <Route path="psychologists/:id" element={<PsychologistDetailsPage />} />
            <Route path="subscriptions" element={<SubscriptionsPage />} />
          </Route>

          <Route
            path={ROUTES.subscriptionRequired}
            element={
              <RequireAuth>
                <RequireRole roles={["psychologist"]}>
                  <SubscriptionRequiredPage />
                </RequireRole>
              </RequireAuth>
            }
          />
          <Route
            path={ROUTES.forbidden}
            element={
              <RequireAuth>
                <ForbiddenPage />
              </RequireAuth>
            }
          />

          <Route path={ROUTES.notFound} element={<NotFoundPage />} />
        </Routes>
      </Suspense>
    </BrowserRouter>
  );
}
