// src/context/auth.ts
// Contexto, tipos y hook de autenticación. Separados del Provider para que
// el archivo de componentes solo exporte componentes (regla de fast refresh).
import { createContext, useContext } from 'react';
import type { Rol, Usuario } from '../services/http';

export type { Rol, Usuario };

export interface AuthContextType {
  isAuthenticated: boolean;
  user: Usuario | null;
  token: string | null;
  login: (usuario: Usuario, token: string) => void;
  logout: () => void;
}

export const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth debe ser usado dentro de un AuthProvider");
  }
  return context;
};
