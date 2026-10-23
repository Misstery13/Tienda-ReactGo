// src/services/usuariosService.ts
// Registro público de clientes y administración de cuentas.

import { apiFetch } from "./http";
import type { LoginResponse } from "./api";

export interface RegistroDatos {
  nombre: string;
  email: string;
  password: string;
  codigo_referido?: string;
}

export interface UsuarioAdmin {
  id: number;
  email: string;
  nombre: string;
  rol: "admin" | "cliente";
  activo: boolean;
  referido_id?: number;
  creado_en?: string;
}

export interface Pagina<T> {
  datos: T[];
  total: number;
  pagina: number;
  limite: number;
}

// POST /api/register — devuelve la sesión ya iniciada.
export const registrar = (datos: RegistroDatos): Promise<LoginResponse> =>
  apiFetch<LoginResponse>(
    "/api/register",
    { method: "POST", body: JSON.stringify(datos) },
    "No se pudo crear la cuenta"
  );

// GET /api/usuarios — solo administradores.
export const getUsuarios = (pagina = 1, limite = 20): Promise<Pagina<UsuarioAdmin>> =>
  apiFetch<Pagina<UsuarioAdmin>>(
    `/api/usuarios?page=${pagina}&limit=${limite}`,
    {},
    "No se pudo obtener la lista de usuarios"
  );

// PATCH /api/usuarios/:id — cambia rol o estado.
export const actualizarUsuario = (
  id: number,
  cambios: { rol?: "admin" | "cliente"; activo?: boolean }
): Promise<UsuarioAdmin> =>
  apiFetch<UsuarioAdmin>(
    `/api/usuarios/${id}`,
    { method: "PATCH", body: JSON.stringify(cambios) },
    "No se pudo actualizar el usuario"
  );
