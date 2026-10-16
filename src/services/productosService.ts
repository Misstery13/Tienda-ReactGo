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
