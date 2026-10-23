// src/services/productosService.ts
// Capa de servicios del catálogo: consume la API RESTful en Go.
// Los datos provienen de PostgreSQL a través de GET /api/productos.

import { apiFetch, ApiError } from "./http";
import type { Producto } from "../data/productos";

// GET /api/productos
export const getProductos = (): Promise<Producto[]> =>
  apiFetch<Producto[]>("/api/productos", {}, "No se pudo obtener el catálogo");

// GET /api/productos/:id
export const getProductoById = async (id: number): Promise<Producto | undefined> => {
  try {
    return await apiFetch<Producto>(`/api/productos/${id}`, {}, "No se pudo obtener el producto");
  } catch (err) {
    // El backend responde 404 cuando el id es válido pero el producto no existe.
    if (err instanceof ApiError && err.status === 404) {
      return undefined;
    }
    throw err;
  }
};

// Las categorías de los filtros salen de los productos recibidos, sin lista fija.
export const obtenerCategorias = (productos: Producto[]): string[] =>
  [...new Set(productos.map((p) => p.categoria))].filter(Boolean).sort();

// ===== Administración del catálogo (solo admin) =====

export interface ProductoAdmin extends Producto {
  stock: number;
  activo: boolean;
}

export interface PaginaProductos {
  datos: ProductoAdmin[];
  total: number;
  pagina: number;
  limite: number;
}

export interface ProductoDatos {
  nombre?: string;
  descripcion?: string;
  precio?: number;
  categoria?: string;
  img?: string;
  galeria?: string[];
  stock?: number;
  activo?: boolean;
}

// GET /api/admin/productos — incluye inactivos y acepta búsqueda.
export const getProductosAdmin = (q = "", pagina = 1, limite = 20): Promise<PaginaProductos> =>
  apiFetch<PaginaProductos>(
    `/api/admin/productos?q=${encodeURIComponent(q)}&page=${pagina}&limit=${limite}`,
    {},
    "No se pudo obtener el catálogo"
  );

// POST /api/productos
export const crearProducto = (datos: ProductoDatos): Promise<ProductoAdmin> =>
  apiFetch<ProductoAdmin>(
    "/api/productos",
    { method: "POST", body: JSON.stringify(datos) },
    "No se pudo crear el producto"
  );

// PUT /api/productos/:id — edición parcial.
export const actualizarProducto = (id: number, datos: ProductoDatos): Promise<ProductoAdmin> =>
  apiFetch<ProductoAdmin>(
    `/api/productos/${id}`,
    { method: "PUT", body: JSON.stringify(datos) },
    "No se pudo actualizar el producto"
  );

// DELETE /api/productos/:id — borrado lógico.
export const desactivarProducto = (id: number): Promise<{ id: number; activo: boolean }> =>
  apiFetch<{ id: number; activo: boolean }>(
    `/api/productos/${id}`,
    { method: "DELETE" },
    "No se pudo desactivar el producto"
  );
