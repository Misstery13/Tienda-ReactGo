-- Migración 008: ledger de comisiones (Unidad 2).
-- Hasta aquí las comisiones solo inflaban un contador en referidos.
-- Un contador no es defendible ante auditoría: no dice quién cobró
-- qué porcentaje por cuál pedido. El ledger sí.

CREATE TABLE IF NOT EXISTS comisiones (
    id          SERIAL PRIMARY KEY,
    pedido_id   INT            NOT NULL REFERENCES pedidos (id) ON DELETE CASCADE,
    referido_id INT            NOT NULL REFERENCES referidos (id),
    nivel       INT            NOT NULL CHECK (nivel BETWEEN 1 AND 3),
    porcentaje  NUMERIC(5, 4)  NOT NULL CHECK (porcentaje > 0),
    monto       NUMERIC(12, 2) NOT NULL CHECK (monto >= 0),
    estado      VARCHAR(20)    NOT NULL DEFAULT 'pendiente'
                    CHECK (estado IN ('pendiente', 'aprobada', 'pagada', 'revertida')),
    creado_en   TIMESTAMPTZ    NOT NULL DEFAULT now()
);

-- Un pedido no puede pagar dos veces al mismo referido por el mismo nivel.
CREATE UNIQUE INDEX IF NOT EXISTS idx_comision_unica
    ON comisiones (pedido_id, referido_id, nivel);

CREATE INDEX IF NOT EXISTS idx_comisiones_referido ON comisiones (referido_id);
CREATE INDEX IF NOT EXISTS idx_comisiones_estado   ON comisiones (estado);

-- referidos.ventas y referidos.comisiones quedan como contadores de lectura
-- rápida; la fuente de verdad del dinero pasa a ser esta tabla.
COMMENT ON COLUMN referidos.comisiones IS 'Cache de SUM(comisiones.monto) no revertidas';
COMMENT ON COLUMN referidos.ventas     IS 'Cache de ventas acumuladas del nodo';
