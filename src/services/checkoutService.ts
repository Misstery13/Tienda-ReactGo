// src/services/checkoutService.ts
// Envío del pedido al microservicio de checkout (POST /api/checkout).

import { API_URL, leerError } from "./api";
import type { CartItem } from "../context/CartContext";

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
export const crearPedido = async (
  datos: DatosEnvio,
  carrito: CartItem[]
): Promise<PedidoConfirmado> => {
  const response = await fetch(`${API_URL}/api/checkout`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      ...datos,
      items: carrito.map((item) => ({
        producto_id: item.id,
        cantidad: item.cantidad,
      })),
    }),
  });

  if (!response.ok) {
    throw new Error(await leerError(response, "No se pudo procesar el pedido"));
  }

  return (await response.json()) as PedidoConfirmado;
};
