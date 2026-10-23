// src/components/AdminPedidos.tsx
// Gestión de pedidos: filtro por estado y avance del ciclo de vida.
import { useEffect, useState } from 'react';
import {
  cambiarEstado,
  COLOR_ESTADO,
  getPedido,
  getPedidos,
  SIGUIENTES,
  type DetallePedido,
  type EstadoPedido,
  type ResumenPedido,
} from '../services/pedidosService';

const ESTADOS: EstadoPedido[] = ['pendiente', 'preparando', 'enviado', 'entregado', 'cancelado'];

const AdminPedidos = () => {
  const [pedidos, setPedidos] = useState<ResumenPedido[]>([]);
  const [total, setTotal] = useState<number>(0);
  const [pagina, setPagina] = useState<number>(1);
  const [filtro, setFiltro] = useState<string>('');
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string>('');
  const [recarga, setRecarga] = useState<number>(0);

  const [abierto, setAbierto] = useState<number | null>(null);
  const [detalle, setDetalle] = useState<DetallePedido | null>(null);

  const LIMITE = 10;

  useEffect(() => {
    let vivo = true;

    getPedidos(filtro, pagina, LIMITE)
      .then((data) => {
        if (!vivo) return;
        setPedidos(data.datos);
        setTotal(data.total);
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
  }, [filtro, pagina, recarga]);

  const verDetalle = async (id: number) => {
    if (abierto === id) {
      setAbierto(null);
      return;
    }
    setAbierto(id);
    setDetalle(null);
    try {
      setDetalle(await getPedido(id));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'No se pudo cargar el detalle');
    }
  };

  const avanzar = async (p: ResumenPedido, nuevo: EstadoPedido) => {
    if (nuevo === 'cancelado' && !confirm(`Cancelar ${p.numero} devolverá el stock. ¿Continuar?`)) {
      return;
    }

    try {
      await cambiarEstado(p.id, nuevo);
      setRecarga((n) => n + 1);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'No se pudo cambiar el estado');
    }
  };

  const totalPaginas = Math.max(Math.ceil(total / LIMITE), 1);

  return (
    <div>
      <h1 className="text-2xl font-bold text-slate-800 mb-2">Pedidos</h1>
      <p className="text-slate-500 mb-6">{total} pedido(s)</p>

      {error && (
        <div className="bg-red-50 text-red-600 p-3 rounded-lg text-sm mb-6 border border-red-200">
          {error}
        </div>
      )}

      {/* ===== FILTRO POR ESTADO ===== */}
      <div className="flex flex-wrap gap-2 mb-6">
        <button
          onClick={() => { setFiltro(''); setPagina(1); }}
          className={`min-h-11 px-4 rounded-lg border text-sm font-semibold transition ${filtro === '' ? 'bg-slate-900 text-white border-slate-900' : 'border-slate-300 text-slate-600 hover:bg-slate-50'}`}
        >
          Todos
        </button>
        {ESTADOS.map((e) => (
          <button
            key={e}
            onClick={() => { setFiltro(e); setPagina(1); }}
            className={`min-h-11 px-4 rounded-lg border text-sm font-semibold capitalize transition ${filtro === e ? 'bg-slate-900 text-white border-slate-900' : 'border-slate-300 text-slate-600 hover:bg-slate-50'}`}
          >
            {e}
          </button>
        ))}
      </div>

      {loading ? (
        <p className="text-slate-500">Cargando pedidos...</p>
      ) : pedidos.length === 0 ? (
        <div className="bg-white p-10 rounded-lg border border-slate-200 text-center text-slate-500">
          No hay pedidos con ese estado.
        </div>
      ) : (
        <div className="space-y-3">
          {pedidos.map((p) => (
            <div key={p.id} className="bg-white rounded-lg border border-slate-200 shadow-sm overflow-hidden">
              <div className="p-4 flex flex-wrap items-center justify-between gap-3">
                <button onClick={() => verDetalle(p.id)} className="text-left min-h-11">
                  <p className="font-mono font-semibold text-slate-800">{p.numero}</p>
                  <p className="text-xs text-slate-500">{p.cliente} · {p.fecha}</p>
                </button>

                <div className="flex flex-wrap items-center gap-3">
                  <span className={`text-xs font-semibold px-2 py-1 rounded-full uppercase ${COLOR_ESTADO[p.estado]}`}>
                    {p.estado}
                  </span>
                  <span className="font-bold text-indigo-600 w-20 text-right">${p.total.toFixed(2)}</span>

                  {/* Solo se ofrecen las transiciones que el backend acepta */}
                  {SIGUIENTES[p.estado].map((siguiente) => (
                    <button
                      key={siguiente}
                      onClick={() => avanzar(p, siguiente)}
                      className={`min-h-11 px-3 rounded-lg text-xs font-semibold capitalize transition ${siguiente === 'cancelado' ? 'border border-red-300 text-red-600 hover:bg-red-50' : 'bg-indigo-600 text-white hover:bg-indigo-700'}`}
                    >
                      {siguiente}
                    </button>
                  ))}
                </div>
              </div>

              {abierto === p.id && (
                <div className="border-t border-slate-100 p-4 bg-slate-50">
                  {detalle === null ? (
                    <p className="text-sm text-slate-500">Cargando detalle...</p>
                  ) : (
                    <>
                      <p className="text-xs text-slate-500 mb-3">
                        {detalle.cliente.email} · {detalle.cliente.ciudad}
                      </p>
                      {detalle.items.map((i) => (
                        <div key={i.producto_id} className="flex justify-between text-sm">
                          <span className="text-slate-600">{i.nombre} × {i.cantidad}</span>
                          <span className="font-semibold text-slate-800">${i.subtotal.toFixed(2)}</span>
                        </div>
                      ))}
                    </>
                  )}
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      {totalPaginas > 1 && (
        <div className="flex items-center justify-center gap-4 mt-6">
          <button
            onClick={() => setPagina((n) => Math.max(n - 1, 1))}
            disabled={pagina === 1}
            className="min-h-11 px-4 border border-slate-300 rounded-lg disabled:opacity-40 hover:bg-slate-50 transition"
          >
            Anterior
          </button>
          <span className="text-sm text-slate-600">Página {pagina} de {totalPaginas}</span>
          <button
            onClick={() => setPagina((n) => Math.min(n + 1, totalPaginas))}
            disabled={pagina === totalPaginas}
            className="min-h-11 px-4 border border-slate-300 rounded-lg disabled:opacity-40 hover:bg-slate-50 transition"
          >
            Siguiente
          </button>
        </div>
      )}
    </div>
  );
};

export default AdminPedidos;
