// src/services/api.ts
// Autenticación contra la API RESTful (backend Go/Fiber).

import { apiFetch, type Rol } from './http';

export { API_URL, ApiError } from './http';

export interface LoginResponse {
  token: string;
  email: string;
  rol: Rol;
}

// POST /api/login
export const loginRequest = (email: string, password: string): Promise<LoginResponse> =>
  apiFetch<LoginResponse>(
    '/api/login',
    { method: 'POST', body: JSON.stringify({ email, password }) },
    'Credenciales incorrectas'
  );
