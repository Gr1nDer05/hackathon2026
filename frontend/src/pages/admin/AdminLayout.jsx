import { CreditCard, LayoutDashboard, LogOut, Menu, Users, X } from "lucide-react";
import { AnimatePresence, motion, useReducedMotion } from "motion/react";
import { useEffect, useState } from "react";
import { NavLink, Outlet, useLocation } from "react-router-dom";
import { useAuth } from "../../modules/auth/model/useAuth";
import { ROUTES } from "../../shared/config/routes";
import {
  MOTION_EASE,
  createFadeMove,
  createRevealContainer,
} from "../../shared/lib/motion";

const NAV_ITEMS = [
  {
    to: ROUTES.adminDashboard,
    label: "Обзор",
    end: true,
    icon: LayoutDashboard,
  },
  { to: ROUTES.adminPsychologists, label: "Психологи", icon: Users },
  { to: ROUTES.adminSubscriptions, label: "Подписки", icon: CreditCard },
];

export default function AdminLayout() {
  const { signOut } = useAuth();
  const location = useLocation();
  const reducedMotion = useReducedMotion();
  const [isNavOpen, setIsNavOpen] = useState(false);
  const shellVariants = createRevealContainer(reducedMotion, {
    staggerChildren: 0.08,
    delayChildren: 0.03,
  });
  const sidebarVariants = createFadeMove(reducedMotion, {
    axis: "x",
    distance: -20,
    scale: 0.996,
  });
  const contentVariants = createFadeMove(reducedMotion, {
    axis: "y",
    distance: 18,
    scale: 0.994,
  });
  const navVariants = createRevealContainer(reducedMotion, {
    staggerChildren: 0.05,
    delayChildren: 0.04,
  });
  const navItemVariants = createFadeMove(reducedMotion, {
    axis: "y",
    distance: 12,
    scale: 0.998,
  });

  useEffect(() => {
    setIsNavOpen(false);
  }, [location.pathname]);

  return (
    <motion.main
      animate="visible"
      className="admin-shell"
      initial="hidden"
      variants={shellVariants}
    >
      <motion.aside className="admin-shell__sidebar" variants={sidebarVariants}>
        <motion.div className="admin-shell__brand" variants={navItemVariants}>
          <div>
            <p className="admin-shell__badge">TitanIT Admin</p>
            <h1 className="admin-shell__title">ПрофДНК</h1>
          </div>

          <button
            aria-expanded={isNavOpen}
            aria-label={isNavOpen ? "Свернуть навигацию" : "Развернуть навигацию"}
            className="admin-shell__menu-toggle"
            type="button"
            onClick={() => setIsNavOpen((prev) => !prev)}
          >
            {isNavOpen ? <X size={18} strokeWidth={2.1} /> : <Menu size={18} strokeWidth={2.1} />}
          </button>
        </motion.div>

        <motion.div
          className={
            isNavOpen
              ? "admin-shell__sidebar-body admin-shell__sidebar-body--open"
              : "admin-shell__sidebar-body"
          }
          variants={navVariants}
        >
          <motion.nav
            className="admin-shell__nav"
            aria-label="Admin sections"
            variants={navVariants}
          >
            {NAV_ITEMS.map((item) => (
              <motion.div key={item.to} variants={navItemVariants}>
                <NavLink
                  to={item.to}
                  end={item.end}
                  className={({ isActive }) =>
                    isActive
                      ? "admin-shell__link admin-shell__link--active"
                      : "admin-shell__link"
                  }
                >
                  <item.icon size={16} strokeWidth={2.1} />
                  <span>{item.label}</span>
                </NavLink>
              </motion.div>
            ))}
          </motion.nav>

          <motion.button
            className="admin-shell__logout"
            type="button"
            variants={navItemVariants}
            whileHover={reducedMotion ? undefined : { y: -1 }}
            whileTap={reducedMotion ? undefined : { scale: 0.992 }}
            onClick={signOut}
          >
            <LogOut size={16} strokeWidth={2.1} />
            <span>Выйти</span>
          </motion.button>
        </motion.div>
      </motion.aside>

      <motion.section className="admin-shell__content" variants={contentVariants}>
        <motion.header className="admin-shell__header" variants={navItemVariants}>
          <p className="admin-shell__header-title">
            Администрирование платформы ПрофДНК
          </p>
          <p className="admin-shell__header-meta">
            Управление специалистами, подписками и доступом.
          </p>
        </motion.header>
        <AnimatePresence mode="wait" initial={false}>
          <motion.div
            key={location.pathname}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: reducedMotion ? 0 : -10 }}
            initial={{ opacity: 0, y: reducedMotion ? 0 : 10 }}
            transition={{
              duration: reducedMotion ? 0.01 : 0.26,
              ease: MOTION_EASE,
            }}
          >
            <Outlet />
          </motion.div>
        </AnimatePresence>
      </motion.section>
    </motion.main>
  );
}
