// src/services/api.ts
// Llamadas a la API RESTful (backend Go/Fiber).
// URL base del backend (se puede cambiar en .env)
export const API_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:3000';

export interface LoginResponse {
  token: string;
  email: string;
  rol: string;
}

// Formato estándar de error del backend: { status, message, details }
interface APIError {
  status?: number;
  message?: string;
  details?: unknown;
}

// Extrae el mensaje que envió el servidor; si no hay JSON válido, usa el de respaldo.
export const leerError = async (response: Response, porDefecto: string): Promise<string> => {
  const data = (await response.json().catch(() => null)) as APIError | null;
  return data?.message ?? porDefecto;
};

// POST /api/login
export const loginRequest = async (email: string, password: string): Promise<LoginResponse> => {
  const response = await fetch(`${API_URL}/api/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  });

  if (!response.ok) {
    throw new Error(await leerError(response, 'Credenciales incorrectas'));
  }

  return (await response.json()) as LoginResponse;
};
