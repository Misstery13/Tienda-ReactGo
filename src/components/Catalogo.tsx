import { useEffect, useState } from 'react';
import { useCart, type Producto } from '../context/CartContext';
import { getProductos } from '../services/api';
const Catalogo = () => {
 const { addToCart } = useCart(); // <-- Usamos la función del contexto
 const [productos, setProductos] = useState<Producto[]>([]);
 const [loading, setLoading] = useState<boolean>(true);
 const [error, setError] = useState<string>('');

 // Carga de productos desde el backend
 useEffect(() => {
 const cargarProductos = async () => {
 try {
 const data = await getProductos();
 setProductos(data);
 } catch (err) {
 setError(err instanceof Error ? err.message : 'Error al conectar con el servidor');
 } finally {
 setLoading(false);
 }
 };
 cargarProductos();
 }, []);

 if (loading) {
 return <p className="text-slate-500">Cargando productos...</p>;
 }

 if (error) {
 return (
 <div className="bg-red-50 text-red-600 p-3 rounded-lg text-sm border border-red-200">
 {error}
 </div>
 );
 }

 return (
 <div>
 <h1 className="text-2xl font-bold text-slate-800 mb-6">Catálogo de 
Productos</h1>
 <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
 {productos.map((prod) => (
 <div key={prod.id} className="bg-white rounded-lg overflow-hidden 
border border-slate-200 shadow-sm flex flex-col">
 <img src={prod.img} alt={prod.nombre} className="w-full h-40 
object-cover" />
 <div className="p-4 flex flex-col flex-1">
 <h3 className="font-semibold text-slate-700">{prod.nombre}</h3>
 <p className="text-indigo-600 font-bold mt-2 mb4">${prod.precio.toFixed(2)}</p> 
 
 <button 
 onClick={() => addToCart(prod)}
 className="mt-auto w-full bg-slate-900 text-white py-2 rounded 
text-sm hover:bg-indigo-600 transition"
 >
 Añadir al Carrito
 </button>
 </div>
 </div>
 ))}
 </div>
 </div>
 );
};
export default Catalogo;
