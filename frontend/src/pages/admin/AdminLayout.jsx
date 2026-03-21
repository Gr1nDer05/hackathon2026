import { CreditCard, LayoutDashboard, LogOut, Menu, Users, X } from "lucide-react";
import { useEffect, useState } from "react";
import { NavLink, Outlet, useLocation } from "react-router-dom";
import { useAuth } from "../../modules/auth/model/useAuth";
import { ROUTES } from "../../shared/config/routes";

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
  const [isNavOpen, setIsNavOpen] = useState(false);

  useEffect(() => {
    setIsNavOpen(false);
  }, [location.pathname]);

  return (
    <main className="admin-shell">
      <aside className="admin-shell__sidebar">
        <div className="admin-shell__brand">
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
        </div>

        <div
          className={
            isNavOpen
              ? "admin-shell__sidebar-body admin-shell__sidebar-body--open"
              : "admin-shell__sidebar-body"
          }
        >
          <nav className="admin-shell__nav" aria-label="Admin sections">
            {NAV_ITEMS.map((item) => (
              <NavLink
                key={item.to}
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
            ))}
          </nav>

          <button className="admin-shell__logout" type="button" onClick={signOut}>
            <LogOut size={16} strokeWidth={2.1} />
            <span>Выйти</span>
          </button>
        </div>
      </aside>

      <section className="admin-shell__content">
        <header className="admin-shell__header">
          <p className="admin-shell__header-title">
            Администрирование платформы ПрофДНК
          </p>
          <p className="admin-shell__header-meta">
            Рабочий контур админа для управления специалистами и их доступом.
          </p>
        </header>
        <Outlet />
      </section>
    </main>
  );
}
