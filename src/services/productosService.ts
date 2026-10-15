// src/services/productosService.ts
// Capa de servicios del catálogo: consume la API RESTful en Go.
// Los datos provienen de PostgreSQL a través de GET /api/productos.

import { API_URL, leerError } from "./api";
import type { Producto } from "../data/productos";

// GET /api/productos
export const getProductos = async (): Promise<Producto[]> => {
  const response = await fetch(`${API_URL}/api/productos`);

  if (!response.ok) {
    throw new Error(await leerError(response, "No se pudo obtener el catálogo"));
  }

  return (await response.json()) as Producto[];
};

// GET /api/productos/:id
export const getProductoById = async (id: number): Promise<Producto | undefined> => {
  const response = await fetch(`${API_URL}/api/productos/${id}`);

  // El backend responde 404 cuando el id es válido pero el producto no existe.
  if (response.status === 404) {
    return undefined;
  }

  if (!response.ok) {
    throw new Error(await leerError(response, "No se pudo obtener el producto"));
  }

  return (await response.json()) as Producto;
};

// Las categorías de los filtros salen de los productos recibidos, sin lista fija.
export const obtenerCategorias = (productos: Producto[]): string[] =>
  [...new Set(productos.map((p) => p.categoria))].filter(Boolean).sort();
