// src/services/checkoutService.ts
// Envío del pedido al microservicio de checkout (POST /api/checkout).

import { apiFetch } from "./http";
import type { CartItem } from "../context/cart";

export interface DatosEnvio {
  nombre: string;
  email: string;
  direccion: string;
  ciudad: string;
  telefono: string;
  metodoPago: string;
}

// Forma del pedido que devuelve el backend y muestra Confirmacion.tsx.
export interface PedidoConfirmado {
  numero: string;
  items: { nombre: string; cantidad: number; precio: number }[];
  total: number;
  cliente: { nombre: string; email: string; ciudad: string };
  fecha: string;
}

// POST /api/checkout — solo viajan id y cantidad: el precio lo pone el servidor.
export const crearPedido = (
  datos: DatosEnvio,
  carrito: CartItem[]
): Promise<PedidoConfirmado> =>
  apiFetch<PedidoConfirmado>(
    "/api/checkout",
    {
      method: "POST",
      body: JSON.stringify({
        ...datos,
        items: carrito.map((item) => ({
          producto_id: item.id,
          cantidad: item.cantidad,
        })),
      }),
    },
    "No se pudo procesar el pedido"
  );
