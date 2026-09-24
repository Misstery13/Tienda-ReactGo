import { Link, useLocation } from "react-router-dom";
import { House, LayoutGrid, Store, Users, type LucideIcon } from "lucide-react";
import { useAuth } from "../context/AuthContext";

interface SidebarProps {
  isCollapsed: boolean;
  onToggle: () => void;
}

// Opciones de navegación disponibles según el rol (Tema 5)
interface NavItem {
  to: string;
  label: string;
  icon: LucideIcon;
  soloAdmin?: boolean;
}

const NAV_ITEMS: NavItem[] = [
  { to: "/", label: "Dashboard", icon: House, soloAdmin: true },
  { to: "/tienda", label: "Tienda", icon: Store },
  { to: "/catalogo", label: "Catálogo", icon: LayoutGrid },
  { to: "/mi-red", label: "Mi Red", icon: Users, soloAdmin: true },
];

const Sidebar = ({ isCollapsed, onToggle }: SidebarProps) => {
  const { user } = useAuth();
  const { pathname } = useLocation();

  const handleOptionClick = () => {
    if (!isCollapsed) {
      onToggle();
    }
  };

  // Las opciones de admin se ocultan para el rol cliente
  const items = NAV_ITEMS.filter((item) => !item.soloAdmin || user?.rol === "admin");

  return (
    <aside className={`${isCollapsed ? "fixed inset-y-0 left-0 z-50 w-full -translate-x-full opacity-0 pointer-events-none sm:static sm:inset-auto sm:flex sm:w-20 sm:translate-x-0 sm:opacity-100 sm:pointer-events-auto" : "fixed inset-0 z-50 flex w-full translate-x-0 opacity-100 pointer-events-auto sm:static sm:inset-auto sm:w-64"} bg-teal-500 text-white flex-col shrink-0 overflow-y-auto will-change-transform transition-[transform,opacity,width] duration-500 ease-in-out`}>
      <div className="p-4 border-b border-slate-700 text-xl font-bold">
        {!isCollapsed && <span>MultiCatálogo</span>}
      </div>
      <nav className="flex-1 p-2 sm:p-4 space-y-2">
        {items.map((item) => {
          const Icono = item.icon;
          // La opción activa se resalta según la ruta actual
          const esActivo =
            item.to === "/" ? pathname === "/" : pathname.startsWith(item.to);

          return (
            <Link
              key={item.to}
              to={item.to}
              onClick={handleOptionClick}
              title={item.label}
              aria-label={item.label}
              className={`${isCollapsed ? "text-center" : "text-left"} block p-3 rounded transition ${
                esActivo ? "bg-teal-700" : "hover:bg-slate-800"
              }`}
            >
              <Icono size={20} aria-hidden="true" className="inline-block" />
              {!isCollapsed && <span className="ml-3">{item.label}</span>}
            </Link>
          );
        })}
      </nav>
      <div className="p-4 border-t border-teal-600 text-xs text-teal-50">
        {isCollapsed ? (
          <p className="text-center uppercase">{user?.rol}</p>
        ) : (
          <p>
            Conectado como <span className="font-semibold uppercase">{user?.rol}</span>
          </p>
        )}
      </div>
    </aside>
  );
};

export default Sidebar;
