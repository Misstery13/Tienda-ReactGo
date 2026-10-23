// src/services/reportesService.ts
// Reportes de negocio y liquidaciones a distribuidores.

import { apiFetch } from "./http";

export interface ResumenVentas {
  total_vendido: number;
  pedidos: number;
  ticket_promedio: number;
  comisiones_devengadas: number;
  cancelados: number;
}

export interface FilaReporte {
  id: number;
  nombre: string;
  cantidad: number;
  total: number;
}

export interface DeudaDistribuidor {
  referido_id: number;
  nombre: string;
  codigo: string;
  pendiente: number;
  lineas: number;
}

export interface Liquidacion {
  id: number;
  referido_id: number;
  nombre: string;
  total: number;
  estado: string;
  comisiones: number;
  fecha: string;
}

// GET /api/reportes/resumen
export const getResumen = (desde = "", hasta = ""): Promise<ResumenVentas> =>
  apiFetch<ResumenVentas>(
    `/api/reportes/resumen?desde=${desde}&hasta=${hasta}`,
    {},
    "No se pudo obtener el resumen"
  );

// GET /api/reportes/top-productos
export const getTopProductos = (limite = 5): Promise<FilaReporte[]> =>
  apiFetch<FilaReporte[]>(`/api/reportes/top-productos?limit=${limite}`, {}, "No se pudo obtener el ranking");

// GET /api/reportes/top-referidos
export const getTopReferidos = (limite = 5): Promise<FilaReporte[]> =>
  apiFetch<FilaReporte[]>(`/api/reportes/top-referidos?limit=${limite}`, {}, "No se pudo obtener el ranking");

// GET /api/comisiones/deuda
export const getDeuda = (): Promise<DeudaDistribuidor[]> =>
  apiFetch<DeudaDistribuidor[]>("/api/comisiones/deuda", {}, "No se pudo obtener la deuda");

// GET /api/liquidaciones
export const getLiquidaciones = (): Promise<Liquidacion[]> =>
  apiFetch<Liquidacion[]>("/api/liquidaciones", {}, "No se pudieron obtener las liquidaciones");

// POST /api/liquidaciones
export const liquidar = (referidoID: number): Promise<Liquidacion> =>
  apiFetch<Liquidacion>(
    "/api/liquidaciones",
    { method: "POST", body: JSON.stringify({ referido_id: referidoID }) },
    "No se pudo liquidar"
  );
