// src/services/http.ts
// Cliente HTTP único de la aplicación: centraliza la URL base, la cabecera
// Authorization y el formato de error del backend.

export const API_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:3000';

// Clave con la que se guarda la sesión en el navegador.
const CLAVE_SESION = 'multicatalogo_sesion';

export type Rol = 'admin' | 'cliente';

export interface Usuario {
  email: string;
  rol: Rol;
}

export interface Sesion {
  usuario: Usuario;
  token: string;
}

// Error tipado: conserva el código HTTP para que cada vista reaccione distinto.
export class ApiError extends Error {
  status: number;
  details?: unknown;

  constructor(status: number, message: string, details?: unknown) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.details = details;
  }
}

// ===== Helpers de sesión =====

export const leerSesion = (): Sesion | null => {
  try {
    const guardada = localStorage.getItem(CLAVE_SESION);
    return guardada ? (JSON.parse(guardada) as Sesion) : null;
  } catch {
    return null;
  }
};

export const guardarSesion = (sesion: Sesion) => {
  try {
    localStorage.setItem(CLAVE_SESION, JSON.stringify(sesion));
  } catch {
    // Si el navegador bloquea el almacenamiento, la sesión dura lo que la pestaña.
  }
};

export const borrarSesion = () => {
  try {
    localStorage.removeItem(CLAVE_SESION);
  } catch {
    // Sin almacenamiento no hay nada que borrar.
  }
};

// Estructura de error que devuelve el backend.
interface RespuestaError {
  status?: number;
  message?: string;
  details?: unknown;
}

// apiFetch adjunta el token automáticamente y convierte los errores en ApiError.
export const apiFetch = async <T>(
  ruta: string,
  opciones: RequestInit = {},
  mensajePorDefecto = 'Error al comunicarse con el servidor'
): Promise<T> => {
  const sesion = leerSesion();

  const cabeceras = new Headers(opciones.headers);
  if (opciones.body && !cabeceras.has('Content-Type')) {
    cabeceras.set('Content-Type', 'application/json');
  }
  if (sesion?.token) {
    cabeceras.set('Authorization', `Bearer ${sesion.token}`);
  }

  const response = await fetch(`${API_URL}${ruta}`, { ...opciones, headers: cabeceras });

  if (!response.ok) {
    const datos = (await response.json().catch(() => null)) as RespuestaError | null;

    // El token expiró o es inválido: la sesión guardada ya no sirve.
    if (response.status === 401) {
      borrarSesion();
    }

    throw new ApiError(response.status, datos?.message ?? mensajePorDefecto, datos?.details);
  }

  // 204 y similares no traen cuerpo.
  if (response.status === 204) {
    return undefined as T;
  }

  return (await response.json()) as T;
};
