// src/services/redService.ts
// Red multinivel de referidos servida por la API (GET /api/red).
// La identidad sale del token: el backend ignora cualquier email enviado por el cliente.

import { apiFetch } from "./http";
import type { Referido } from "../data/red";

// GET /api/red
export const getRed = (): Promise<Referido> =>
  apiFetch<Referido>("/api/red", {}, "No se pudo obtener la red de referidos");
