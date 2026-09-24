import { useState } from "react";
import { Outlet, useLocation } from "react-router-dom";
import Sidebar from "./Sidebar";
import Navbar from "./Navbar";

const Layout = () => {
  const [isCollapsed, setIsCollapsed] = useState(() => window.innerWidth < 640);
  const { pathname } = useLocation();

  // El storefront ocupa todo el viewport, por eso se muestra sin el padding del área de contenido
  const esPantallaCompleta = pathname === "/tienda";

  return (
    <div className="flex h-screen bg-slate-50">
      <Sidebar
        isCollapsed={isCollapsed}
        onToggle={() => setIsCollapsed((collapsed) => !collapsed)}
      />
      {/* Área de Contenido Principal */}
      <div className="flex-1 flex flex-col overflow-hidden">
        <Navbar
          isCollapsed={isCollapsed}
          onToggle={() => setIsCollapsed((collapsed) => !collapsed)}
        />
        <main className={`flex-1 overflow-x-hidden overflow-y-auto ${esPantallaCompleta ? "" : "p-8"}`}>
          <Outlet />
        </main>
      </div>
    </div>
  );
};

export default Layout;
