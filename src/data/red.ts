// src/data/red.ts
// Tipo del árbol multinivel y funciones puras de cálculo.
// El árbol ya no vive aquí: proviene de PostgreSQL a través de GET /api/red
// (ver services/redService.ts). Estas funciones operan sobre el árbol recibido.

export interface Referido {
  id: number;
  nombre: string;
  nivel: number;       // relativo a la raíz del árbol recibido
  ventas: number;      // ventas acumuladas del nodo
  codigo?: string;     // código de invitación
  comisiones?: number; // acumulado del ledger
  hijos?: Referido[];  // referidos de un nivel más abajo
}

// Porcentaje de comisión que recibe el usuario por las ventas de cada nivel
export const TASA_COMISION: Record<number, number> = {
  1: 0.10, // 10 % de las ventas de nivel 1
  2: 0.05, // 5 % de las ventas de nivel 2
  3: 0.02, // 2 % de las ventas de nivel 3
};

// Total de referidos de la red (sin contar al usuario raíz)
export const contarRed = (raiz: Referido): number => {
  const directos = raiz.hijos ?? [];
  return directos.reduce((total, hijo) => total + 1 + contarRed(hijo), 0);
};

// Suma de ventas de toda la red (excluye las ventas propias del usuario raíz)
export const sumarVentasRed = (raiz: Referido): number => {
  const directos = raiz.hijos ?? [];
  return directos.reduce((total, hijo) => total + hijo.ventas + sumarVentasRed(hijo), 0);
};

// Comisión que recibe el usuario por las ventas de cada referido (ventas * tasa del nivel)
export const comisionDeReferido = (ref: Referido): number =>
  ref.ventas * (TASA_COMISION[ref.nivel] ?? 0);

// Comisiones totales del mes acumuladas sobre toda la red
export const sumarComisiones = (raiz: Referido): number => {
  const directos = raiz.hijos ?? [];
  return directos.reduce(
    (total, hijo) => total + comisionDeReferido(hijo) + sumarComisiones(hijo),
    0
  );
};

// Nivel alcanzado por el usuario según la cantidad de referidos directos
export const nivelAlcanzado = (referidosDirectos: number): string => {
  if (referidosDirectos >= 6) return "Diamante";
  if (referidosDirectos >= 4) return "Oro";
  if (referidosDirectos >= 2) return "Plata";
  return "Bronce";
};
