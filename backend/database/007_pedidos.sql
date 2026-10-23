-- Migración 007: ciclo de vida del pedido (Unidad 2).
-- Hasta aquí un pedido nacía "pagado" de hecho y nadie podía consultarlo.

ALTER TABLE pedidos ADD COLUMN IF NOT EXISTS estado VARCHAR(20) NOT NULL DEFAULT 'pendiente';

-- Estados válidos del flujo. Se agrega como constraint con nombre para poder
-- volver a ejecutar la migración sin duplicarla.
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'pedidos_estado_valido') THEN
        ALTER TABLE pedidos ADD CONSTRAINT pedidos_estado_valido
            CHECK (estado IN ('pendiente', 'preparando', 'enviado', 'entregado', 'cancelado'));
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_pedidos_estado  ON pedidos (estado);
CREATE INDEX IF NOT EXISTS idx_pedidos_usuario ON pedidos (usuario_id);
