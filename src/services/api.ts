// src/services/api.ts
// Llamadas a la API RESTful (backend Go/Fiber).
// URL base del backend (se puede cambiar en .env)
export const API_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:3000';

export interface LoginResponse {
  token: string;
  email: string;
  rol: string;
}

interface ApiError {
  error: string;
}

// POST /api/login
export const loginRequest = async (email: string, password: string): Promise<LoginResponse> => {
  const response = await fetch(`${API_URL}/api/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  });

  if (!response.ok) {
    const data = (await response.json().catch(() => null)) as ApiError | null;
    throw new Error(data?.error ?? 'Credenciales incorrectas');
  }

  return (await response.json()) as LoginResponse;
};
