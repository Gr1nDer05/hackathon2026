import { Suspense, lazy } from "react";
import { motion, useReducedMotion } from "motion/react";
import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import RequireActiveSubscription from "./guards/RequireActiveSubscription";
import RequireAuth from "./guards/RequireAuth";
import RequireRole from "./guards/RequireRole";
import { useAuth } from "../modules/auth/model/useAuth";
import { resolveHomeRoute } from "../modules/auth/model/access";
import { ROUTES } from "../shared/config/routes";
import { MOTION_EASE } from "../shared/lib/motion";

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
  return (
    <motion.main
      animate={{ opacity: 1, y: 0 }}
      className="screen"
      initial={{ opacity: 0, y: 12 }}
      transition={{ duration: 0.24, ease: MOTION_EASE }}
    >
      Загрузка...
    </motion.main>
  );
}

function RouteScene({ children }) {
  const reducedMotion = useReducedMotion();

  return (
    <motion.div
      animate={{ opacity: 1, y: 0, scale: 1 }}
      initial={{
        opacity: 0,
        y: reducedMotion ? 0 : 16,
        scale: reducedMotion ? 1 : 0.996,
      }}
      transition={{
        duration: reducedMotion ? 0.01 : 0.34,
        ease: MOTION_EASE,
      }}
    >
      {children}
    </motion.div>
  );
}

export default function AppRouter() {
  return (
    <BrowserRouter>
      <Suspense fallback={<RouteLoader />}>
        <Routes>
          <Route path={ROUTES.root} element={<RoleHomeRedirect />} />

          <Route path={ROUTES.clientSession} element={<RouteScene><ClientSessionPage /></RouteScene>} />
          <Route path={ROUTES.clientAuthor} element={<RouteScene><ClientAuthorPage /></RouteScene>} />
          <Route path={ROUTES.clientResult} element={<RouteScene><ClientResultPage /></RouteScene>} />

          <Route
            path={ROUTES.dashboard}
            element={
              <PsychologistRoute>
                <RouteScene>
                  <DashboardPage />
                </RouteScene>
              </PsychologistRoute>
            }
          />
          <Route
            path={ROUTES.tests}
            element={
              <PsychologistRoute>
                <RouteScene>
                  <TestsPage />
                </RouteScene>
              </PsychologistRoute>
            }
          />
          <Route
            path={ROUTES.testBuilder}
            element={
              <PsychologistRoute>
                <RouteScene>
                  <BuilderPage />
                </RouteScene>
              </PsychologistRoute>
            }
          />
          <Route
            path={ROUTES.testResults}
            element={
              <PsychologistRoute>
                <RouteScene>
                  <TestResultsPage />
                </RouteScene>
              </PsychologistRoute>
            }
          />
          <Route
            path={ROUTES.testSubmission}
            element={
              <PsychologistRoute>
                <RouteScene>
                  <TestSubmissionPage />
                </RouteScene>
              </PsychologistRoute>
            }
          />
          <Route
            path={ROUTES.reportTemplates}
            element={
              <PsychologistRoute>
                <RouteScene>
                  <ReportTemplatesPage />
                </RouteScene>
              </PsychologistRoute>
            }
          />
          <Route
            path={ROUTES.profile}
            element={
              <PsychologistRoute>
                <RouteScene>
                  <ProfilePage />
                </RouteScene>
              </PsychologistRoute>
            }
          />

          <Route
            path="/admin"
            element={
              <AdminRoute>
                <RouteScene>
                  <AdminLayout />
                </RouteScene>
              </AdminRoute>
            }
          >
            <Route index element={<RouteScene><AdminDashboardPage /></RouteScene>} />
            <Route path="psychologists" element={<RouteScene><PsychologistsPage /></RouteScene>} />
            <Route path="psychologists/:id" element={<RouteScene><PsychologistDetailsPage /></RouteScene>} />
            <Route path="subscriptions" element={<RouteScene><SubscriptionsPage /></RouteScene>} />
          </Route>

          <Route
            path={ROUTES.subscriptionRequired}
            element={
              <RequireAuth>
                <RequireRole roles={["psychologist"]}>
                  <RouteScene>
                    <SubscriptionRequiredPage />
                  </RouteScene>
                </RequireRole>
              </RequireAuth>
            }
          />
          <Route
            path={ROUTES.forbidden}
            element={
              <RequireAuth>
                <RouteScene>
                  <ForbiddenPage />
                </RouteScene>
              </RequireAuth>
            }
          />

          <Route path={ROUTES.notFound} element={<RouteScene><NotFoundPage /></RouteScene>} />
        </Routes>
      </Suspense>
    </BrowserRouter>
  );
}
