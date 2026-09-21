import { Link } from "react-router-dom";
import { House, LayoutGrid, Users } from "lucide-react";
interface SidebarProps {
 isCollapsed: boolean;
 onToggle: () => void;
}

const Sidebar = ({ isCollapsed, onToggle }: SidebarProps) => {
const handleOptionClick = () => {
 if (!isCollapsed) {
 onToggle();
 }
};

return (
		<aside className={`${isCollapsed ? "fixed inset-y-0 left-0 z-50 w-full -translate-x-full opacity-0 pointer-events-none sm:static sm:inset-auto sm:flex sm:w-20 sm:translate-x-0 sm:opacity-100 sm:pointer-events-auto" : "fixed inset-0 z-50 flex w-full translate-x-0 opacity-100 pointer-events-auto sm:static sm:inset-auto sm:w-64"} bg-teal-500 text-white flex-col shrink-0 overflow-y-auto will-change-transform transition-[transform,opacity,width] duration-500 ease-in-out`}>
			<div className="p-4 border-b border-slate-700 text-xl font-bold">
				{!isCollapsed && <span>MultiCatálogo</span>}
 			</div>
 <nav className="flex-1 p-2 sm:p-4 space-y-2">
 <Link
 to="/"
 onClick={handleOptionClick}
 title="Dashboard"
 aria-label="Dashboard"
 className={`${isCollapsed ? "text-center" : "text-left"} block p-3 rounded hover:bg-slate-800 transition`}
 >
 <House size={20} aria-hidden="true" className="inline-block" />
 {!isCollapsed && <span className="ml-3">Dashboard</span>}
 </Link>
 <Link
 to="/catalogo"
 onClick={handleOptionClick}
 title="Catálogo"
 aria-label="Catálogo"
 className={`${isCollapsed ? "text-center" : "text-left"} block p-3 rounded hover:bg-slate-800 transition`}
 >
 <LayoutGrid size={20} aria-hidden="true" className="inline-block" />
 {!isCollapsed && <span className="ml-3">Catálogo</span>}
 </Link>
 <Link
 to="/mi-red"
 onClick={handleOptionClick}
 title="Mi Red"
 aria-label="Mi Red"
 className={`${isCollapsed ? "text-center" : "text-left"} block p-3 rounded hover:bg-slate-800 transition`}
 >
 <Users size={20} aria-hidden="true" className="inline-block" />
 {!isCollapsed && <span className="ml-3">Mi Red</span>}
 </Link>
 </nav>
 </aside>
 );
};
export default Sidebar;