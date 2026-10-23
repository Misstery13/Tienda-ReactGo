// src/data/productos.ts
// Tipo del catálogo. Los datos ya no viven aquí: provienen de PostgreSQL
// a través de GET /api/productos (ver services/productosService.ts).

export interface Producto {
  id: number;
  nombre: string;
  descripcion: string;
  precio: number;
  categoria: string;
  img: string;        // Imagen principal (portada)
  galeria: string[];  // Imágenes adicionales para la galería/lightbox
  stock?: number;     // Existencias; el checkout las descuenta
  activo?: boolean;   // Falso oculta el producto del catálogo
}
