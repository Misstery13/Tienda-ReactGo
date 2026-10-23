-- Migración 009: liquidaciones a distribuidores (Unidad 2).
-- Cierra el ciclo del dinero: el admin paga lo que el ledger debe.

CREATE TABLE IF NOT EXISTS liquidaciones (
    id          SERIAL PRIMARY KEY,
    referido_id INT            NOT NULL REFERENCES referidos (id),
    total       NUMERIC(12, 2) NOT NULL CHECK (total > 0),
    estado      VARCHAR(20)    NOT NULL DEFAULT 'pagada'
                    CHECK (estado IN ('pendiente', 'pagada', 'anulada')),
    creado_en   TIMESTAMPTZ    NOT NULL DEFAULT now(),
    pagada_en   TIMESTAMPTZ
);

-- Cada comisión pagada queda atada a la liquidación que la pagó.
ALTER TABLE comisiones ADD COLUMN IF NOT EXISTS liquidacion_id INT REFERENCES liquidaciones (id);

CREATE INDEX IF NOT EXISTS idx_liquidaciones_referido ON liquidaciones (referido_id);
CREATE INDEX IF NOT EXISTS idx_comisiones_liquidacion ON comisiones (liquidacion_id);
