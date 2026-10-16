// src/context/cart.ts
// Contexto, tipos y hook del carrito, separados del Provider.
import { createContext, useContext } from 'react';
import type { Producto } from '../data/productos';

// Elemento del carrito: un producto + la cantidad seleccionada
export interface CartItem extends Producto {
  cantidad: number;
}

export interface CartContextType {
  cart: CartItem[];
  addToCart: (producto: Producto) => void;
  removeFromCart: (id: number) => void;
  incrementQuantity: (id: number) => void;
  decrementQuantity: (id: number) => void;
  clearCart: () => void;
  totalItems: number;
  totalPrice: number;
}

// Prefijo de la clave de persistencia en localStorage.
// Se completa con el correo del usuario para que CADA CUENTA tenga su propio carrito.
export const CART_STORAGE_PREFIX = "multicatalogo_carrito_";

export const CartContext = createContext<CartContextType | undefined>(undefined);

export const useCart = () => {
  const context = useContext(CartContext);
  if (!context) {
    throw new Error("useCart debe ser usado dentro de un CartProvider");
  }
  return context;
};
