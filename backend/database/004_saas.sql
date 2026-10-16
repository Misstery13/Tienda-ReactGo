-- Migración 004: capa SaaS/multitenant (Unidad 3).
-- Modelo de multitenancy por discriminador: cada fila lleva su tenant_id.

CREATE TABLE IF NOT EXISTS planes (
    id             SERIAL PRIMARY KEY,
    nombre         VARCHAR(60)    NOT NULL UNIQUE,
    precio_mensual NUMERIC(10, 2) NOT NULL CHECK (precio_mensual >= 0),
    max_usuarios   INT            NOT NULL CHECK (max_usuarios > 0),
    max_productos  INT            NOT NULL CHECK (max_productos > 0)
);

CREATE TABLE IF NOT EXISTS tenants (
    id        SERIAL PRIMARY KEY,
    slug      VARCHAR(40)  NOT NULL UNIQUE,
    nombre    VARCHAR(150) NOT NULL,
    activo    BOOLEAN      NOT NULL DEFAULT TRUE,
    creado_en TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS suscripciones (
    id        SERIAL PRIMARY KEY,
    tenant_id INT         NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
    plan_id   INT         NOT NULL REFERENCES planes (id),
    estado    VARCHAR(20) NOT NULL DEFAULT 'activa'
                  CHECK (estado IN ('activa', 'suspendida', 'cancelada')),
    inicia    DATE        NOT NULL DEFAULT CURRENT_DATE,
    renueva   DATE        NOT NULL DEFAULT (CURRENT_DATE + INTERVAL '1 month')
);

-- Un tenant no puede tener dos suscripciones activas a la vez.
CREATE UNIQUE INDEX IF NOT EXISTS idx_suscripcion_activa
    ON suscripciones (tenant_id) WHERE estado = 'activa';

-- Las tablas del núcleo se vuelven multitenant.
ALTER TABLE usuarios  ADD COLUMN IF NOT EXISTS tenant_id INT REFERENCES tenants (id);
ALTER TABLE productos ADD COLUMN IF NOT EXISTS tenant_id INT REFERENCES tenants (id);
ALTER TABLE referidos ADD COLUMN IF NOT EXISTS tenant_id INT REFERENCES tenants (id);

CREATE INDEX IF NOT EXISTS idx_usuarios_tenant  ON usuarios  (tenant_id);
CREATE INDEX IF NOT EXISTS idx_productos_tenant ON productos (tenant_id);
CREATE INDEX IF NOT EXISTS idx_referidos_tenant ON referidos (tenant_id);

-- ============ DATOS SEMILLA ============

INSERT INTO planes (id, nombre, precio_mensual, max_usuarios, max_productos) VALUES
    (1, 'Básico',       0.00,   5,   20),
    (2, 'Profesional', 29.90,  50,  500),
    (3, 'Empresarial', 99.90, 500, 5000)
ON CONFLICT (id) DO NOTHING;

-- Tenant por defecto: los datos que ya existían pertenecen a este inquilino.
INSERT INTO tenants (id, slug, nombre) VALUES
    (1, 'upse', 'MultiCatálogo UPSE')
ON CONFLICT (id) DO NOTHING;

INSERT INTO suscripciones (tenant_id, plan_id, estado) VALUES (1, 2, 'activa')
ON CONFLICT DO NOTHING;

UPDATE usuarios  SET tenant_id = 1 WHERE tenant_id IS NULL;
UPDATE productos SET tenant_id = 1 WHERE tenant_id IS NULL;
UPDATE referidos SET tenant_id = 1 WHERE tenant_id IS NULL;

SELECT setval('planes_id_seq',  (SELECT MAX(id) FROM planes));
SELECT setval('tenants_id_seq', (SELECT MAX(id) FROM tenants));
