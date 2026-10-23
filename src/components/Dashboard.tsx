// src/components/Dashboard.tsx
// Dashboard del negocio: KPIs y rankings calculados por el servidor
// con SQL de agregación, no derivados de datos semilla.
import { useEffect, useState } from 'react';
import {
  getResumen,
  getTopProductos,
  getTopReferidos,
  type FilaReporte,
  type ResumenVentas,
} from '../services/reportesService';

const Dashboard = () => {
  const [resumen, setResumen] = useState<ResumenVentas | null>(null);
  const [productos, setProductos] = useState<FilaReporte[]>([]);
  const [referidos, setReferidos] = useState<FilaReporte[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string>('');

  useEffect(() => {
    let vivo = true;

    Promise.all([getResumen(), getTopProductos(5), getTopReferidos(5)])
      .then(([r, p, d]) => {
        if (!vivo) return;
        setResumen(r);
        setProductos(p);
        setReferidos(d);
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

  if (loading) {
    return <p className="text-slate-500">Cargando resumen...</p>;
  }

  if (error || !resumen) {
    return (
      <div className="bg-red-50 text-red-600 p-3 rounded-lg text-sm border border-red-200">
        {error || 'No se pudo cargar el resumen'}
      </div>
    );
  }

  const kpis = [
    { etiqueta: 'Total Vendido', valor: `$${resumen.total_vendido.toFixed(2)}`, color: 'text-indigo-600' },
    { etiqueta: 'Pedidos', valor: resumen.pedidos.toString(), color: 'text-indigo-600' },
    { etiqueta: 'Ticket Promedio', valor: `$${resumen.ticket_promedio.toFixed(2)}`, color: 'text-teal-600' },
    { etiqueta: 'Comisiones Devengadas', valor: `$${resumen.comisiones_devengadas.toFixed(2)}`, color: 'text-green-600' },
  ];

  return (
    <div>
      <h1 className="text-2xl font-bold text-slate-800 mb-2">Resumen General</h1>
      <p className="text-slate-500 mb-6">
        Datos reales del negocio
        {resumen.cancelados > 0 && ` · ${resumen.cancelados} pedido(s) cancelado(s) excluido(s)`}
      </p>

      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-6 mb-8">
        {kpis.map((kpi) => (
          <div key={kpi.etiqueta} className="bg-white p-6 rounded-lg shadow-sm border border-slate-100">
            <p className="text-sm text-slate-500 uppercase font-semibold">{kpi.etiqueta}</p>
            <p className={`text-3xl font-bold mt-2 ${kpi.color}`}>{kpi.valor}</p>
          </div>
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* ===== TOP PRODUCTOS ===== */}
        <div className="bg-white p-6 rounded-lg shadow-sm border border-slate-100">
          <h2 className="text-lg font-bold text-slate-800 mb-4">Productos más vendidos</h2>
          {productos.length === 0 ? (
            <p className="text-slate-500 text-sm">Todavía no hay ventas registradas.</p>
          ) : (
            <div className="space-y-3">
              {productos.map((p, idx) => (
                <div key={p.id} className="flex items-center justify-between p-3 rounded-lg border border-slate-100">
                  <div className="flex items-center gap-3">
                    <span className="w-8 h-8 rounded-full bg-teal-100 text-teal-700 font-bold flex items-center justify-center">
                      {idx + 1}
                    </span>
                    <div>
                      <p className="font-semibold text-slate-800">{p.nombre}</p>
                      <p className="text-xs text-slate-500">{p.cantidad} unidad(es)</p>
                    </div>
                  </div>
                  <p className="font-bold text-slate-800">${p.total.toFixed(2)}</p>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* ===== TOP REFERIDOS ===== */}
        <div className="bg-white p-6 rounded-lg shadow-sm border border-slate-100">
          <h2 className="text-lg font-bold text-slate-800 mb-4">Distribuidores que más ganan</h2>
          {referidos.length === 0 ? (
            <p className="text-slate-500 text-sm">Todavía no se han generado comisiones.</p>
          ) : (
            <div className="space-y-3">
              {referidos.map((d, idx) => (
                <div key={d.id} className="flex items-center justify-between p-3 rounded-lg border border-slate-100">
                  <div className="flex items-center gap-3">
                    <span className="w-8 h-8 rounded-full bg-indigo-100 text-indigo-700 font-bold flex items-center justify-center">
                      {idx + 1}
                    </span>
                    <div>
                      <p className="font-semibold text-slate-800">{d.nombre}</p>
                      <p className="text-xs text-slate-500">{d.cantidad} comisión(es)</p>
                    </div>
                  </div>
                  <p className="font-bold text-green-600">${d.total.toFixed(2)}</p>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
