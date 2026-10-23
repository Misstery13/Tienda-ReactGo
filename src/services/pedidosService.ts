// src/services/pedidosService.ts
// Consulta y gestión de pedidos.

import { apiFetch } from "./http";

export type EstadoPedido =
  | "pendiente"
  | "preparando"
  | "enviado"
  | "entregado"
  | "cancelado";

export interface ResumenPedido {
  id: number;
  numero: string;
  cliente: string;
  email: string;
  total: number;
  estado: EstadoPedido;
  fecha: string;
  articulos: number;
}

export interface DetallePedido {
  id: number;
  numero: string;
  cliente: { nombre: string; email: string; ciudad: string };
  items: { producto_id: number; nombre: string; cantidad: number; precio: number; subtotal: number }[];
  total: number;
  fecha: string;
  estado: EstadoPedido;
}

export interface PaginaPedidos {
  datos: ResumenPedido[];
  total: number;
  pagina: number;
  limite: number;
}

// A qué estados puede avanzar un pedido. Misma tabla que domain/estados.go.
export const SIGUIENTES: Record<EstadoPedido, EstadoPedido[]> = {
  pendiente: ["preparando", "cancelado"],
  preparando: ["enviado", "cancelado"],
  enviado: ["entregado", "cancelado"],
  entregado: [],
  cancelado: [],
};

// Color de la etiqueta según el estado.
export const COLOR_ESTADO: Record<EstadoPedido, string> = {
  pendiente: "bg-amber-100 text-amber-700",
  preparando: "bg-indigo-100 text-indigo-700",
  enviado: "bg-teal-100 text-teal-700",
  entregado: "bg-green-100 text-green-700",
  cancelado: "bg-red-100 text-red-700",
};

// GET /api/pedidos/mios
export const getMisPedidos = (pagina = 1, limite = 10): Promise<PaginaPedidos> =>
  apiFetch<PaginaPedidos>(
    `/api/pedidos/mios?page=${pagina}&limit=${limite}`,
    {},
    "No se pudieron obtener tus pedidos"
  );

// GET /api/pedidos — solo admin, con filtro opcional por estado.
export const getPedidos = (estado = "", pagina = 1, limite = 10): Promise<PaginaPedidos> =>
  apiFetch<PaginaPedidos>(
    `/api/pedidos?estado=${estado}&page=${pagina}&limit=${limite}`,
    {},
    "No se pudieron obtener los pedidos"
  );

// GET /api/pedidos/:id
export const getPedido = (id: number): Promise<DetallePedido> =>
  apiFetch<DetallePedido>(`/api/pedidos/${id}`, {}, "No se pudo obtener el pedido");

// PATCH /api/pedidos/:id/estado — solo admin.
export const cambiarEstado = (id: number, estado: EstadoPedido): Promise<DetallePedido> =>
  apiFetch<DetallePedido>(
    `/api/pedidos/${id}/estado`,
    { method: "PATCH", body: JSON.stringify({ estado }) },
    "No se pudo cambiar el estado"
  );
