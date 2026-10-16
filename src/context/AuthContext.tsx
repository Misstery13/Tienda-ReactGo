// src/context/AuthContext.tsx
import { useState, type ReactNode } from 'react';
import { AuthContext, type Usuario } from './auth';
import { borrarSesion, guardarSesion, leerSesion } from '../services/http';

interface AuthProviderProps {
  children: ReactNode;
}

export const AuthProvider = ({ children }: AuthProviderProps) => {
  // La sesión se recupera de localStorage al montar: así sobrevive a un F5.
  const sesionGuardada = leerSesion();

  const [user, setUser] = useState<Usuario | null>(sesionGuardada?.usuario ?? null);
  const [token, setToken] = useState<string | null>(sesionGuardada?.token ?? null);

  const login = (usuario: Usuario, apiToken: string) => {
    setUser(usuario);
    setToken(apiToken);
    guardarSesion({ usuario, token: apiToken });
  };

  const logout = () => {
    setUser(null);
    setToken(null);
    borrarSesion();
  };

  return (
    <AuthContext.Provider
      value={{ isAuthenticated: token !== null, user, token, login, logout }}
    >
      {children}
    </AuthContext.Provider>
  );
};
