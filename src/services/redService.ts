// src/services/redService.ts
// Red multinivel de referidos servida por la API (GET /api/red).
// La identidad sale del token: el backend ignora cualquier email enviado por el cliente.

import { apiFetch } from "./http";
import type { Referido } from "../domain/red";

// GET /api/red — árbol global, solo admin.
export const getRed = (): Promise<Referido> =>
  apiFetch<Referido>("/api/red", {}, "No se pudo obtener la red de referidos");

// GET /api/red/mia — subárbol del usuario del token.
export const getMiRed = (): Promise<Referido> =>
  apiFetch<Referido>("/api/red/mia", {}, "No se pudo obtener tu red");

export interface LineaComision {
  id: number;
  pedido: string;
  nivel: number;
  porcentaje: number;
  monto: number;
  estado: "pendiente" | "aprobada" | "pagada" | "revertida";
  fecha: string;
}

export interface ResumenComisiones {
  lineas: LineaComision[];
  pendiente: number;
  pagada: number;
  revertida: number;
}

// GET /api/comisiones/mias — ledger del usuario.
export const getMisComisiones = (): Promise<ResumenComisiones> =>
  apiFetch<ResumenComisiones>("/api/comisiones/mias", {}, "No se pudieron obtener tus comisiones");
