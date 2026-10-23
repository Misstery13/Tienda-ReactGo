-- Migración 006: administración del catálogo (Unidad 2).
-- Borrado lógico: un producto desactivado no se borra, porque
-- pedido_items histórico lo sigue referenciando.

ALTER TABLE productos ADD COLUMN IF NOT EXISTS activo BOOLEAN NOT NULL DEFAULT TRUE;

-- El catálogo público solo muestra activos: el índice acelera ese filtro.
CREATE INDEX IF NOT EXISTS idx_productos_activo ON productos (activo) WHERE activo;
