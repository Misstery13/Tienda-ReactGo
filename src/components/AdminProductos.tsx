// src/components/AdminProductos.tsx
// Administración del catálogo: alta, edición y borrado lógico.
import { useEffect, useState, type FormEvent } from 'react';
import { Pencil, Plus, Power, Search } from 'lucide-react';
import {
  actualizarProducto,
  crearProducto,
  desactivarProducto,
  getProductosAdmin,
  type ProductoAdmin,
  type ProductoDatos,
} from '../services/productosService';

const FORMULARIO_VACIO: ProductoDatos = {
  nombre: '',
  descripcion: '',
  precio: 0,
  categoria: '',
  img: '',
  stock: 0,
};

const AdminProductos = () => {
  const [productos, setProductos] = useState<ProductoAdmin[]>([]);
  const [total, setTotal] = useState<number>(0);
  const [pagina, setPagina] = useState<number>(1);
  const [busqueda, setBusqueda] = useState<string>('');
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string>('');

  // null = formulario cerrado; 0 = alta nueva; >0 = editando ese id
  const [editando, setEditando] = useState<number | null>(null);
  const [form, setForm] = useState<ProductoDatos>(FORMULARIO_VACIO);
  const [guardando, setGuardando] = useState<boolean>(false);

  // Cambiar este contador vuelve a pedir la lista tras guardar o desactivar.
  const [recarga, setRecarga] = useState<number>(0);
  const cargar = () => setRecarga((n) => n + 1);

  const LIMITE = 10;

  useEffect(() => {
    // Sin setState síncrono aquí: el efecto solo lanza la petición.
    let vivo = true;

    getProductosAdmin(busqueda, pagina, LIMITE)
      .then((data) => {
        if (!vivo) return;
        setProductos(data.datos);
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
  }, [busqueda, pagina, recarga]);

  const abrirAlta = () => {
    setForm(FORMULARIO_VACIO);
    setEditando(0);
  };

  const abrirEdicion = (p: ProductoAdmin) => {
    setForm({
      nombre: p.nombre,
      descripcion: p.descripcion,
      precio: p.precio,
      categoria: p.categoria,
      img: p.img,
      stock: p.stock,
    });
    setEditando(p.id);
  };

  const guardar = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setGuardando(true);
    setError('');

    try {
      if (editando === 0) {
        await crearProducto(form);
      } else if (editando !== null) {
        await actualizarProducto(editando, form);
      }
      setEditando(null);
      cargar();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'No se pudo guardar');
    } finally {
      setGuardando(false);
    }
  };

  const alternarEstado = async (p: ProductoAdmin) => {
    const accion = p.activo ? 'desactivar' : 'reactivar';
    if (!confirm(`¿Seguro que quieres ${accion} "${p.nombre}"?`)) return;

    try {
      // Desactivar usa el borrado lógico; reactivar es una edición normal.
      if (p.activo) {
        await desactivarProducto(p.id);
      } else {
        await actualizarProducto(p.id, { activo: true });
      }
      cargar();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'No se pudo cambiar el estado');
    }
  };

  const totalPaginas = Math.max(Math.ceil(total / LIMITE), 1);

  return (
    <div>
      <div className="flex flex-wrap items-center justify-between gap-4 mb-6">
        <div>
          <h1 className="text-2xl font-bold text-slate-800">Catálogo</h1>
          <p className="text-slate-500 text-sm">{total} producto(s) en total</p>
        </div>
        <button
          onClick={abrirAlta}
          className="min-h-11 inline-flex items-center gap-2 bg-indigo-600 text-white font-semibold px-5 rounded-lg hover:bg-indigo-700 transition"
        >
          <Plus size={18} aria-hidden="true" /> Nuevo producto
        </button>
      </div>

      {error && (
        <div className="bg-red-50 text-red-600 p-3 rounded-lg text-sm mb-6 border border-red-200">
          {error}
        </div>
      )}

      {/* ===== BUSCADOR ===== */}
      <div className="relative mb-6">
        <Search size={18} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" aria-hidden="true" />
        <input
          type="text"
          value={busqueda}
          onChange={(e) => {
            setBusqueda(e.target.value);
            setPagina(1);
          }}
          placeholder="Buscar por nombre o categoría..."
          className="w-full min-h-11 pl-10 pr-4 py-3 rounded-lg border border-slate-300 focus:ring-2 focus:ring-indigo-600 outline-none transition"
        />
      </div>

      {/* ===== FORMULARIO ===== */}
      {editando !== null && (
        <form onSubmit={guardar} className="bg-white p-6 rounded-lg border border-slate-200 shadow-sm mb-6 space-y-4">
          <h2 className="text-lg font-bold text-slate-800">
            {editando === 0 ? 'Nuevo producto' : `Editando producto #${editando}`}
          </h2>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-slate-700 mb-1">Nombre *</label>
              <input
                type="text"
                required
                value={form.nombre ?? ''}
                onChange={(e) => setForm({ ...form, nombre: e.target.value })}
                className="w-full min-h-11 px-4 rounded-lg border border-slate-300 focus:ring-2 focus:ring-indigo-600 outline-none"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-slate-700 mb-1">Categoría *</label>
              <input
                type="text"
                required
                value={form.categoria ?? ''}
                onChange={(e) => setForm({ ...form, categoria: e.target.value })}
                className="w-full min-h-11 px-4 rounded-lg border border-slate-300 focus:ring-2 focus:ring-indigo-600 outline-none"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-slate-700 mb-1">Precio *</label>
              <input
                type="number"
                step="0.01"
                min="0.01"
                required
                value={form.precio ?? 0}
                onChange={(e) => setForm({ ...form, precio: Number(e.target.value) })}
                className="w-full min-h-11 px-4 rounded-lg border border-slate-300 focus:ring-2 focus:ring-indigo-600 outline-none"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-slate-700 mb-1">Stock</label>
              <input
                type="number"
                min="0"
                value={form.stock ?? 0}
                onChange={(e) => setForm({ ...form, stock: Number(e.target.value) })}
                className="w-full min-h-11 px-4 rounded-lg border border-slate-300 focus:ring-2 focus:ring-indigo-600 outline-none"
              />
            </div>
            <div className="md:col-span-2">
              <label className="block text-sm font-medium text-slate-700 mb-1">Imagen (URL)</label>
              <input
                type="url"
                value={form.img ?? ''}
                onChange={(e) => setForm({ ...form, img: e.target.value })}
                placeholder="https://..."
                className="w-full min-h-11 px-4 rounded-lg border border-slate-300 focus:ring-2 focus:ring-indigo-600 outline-none"
              />
            </div>
            <div className="md:col-span-2">
              <label className="block text-sm font-medium text-slate-700 mb-1">Descripción</label>
              <textarea
                rows={3}
                value={form.descripcion ?? ''}
                onChange={(e) => setForm({ ...form, descripcion: e.target.value })}
                className="w-full px-4 py-3 rounded-lg border border-slate-300 focus:ring-2 focus:ring-indigo-600 outline-none"
              />
            </div>
          </div>

          <div className="flex flex-wrap gap-3">
            <button
              type="submit"
              disabled={guardando}
              className="min-h-11 bg-indigo-600 text-white font-semibold px-6 rounded-lg hover:bg-indigo-700 transition disabled:opacity-50"
            >
              {guardando ? 'Guardando...' : 'Guardar'}
            </button>
            <button
              type="button"
              onClick={() => setEditando(null)}
              className="min-h-11 border border-slate-300 text-slate-700 font-semibold px-6 rounded-lg hover:bg-slate-50 transition"
            >
              Cancelar
            </button>
          </div>
        </form>
      )}

      {/* ===== LISTADO ===== */}
      {loading ? (
        <p className="text-slate-500">Cargando catálogo...</p>
      ) : productos.length === 0 ? (
        <div className="bg-white p-10 rounded-lg border border-slate-200 text-center text-slate-500">
          No hay productos que coincidan.
        </div>
      ) : (
        <div className="space-y-3">
          {productos.map((p) => (
            <div
              key={p.id}
              className={`bg-white p-4 rounded-lg border shadow-sm flex flex-wrap items-center justify-between gap-4 ${p.activo ? 'border-slate-200' : 'border-slate-200 opacity-60'}`}
            >
              <div className="flex items-center gap-4 min-w-0">
                <img src={p.img} alt={p.nombre} className="w-14 h-14 object-cover rounded-md shrink-0" />
                <div className="min-w-0">
                  <p className="font-semibold text-slate-800 truncate">{p.nombre}</p>
                  <p className="text-xs text-slate-500">
                    {p.categoria} · stock {p.stock}
                    {!p.activo && <span className="ml-2 text-red-600 font-semibold">INACTIVO</span>}
                  </p>
                </div>
              </div>

              <div className="flex items-center gap-3">
                <p className="font-bold text-indigo-600 w-20 text-right">${p.precio.toFixed(2)}</p>
                <button
                  onClick={() => abrirEdicion(p)}
                  className="min-h-11 min-w-11 flex items-center justify-center text-slate-600 hover:bg-slate-100 rounded-lg transition"
                  title="Editar"
                  aria-label={`Editar ${p.nombre}`}
                >
                  <Pencil size={18} aria-hidden="true" />
                </button>
                <button
                  onClick={() => alternarEstado(p)}
                  className={`min-h-11 min-w-11 flex items-center justify-center rounded-lg transition ${p.activo ? 'text-red-500 hover:bg-red-50' : 'text-green-600 hover:bg-green-50'}`}
                  title={p.activo ? 'Desactivar' : 'Reactivar'}
                  aria-label={`${p.activo ? 'Desactivar' : 'Reactivar'} ${p.nombre}`}
                >
                  <Power size={18} aria-hidden="true" />
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* ===== PAGINACIÓN ===== */}
      {totalPaginas > 1 && (
        <div className="flex items-center justify-center gap-4 mt-6">
          <button
            onClick={() => setPagina((p) => Math.max(p - 1, 1))}
            disabled={pagina === 1}
            className="min-h-11 px-4 border border-slate-300 rounded-lg disabled:opacity-40 hover:bg-slate-50 transition"
          >
            Anterior
          </button>
          <span className="text-sm text-slate-600">
            Página {pagina} de {totalPaginas}
          </span>
          <button
            onClick={() => setPagina((p) => Math.min(p + 1, totalPaginas))}
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

export default AdminProductos;
