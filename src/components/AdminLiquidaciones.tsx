// src/components/AdminLiquidaciones.tsx
// Deuda con los distribuidores y pago de comisiones pendientes.
import { useEffect, useState } from 'react';
import { Wallet } from 'lucide-react';
import {
  getDeuda,
  getLiquidaciones,
  liquidar,
  type DeudaDistribuidor,
  type Liquidacion,
} from '../services/reportesService';

const AdminLiquidaciones = () => {
  const [deudas, setDeudas] = useState<DeudaDistribuidor[]>([]);
  const [historial, setHistorial] = useState<Liquidacion[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string>('');
  const [pagando, setPagando] = useState<number | null>(null);
  const [recarga, setRecarga] = useState<number>(0);

  useEffect(() => {
    let vivo = true;

    Promise.all([getDeuda(), getLiquidaciones()])
      .then(([d, h]) => {
        if (!vivo) return;
        setDeudas(d);
        setHistorial(h);
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
  }, [recarga]);

  const pagar = async (d: DeudaDistribuidor) => {
    if (!confirm(`¿Pagar $${d.pendiente.toFixed(2)} a ${d.nombre}?`)) return;

    setPagando(d.referido_id);
    try {
      await liquidar(d.referido_id);
      setRecarga((n) => n + 1);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'No se pudo liquidar');
    } finally {
      setPagando(null);
    }
  };

  const deudaTotal = deudas.reduce((suma, d) => suma + d.pendiente, 0);

  if (loading) {
    return <p className="text-slate-500">Cargando liquidaciones...</p>;
  }

  return (
    <div>
      <h1 className="text-2xl font-bold text-slate-800 mb-2">Liquidaciones</h1>
      <p className="text-slate-500 mb-6">
        Deuda total con distribuidores: <span className="font-bold text-amber-700">${deudaTotal.toFixed(2)}</span>
      </p>

      {error && (
        <div className="bg-red-50 text-red-600 p-3 rounded-lg text-sm mb-6 border border-red-200">
          {error}
        </div>
      )}

      {/* ===== DEUDA PENDIENTE ===== */}
      <div className="bg-white rounded-lg border border-slate-200 shadow-sm p-6 mb-8">
        <h2 className="text-lg font-bold text-slate-800 mb-4">Por pagar</h2>

        {deudas.length === 0 ? (
          <p className="text-slate-500 text-sm">No hay comisiones pendientes de pago.</p>
        ) : (
          <div className="space-y-3">
            {deudas.map((d) => (
              <div key={d.referido_id} className="flex flex-wrap items-center justify-between gap-3 p-4 rounded-lg border border-amber-100 bg-amber-50/40">
                <div>
                  <p className="font-semibold text-slate-800">{d.nombre}</p>
                  <p className="text-xs text-slate-500 font-mono">{d.codigo} · {d.lineas} comisión(es)</p>
                </div>
                <div className="flex items-center gap-4">
                  <p className="font-bold text-amber-700 text-lg">${d.pendiente.toFixed(2)}</p>
                  <button
                    onClick={() => pagar(d)}
                    disabled={pagando === d.referido_id}
                    className="min-h-11 inline-flex items-center gap-2 bg-green-600 text-white font-semibold px-5 rounded-lg hover:bg-green-700 transition disabled:opacity-50"
                  >
                    <Wallet size={18} aria-hidden="true" />
                    {pagando === d.referido_id ? 'Pagando...' : 'Liquidar'}
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* ===== HISTORIAL ===== */}
      <div className="bg-white rounded-lg border border-slate-200 shadow-sm p-6">
        <h2 className="text-lg font-bold text-slate-800 mb-4">Pagos realizados</h2>

        {historial.length === 0 ? (
          <p className="text-slate-500 text-sm">Todavía no se ha liquidado a nadie.</p>
        ) : (
          <div className="space-y-2">
            {historial.map((l) => (
              <div key={l.id} className="flex flex-wrap items-center justify-between gap-3 p-3 rounded-lg border border-slate-100">
                <div>
                  <p className="font-semibold text-slate-800">{l.nombre}</p>
                  <p className="text-xs text-slate-500">
                    Liquidación #{l.id} · {l.comisiones} comisión(es) · {l.fecha}
                  </p>
                </div>
                <div className="flex items-center gap-3">
                  <span className="text-xs font-semibold px-2 py-1 rounded-full uppercase bg-green-100 text-green-700">
                    {l.estado}
                  </span>
                  <span className="font-bold text-slate-800 w-24 text-right">${l.total.toFixed(2)}</span>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default AdminLiquidaciones;
