import { useState, type FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { loginRequest } from '../services/api';
const Login = () => {
 const [email, setEmail] = useState<string>('');
 const [password, setPassword] = useState<string>('');
 const [error, setError] = useState<string>('');
 const [loading, setLoading] = useState<boolean>(false);
 
 const navigate = useNavigate();
 const { login } = useAuth();
 const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
 e.preventDefault();
 setError('');
 setLoading(true);

 try {
 const data = await loginRequest(email, password);
 login(data.email, data.token); // Cambiamos el estado global a autenticado
 navigate('/'); // Redirigimos al Dashboard
 } catch (err) {
 setError(err instanceof Error ? err.message : 'Error al conectar con el servidor');
 } finally {
 setLoading(false);
 }
 };
 return (
 <div className="min-h-screen flex items-center justify-center bg-slate-100 
px-4">
 <div className="max-w-md w-full bg-white rounded-xl shadow-md p-8 
border border-slate-200">
 <div className="text-center mb-8">
 <h2 className="text-3xl font-bold text-slate-900">MultiCatálogo</h2>
 <p className="text-slate-500 mt-2">Ingresa a tu cuenta para 
continuar</p>
 </div>
 {error && (
 <div className="bg-red-50 text-red-600 p-3 rounded-lg text-sm mb-6 
text-center border border-red-200">
 {error}
 </div>
 )}
 <form onSubmit={handleSubmit} className="space-y-6">
 <div>
 <label className="block text-sm font-medium text-slate-700 mb2">Correo Electrónico</label>
 <input
 type="email"
 value={email}
 onChange={(e) => setEmail(e.target.value)}
 className="w-full px-4 py-3 rounded-lg border border-slate-300
focus:ring-2 focus:ring-indigo-600 focus:border-indigo-600 outline-none 
transition"
 placeholder="admin@upse.edu.ec"
 required
 />
 </div>
 <div>
 <label className="block text-sm font-medium text-slate-700 mb2">Contraseña</label>
 <input
 type="password"
 value={password}
 onChange={(e) => setPassword(e.target.value)}
 className="w-full px-4 py-3 rounded-lg border border-slate-300 
focus:ring-2 focus:ring-indigo-600 focus:border-indigo-600 outline-none 
transition"
 placeholder="••••••"
 required
 />
 </div>
 <button
 type="submit"
 disabled={loading}
 className="w-full bg-indigo-600 text-white font-bold py-3 rounded-lg 
hover:bg-indigo-700 transition disabled:opacity-60 disabled:cursor-not-allowed"
 >
 {loading ? 'Ingresando...' : 'Iniciar Sesión'}
 </button>
 </form>
 </div>
 </div>
 );
};
export default Login;
