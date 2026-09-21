import { Link, useNavigate } from "react-router-dom";
import { useCart } from "../context/CartContext";
import { useAuth } from "../context/AuthContext";
import { Menu, ShoppingCart, X } from "lucide-react";
interface NavbarProps {
 isCollapsed: boolean;
 onToggle: () => void;
}

const Navbar = ({ isCollapsed, onToggle }: NavbarProps) => {
const { totalItems } = useCart();
const { logout, userEmail } = useAuth();
const navigate = useNavigate();
const handleLogout = () => {
 logout();
 navigate("/login");
 };
 return (
 <header className="relative z-60 h-16 bg-white border-b border-slate-200 flex items-center justify-between px-4 sm:px-8">
 <button
 type="button"
 onClick={onToggle}
 aria-label={isCollapsed ? "Abrir menú" : "Cerrar menú"}
 aria-expanded={!isCollapsed}
 className="p-2 text-slate-600 hover:bg-slate-100 rounded transition"
 >
 {isCollapsed ? <Menu size={24} /> : <X size={24} />}
 </button>
 <h2 className="hidden sm:block text-slate-600 font-medium text-lg">
 Panel de Administración
 </h2>
 <div className="flex items-end gap-6">
 <Link
 to="/carrito"
 className="relative p-2 hover:bg-slate-100 rounded-full transition"
 >
 <ShoppingCart size={22} aria-hidden="true" />
 {totalItems > 0 && (
 <span className="absolute top-0 right-0 bg-indigo-600 text-white 
text-xs font-bold w-5 h-5 flex items-center justify-center rounded-full 
transform translate-x-1 -translate-y-1">
 {totalItems}
 </span>
 )}
 </Link>
 <div className="flex items-center gap-4">
 <span className="text-sm text-slate-500">{userEmail}</span>
 {/* Contenedor relativo con la clase 'group' para detectar el hover */}
 <div className="relative group cursor-pointer pb-2">
 {/* Círculo del usuario / Avatar */}
 <div className="w-8 h-8 rounded-full bg-slate-200 overflow-hidden 
border border-slate-300 flex items-center justify-center">
 {/* 
 NOTA PARA LA API: 
 Aquí reemplazarás el 'src' quemado por la variable de tu estado, 
 por ejemplo: src={userAvatar || defaultImage} 
 */}
 <img
 
src="https://fastly.picsum.photos/id/64/4326/2884.jpg?hmac=9_SzX666YRpR
_fOyYStXpfSiJ_edO3ghlSRnH2w09Kg"
 alt="Avatar del usuario"
 className="w-full h-full object-cover"
 />
 </div>
 {/* Menú desplegable */}
 <div className="absolute right-0 top-full mt-1 w-36 bg-white border 
border-slate-200 rounded-lg shadow-lg opacity-0 invisible grouphover:opacity-100 group-hover:visible transition-all duration-200 z-50">
 <button
 onClick={handleLogout}
 className="w-full text-left px-4 py-2 text-sm text-red-600 fontsemibold hover:bg-red-50 rounded-md transition-colors"
 >
 Cerrar Sesión
 </button>
 </div>
 </div>
 </div>
 </div>
 </header>
 );
};
export default Navbar;