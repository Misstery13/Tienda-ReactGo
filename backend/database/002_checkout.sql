-- Migración 002: pedidos, stock y comisiones (Unidad 2, Tema 4).
-- Docker ejecuta los scripts de esta carpeta en orden alfabético.

-- ============ FASE 4: PEDIDOS, STOCK Y COMISIONES ============

-- Stock simulado que el checkout descuenta dentro de la transacción.
ALTER TABLE productos  ADD COLUMN IF NOT EXISTS stock      INT            NOT NULL DEFAULT 100;
-- Un referido puede corresponder a un usuario que compra en la tienda.
ALTER TABLE referidos  ADD COLUMN IF NOT EXISTS usuario_id INT REFERENCES usuarios (id);
-- Comisiones acreditadas por la goroutine al cerrar cada pedido.
ALTER TABLE referidos  ADD COLUMN IF NOT EXISTS comisiones NUMERIC(12, 2) NOT NULL DEFAULT 0;

-- El admin es la raíz de la red; el cliente es un referido de nivel 3,
-- así una compra suya reparte comisiones a tres ancestros.
UPDATE referidos SET usuario_id = 1 WHERE nombre = 'Tú'        AND usuario_id IS NULL;
UPDATE referidos SET usuario_id = 2 WHERE nombre = 'Diana Paz' AND usuario_id IS NULL;

CREATE TABLE IF NOT EXISTS pedidos (
    id             SERIAL PRIMARY KEY,
    numero         VARCHAR(30)    NOT NULL UNIQUE,
    usuario_id     INT REFERENCES usuarios (id),
    cliente_nombre VARCHAR(150)   NOT NULL,
    cliente_email  VARCHAR(150)   NOT NULL,
    cliente_ciudad VARCHAR(100)   NOT NULL,
    direccion      TEXT           NOT NULL,
    telefono       VARCHAR(40)    NOT NULL,
    metodo_pago    VARCHAR(60)    NOT NULL,
    total          NUMERIC(12, 2) NOT NULL CHECK (total >= 0),
    creado_en      TIMESTAMPTZ    NOT NULL DEFAULT now()
);

-- producto_id es clave foránea: un id inexistente aborta toda la transacción.
CREATE TABLE IF NOT EXISTS pedido_items (
    id              SERIAL PRIMARY KEY,
    pedido_id       INT            NOT NULL REFERENCES pedidos (id) ON DELETE CASCADE,
    producto_id     INT            NOT NULL REFERENCES productos (id),
    cantidad        INT            NOT NULL CHECK (cantidad > 0),
    precio_unitario NUMERIC(10, 2) NOT NULL,
    subtotal        NUMERIC(12, 2) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_pedido_items_pedido ON pedido_items (pedido_id);
CREATE INDEX IF NOT EXISTS idx_referidos_usuario   ON referidos (usuario_id);
