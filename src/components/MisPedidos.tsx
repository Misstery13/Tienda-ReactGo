// src/components/MisPedidos.tsx
// Historial de pedidos del cliente, con detalle desplegable.
import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { ChevronDown, ChevronUp } from 'lucide-react';
import {
  COLOR_ESTADO,
  getMisPedidos,
  getPedido,
  type DetallePedido,
  type ResumenPedido,
} from '../services/pedidosService';

const MisPedidos = () => {
  const [pedidos, setPedidos] = useState<ResumenPedido[]>([]);
  const [total, setTotal] = useState<number>(0);
  const [pagina, setPagina] = useState<number>(1);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string>('');

  // Detalle cargado bajo demanda al desplegar una fila.
  const [abierto, setAbierto] = useState<number | null>(null);
  const [detalle, setDetalle] = useState<DetallePedido | null>(null);

  const LIMITE = 10;

  useEffect(() => {
    let vivo = true;

    getMisPedidos(pagina, LIMITE)
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
  }, [pagina]);

  const alternar = async (id: number) => {
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

  const totalPaginas = Math.max(Math.ceil(total / LIMITE), 1);

  if (loading) {
    return <p className="text-slate-500">Cargando tus pedidos...</p>;
  }

  return (
    <div>
      <h1 className="text-2xl font-bold text-slate-800 mb-2">Mis Pedidos</h1>
      <p className="text-slate-500 mb-6">{total} pedido(s)</p>

      {error && (
        <div className="bg-red-50 text-red-600 p-3 rounded-lg text-sm mb-6 border border-red-200">
          {error}
        </div>
      )}

      {pedidos.length === 0 ? (
        <div className="bg-white p-10 rounded-lg border border-slate-200 text-center">
          <p className="text-slate-500 mb-4">Todavía no has hecho ningún pedido.</p>
          <Link
            to="/catalogo"
            className="inline-block min-h-11 bg-indigo-600 text-white font-semibold px-6 py-3 rounded-lg hover:bg-indigo-700 transition"
          >
            Explorar Catálogo
          </Link>
        </div>
      ) : (
        <div className="space-y-3">
          {pedidos.map((p) => (
            <div key={p.id} className="bg-white rounded-lg border border-slate-200 shadow-sm overflow-hidden">
              <button
                onClick={() => alternar(p.id)}
                className="w-full min-h-11 p-4 flex flex-wrap items-center justify-between gap-3 text-left hover:bg-slate-50 transition"
              >
                <div>
                  <p className="font-mono font-semibold text-slate-800">{p.numero}</p>
                  <p className="text-xs text-slate-500">{p.fecha} · {p.articulos} artículo(s)</p>
                </div>
                <div className="flex items-center gap-3">
                  <span className={`text-xs font-semibold px-2 py-1 rounded-full uppercase ${COLOR_ESTADO[p.estado]}`}>
                    {p.estado}
                  </span>
                  <span className="font-bold text-indigo-600 w-20 text-right">${p.total.toFixed(2)}</span>
                  {abierto === p.id
                    ? <ChevronUp size={18} className="text-slate-400" aria-hidden="true" />
                    : <ChevronDown size={18} className="text-slate-400" aria-hidden="true" />}
                </div>
              </button>

              {abierto === p.id && (
                <div className="border-t border-slate-100 p-4 bg-slate-50">
                  {detalle === null ? (
                    <p className="text-sm text-slate-500">Cargando detalle...</p>
                  ) : (
                    <>
                      <p className="text-xs text-slate-500 mb-3">
                        Envío a {detalle.cliente.nombre} · {detalle.cliente.ciudad}
                      </p>
                      <div className="space-y-2">
                        {detalle.items.map((i) => (
                          <div key={i.producto_id} className="flex justify-between text-sm">
                            <span className="text-slate-600">{i.nombre} × {i.cantidad}</span>
                            <span className="font-semibold text-slate-800">${i.subtotal.toFixed(2)}</span>
                          </div>
                        ))}
                      </div>
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

export default MisPedidos;
