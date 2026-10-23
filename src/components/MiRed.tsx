// src/components/MiRed.tsx
// Red propia del usuario: su subárbol, su enlace de invitación y su ledger.
import { useEffect, useState } from 'react';
import { Check, Copy } from 'lucide-react';
import {
  contarRed,
  sumarVentasRed,
  nivelAlcanzado,
  TASA_COMISION,
  type Referido,
} from '../data/red';
import { getMiRed, getMisComisiones, type ResumenComisiones } from '../services/redService';

const COLOR_ESTADO: Record<string, string> = {
  pendiente: 'bg-amber-100 text-amber-700',
  aprobada: 'bg-indigo-100 text-indigo-700',
  pagada: 'bg-green-100 text-green-700',
  revertida: 'bg-red-100 text-red-700',
};

const NIVEL_ESTILO: Record<number, { badge: string; border: string; tasa: string }> = {
  1: { badge: 'bg-teal-100 text-teal-700', border: 'border-teal-200', tasa: '10 %' },
  2: { badge: 'bg-indigo-100 text-indigo-700', border: 'border-indigo-200', tasa: '5 %' },
  3: { badge: 'bg-pink-100 text-pink-700', border: 'border-pink-200', tasa: '2 %' },
};

// Componente recursivo: dibuja un referido y debajo a sus hijos.
// La prop se llama "nodo" porque "ref" es una prop reservada en React 19.
const NodoReferido = ({ nodo }: { nodo: Referido }) => {
  const estilo = NIVEL_ESTILO[nodo.nivel] ?? NIVEL_ESTILO[3];

  return (
    <li>
      <div className={`bg-white rounded-lg border ${estilo.border} shadow-sm p-4 flex flex-wrap items-center justify-between gap-3`}>
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-full bg-slate-200 flex items-center justify-center font-bold text-slate-600">
            {nodo.nombre.charAt(0)}
          </div>
          <div>
            <p className="font-semibold text-slate-800">{nodo.nombre}</p>
            <span className={`inline-block mt-1 text-xs font-semibold px-2 py-0.5 rounded-full ${estilo.badge}`}>
              Nivel {nodo.nivel} · {estilo.tasa} comisión
            </span>
          </div>
        </div>
        <div className="text-right">
          <p className="text-sm text-slate-500">Ventas</p>
          <p className="font-bold text-slate-800">${nodo.ventas.toLocaleString()}</p>
        </div>
      </div>

      {nodo.hijos && nodo.hijos.length > 0 && (
        <ul className="mt-3 ml-6 space-y-3 border-l-2 border-slate-200 pl-4">
          {nodo.hijos.map((hijo) => (
            <NodoReferido key={hijo.id} nodo={hijo} />
          ))}
        </ul>
      )}
    </li>
  );
};

