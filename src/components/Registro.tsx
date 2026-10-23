// src/components/Registro.tsx
// Registro público de clientes. Si la URL trae ?ref=CODIGO, el nuevo usuario
// queda colgando de quien lo invitó dentro de la red multinivel.
import { useState, type FormEvent } from 'react';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';
import { useAuth } from '../context/auth';
import { registrar } from '../services/usuariosService';

const Registro = () => {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { login } = useAuth();

  const [nombre, setNombre] = useState<string>('');
  const [email, setEmail] = useState<string>('');
  const [password, setPassword] = useState<string>('');
  // El código llega prellenado desde el enlace de invitación.
  const [codigo, setCodigo] = useState<string>(searchParams.get('ref') ?? '');
  const [error, setError] = useState<string>('');
  const [loading, setLoading] = useState<boolean>(false);

  const invitado = searchParams.get('ref') !== null;

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      const datos = await registrar({
        nombre,
        email,
        password,
        codigo_referido: codigo.trim() || undefined,
      });

      // El backend devuelve la sesión iniciada: se entra directo a la tienda.
      login({ email: datos.email, rol: datos.rol }, datos.token);
      navigate('/tienda');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'No se pudo crear la cuenta');
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-slate-100 px-4 py-10">
      <div className="max-w-md w-full bg-white rounded-xl shadow-md p-8 border border-slate-200">
        <div className="text-center mb-8">
          <h2 className="text-3xl font-bold text-slate-900">Crear cuenta</h2>
          <p className="text-slate-500 mt-2">
            {invitado
              ? 'Te invitaron a unirte a la red de MultiCatálogo'
              : 'Únete a MultiCatálogo y empieza tu red'}
          </p>
        </div>

        {error && (
          <div className="bg-red-50 text-red-600 p-3 rounded-lg text-sm mb-6 text-center border border-red-200">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-5">
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-2">
              Nombre completo
            </label>
            <input
              type="text"
              value={nombre}
              onChange={(e) => setNombre(e.target.value)}
              className="w-full px-4 py-3 rounded-lg border border-slate-300 focus:ring-2 focus:ring-indigo-600 focus:border-indigo-600 outline-none transition"
              placeholder="Tu nombre"
              required
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 mb-2">
              Correo Electrónico
            </label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full px-4 py-3 rounded-lg border border-slate-300 focus:ring-2 focus:ring-indigo-600 focus:border-indigo-600 outline-none transition"
              placeholder="tucorreo@upse.edu.ec"
              required
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 mb-2">
              Contraseña
            </label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              minLength={6}
              className="w-full px-4 py-3 rounded-lg border border-slate-300 focus:ring-2 focus:ring-indigo-600 focus:border-indigo-600 outline-none transition"
              placeholder="Mínimo 6 caracteres"
              required
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 mb-2">
              Código de referido <span className="text-slate-400">(opcional)</span>
            </label>
            <input
              type="text"
              value={codigo}
              onChange={(e) => setCodigo(e.target.value.toUpperCase())}
              readOnly={invitado}
              className={`w-full px-4 py-3 rounded-lg border border-slate-300 font-mono focus:ring-2 focus:ring-indigo-600 focus:border-indigo-600 outline-none transition ${invitado ? 'bg-slate-50 text-slate-500' : ''}`}
              placeholder="ANAGARCI-002"
            />
            <p className="text-xs text-slate-500 mt-1">
              {invitado
                ? 'Quedarás registrado en la red de quien te invitó.'
                : 'Si alguien te invitó, escribe su código para unirte a su red.'}
            </p>
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full min-h-11 bg-indigo-600 text-white font-bold py-3 rounded-lg hover:bg-indigo-700 transition disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {loading ? 'Creando cuenta...' : 'Crear cuenta'}
          </button>
        </form>

        <p className="text-center text-sm text-slate-500 mt-6">
          ¿Ya tienes cuenta?{' '}
          <Link to="/login" className="text-indigo-600 font-semibold hover:underline">
            Inicia sesión
          </Link>
        </p>
      </div>
    </div>
  );
};

export default Registro;
