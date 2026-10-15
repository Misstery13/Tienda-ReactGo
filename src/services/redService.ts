// src/services/redService.ts
// Red multinivel de referidos servida por la API (GET /api/red).

import { API_URL, leerError } from "./api";
import type { Referido } from "../data/red";

// GET /api/red?email=... — el email identifica de quién es la red.
export const getRed = async (email: string): Promise<Referido> => {
  const response = await fetch(`${API_URL}/api/red?email=${encodeURIComponent(email)}`);

  if (!response.ok) {
    throw new Error(await leerError(response, "No se pudo obtener la red de referidos"));
  }

  return (await response.json()) as Referido;
};