const MiRed = () => {
  const [red, setRed] = useState<Referido | null>(null);
  const [comisiones, setComisiones] = useState<ResumenComisiones | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string>('');
  const [copiado, setCopiado] = useState<boolean>(false);

  useEffect(() => {
    let vivo = true;

    Promise.all([getMiRed(), getMisComisiones()])
      .then(([arbol, ledger]) => {
        if (!vivo) return;
        setRed(arbol);
        setComisiones(ledger);
        setError('');
      })
      .catch((err: Error) => {
        if (vivo) setError(err.message);
      })
      .finally(() => {
        if (vivo) setLoading(false);
      });

    return () => {
      vivo = false;
    };
  }, []);

  const enlaceInvitacion = red?.codigo
    ? `${window.location.origin}/registro?ref=${red.codigo}`
    : '';

  const copiarEnlace = async () => {
    try {
      await navigator.clipboard.writeText(enlaceInvitacion);
      setCopiado(true);
      setTimeout(() => setCopiado(false), 2000);
    } catch {
      // Sin permiso de portapapeles el usuario puede copiar el texto a mano.
      setError('No se pudo copiar; selecciona el enlace manualmente.');
    }
  };

  if (loading) {
    return <p className="text-slate-500">Cargando tu red...</p>;
  }

  if (error && !red) {
    return (
      <div className="bg-red-50 text-red-600 p-3 rounded-lg text-sm border border-red-200">
        {error}
      </div>
    );
  }

  if (!red) return null;

  const referidosActivos = contarRed(red);
  const ventasRed = sumarVentasRed(red);
  const nivel = nivelAlcanzado(red.hijos?.length ?? 0);

  const resumen = [
    { etiqueta: 'Referidos en tu red', valor: referidosActivos.toString(), color: 'text-indigo-600' },
    { etiqueta: 'Ventas de la Red', valor: `$${ventasRed.toLocaleString()}`, color: 'text-indigo-600' },
    { etiqueta: 'Comisiones por cobrar', valor: `$${(comisiones?.pendiente ?? 0).toFixed(2)}`, color: 'text-green-600' },
    { etiqueta: 'Nivel Alcanzado', valor: nivel, color: 'text-amber-600' },
  ];

  return (
    <div>
      <h1 className="text-2xl font-bold text-slate-800 mb-2">Mi Red de Referidos</h1>
      <p className="text-slate-500 mb-6">
        Ganas el {TASA_COMISION[1] * 100} % de nivel 1, {TASA_COMISION[2] * 100} % de nivel 2 y {TASA_COMISION[3] * 100} % de nivel 3.
      </p>

      {error && (
        <div className="bg-red-50 text-red-600 p-3 rounded-lg text-sm mb-6 border border-red-200">
          {error}
        </div>
      )}

      {/* ===== ENLACE DE INVITACIÓN ===== */}
      {enlaceInvitacion && (
        <div className="bg-gradient-to-r from-teal-600 to-emerald-700 text-white rounded-xl p-6 mb-8">
          <p className="text-sm text-teal-50 mb-1">Tu enlace de invitación</p>
          <p className="text-xs text-teal-100 mb-4">
            Quien se registre con él quedará colgando de ti y te dará comisión.
          </p>
          <div className="flex flex-wrap items-center gap-3">
            <code className="flex-1 min-w-0 bg-white/15 rounded-lg px-4 py-3 text-sm font-mono break-all">
              {enlaceInvitacion}
            </code>
            <button
              onClick={copiarEnlace}
              className="min-h-11 inline-flex items-center gap-2 bg-white text-teal-700 font-semibold px-5 rounded-lg hover:bg-teal-50 transition"
            >
              {copiado
                ? <><Check size={18} aria-hidden="true" /> Copiado</>
                : <><Copy size={18} aria-hidden="true" /> Copiar</>}
            </button>
          </div>
          <p className="text-xs text-teal-100 mt-3">
            Tu código: <span className="font-mono font-bold">{red.codigo}</span>
          </p>
        </div>
      )}

      {/* ===== RESUMEN ===== */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        {resumen.map((item) => (
          <div key={item.etiqueta} className="bg-white p-6 rounded-lg shadow-sm border border-slate-100">
            <p className="text-sm text-slate-500 uppercase font-semibold">{item.etiqueta}</p>
            <p className={`text-3xl font-bold mt-2 ${item.color}`}>{item.valor}</p>
          </div>
        ))}
      </div>

      {/* ===== ÁRBOL ===== */}
      <div className="bg-white rounded-lg border border-slate-200 shadow-sm p-6 mb-8">
        <h2 className="text-lg font-bold text-slate-800 mb-4">Tu Red</h2>

        <div className="bg-slate-900 text-white rounded-lg p-4 mb-4 flex flex-wrap items-center justify-between gap-3">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-full bg-teal-500 flex items-center justify-center font-bold">
              {red.nombre.charAt(0)}
            </div>
            <div>
              <p className="font-semibold">{red.nombre}</p>
              <p className="text-xs text-slate-300">Nivel alcanzado: {nivel}</p>
            </div>
          </div>
          <p className="font-bold text-green-400">
            Por cobrar: ${(comisiones?.pendiente ?? 0).toFixed(2)}
          </p>
        </div>

        {red.hijos && red.hijos.length > 0 ? (
          <ul className="space-y-3">
            {red.hijos.map((hijo) => (
              <NodoReferido key={hijo.id} nodo={hijo} />
            ))}
          </ul>
        ) : (
          <p className="text-slate-500 text-sm">
            Todavía no tienes referidos. Comparte tu enlace para empezar.
          </p>
        )}
      </div>

      {/* ===== LEDGER ===== */}
      <div className="bg-white rounded-lg border border-slate-200 shadow-sm p-6">
        <h2 className="text-lg font-bold text-slate-800 mb-1">Mis Comisiones</h2>
        <p className="text-sm text-slate-500 mb-4">
          Cada línea es una venta de tu red. Esta es la fuente de verdad de lo que se te debe.
        </p>

        <div className="grid grid-cols-3 gap-4 mb-6">
          <div className="bg-amber-50 rounded-lg p-4 border border-amber-100">
            <p className="text-xs text-amber-700 uppercase font-semibold">Por cobrar</p>
            <p className="text-xl font-bold text-amber-700">${(comisiones?.pendiente ?? 0).toFixed(2)}</p>
          </div>
          <div className="bg-green-50 rounded-lg p-4 border border-green-100">
            <p className="text-xs text-green-700 uppercase font-semibold">Pagadas</p>
            <p className="text-xl font-bold text-green-700">${(comisiones?.pagada ?? 0).toFixed(2)}</p>
          </div>
          <div className="bg-red-50 rounded-lg p-4 border border-red-100">
            <p className="text-xs text-red-700 uppercase font-semibold">Revertidas</p>
            <p className="text-xl font-bold text-red-700">${(comisiones?.revertida ?? 0).toFixed(2)}</p>
          </div>
        </div>

        {comisiones && comisiones.lineas.length > 0 ? (
          <div className="space-y-2">
            {comisiones.lineas.map((l) => (
              <div key={l.id} className="flex flex-wrap items-center justify-between gap-3 p-3 rounded-lg border border-slate-100">
                <div>
                  <p className="font-mono text-sm font-semibold text-slate-800">{l.pedido}</p>
                  <p className="text-xs text-slate-500">
                    Nivel {l.nivel} · {(l.porcentaje * 100).toFixed(0)} % · {l.fecha}
                  </p>
                </div>
                <div className="flex items-center gap-3">
                  <span className={`text-xs font-semibold px-2 py-1 rounded-full uppercase ${COLOR_ESTADO[l.estado]}`}>
                    {l.estado}
                  </span>
                  <span className="font-bold text-slate-800 w-20 text-right">${l.monto.toFixed(2)}</span>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <p className="text-slate-500 text-sm">
            Todavía no has generado comisiones. Se crean cuando alguien de tu red compra.
          </p>
        )}
      </div>
    </div>
  );
};

export default MiRed;
