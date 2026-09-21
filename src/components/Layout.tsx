import { Outlet } from "react-router-dom";
import Sidebar from "./Sidebar";
import Navbar from "./Navbar";
interface LayoutProps {
 isCollapsed: boolean;
 onToggle: () => void;
}

const Layout = ({ isCollapsed, onToggle }: LayoutProps) => {
return (
 <div className="flex h-screen bg-slate-50">
			<Sidebar isCollapsed={isCollapsed} onToggle={onToggle} />
 {/* Área de Contenido Principal */}
 <div className="flex-1 flex flex-col overflow-hidden">
 <Navbar isCollapsed={isCollapsed} onToggle={onToggle} />
 <main className="flex-1 overflow-x-hidden overflow-y-auto p-8"> 
<Outlet />
 </main>
 </div>
 </div>
 );
};
export default Layout;