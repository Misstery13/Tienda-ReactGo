-- Migración 005: registro de clientes y crecimiento real de la red (Unidad 2).
-- Hasta aquí el vínculo usuario↔referido funcionaba por coincidencia de ids.
-- Esta migración lo convierte en una clave foránea de verdad.

ALTER TABLE usuarios ADD COLUMN IF NOT EXISTS nombre      VARCHAR(150) NOT NULL DEFAULT '';
ALTER TABLE usuarios ADD COLUMN IF NOT EXISTS activo      BOOLEAN      NOT NULL DEFAULT TRUE;
ALTER TABLE usuarios ADD COLUMN IF NOT EXISTS creado_en   TIMESTAMPTZ  NOT NULL DEFAULT now();
ALTER TABLE usuarios ADD COLUMN IF NOT EXISTS referido_id INT REFERENCES referidos (id);

-- Código de invitación: es lo que viaja en el enlace /registro?ref=CODIGO.
ALTER TABLE referidos ADD COLUMN IF NOT EXISTS codigo VARCHAR(20);
ALTER TABLE referidos ALTER COLUMN codigo TYPE VARCHAR(20);

-- Se genera un código legible y corto para los nodos que aún no lo tienen:
-- las 8 primeras letras del nombre, sin tildes, más el id.
UPDATE referidos
SET codigo = upper(
        left(regexp_replace(translate(nombre, 'áéíóúÁÉÍÓÚñÑ ', 'aeiouAEIOUnN-'), '[^a-zA-Z0-9]', '', 'g'), 8)
    ) || '-' || lpad(id::text, 3, '0')
WHERE codigo IS NULL;

ALTER TABLE referidos ALTER COLUMN codigo SET NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_referidos_codigo ON referidos (codigo);

-- Se cierra el vínculo de los usuarios semilla con su nodo del árbol.
UPDATE usuarios u SET referido_id = r.id
FROM referidos r
WHERE r.usuario_id = u.id AND u.referido_id IS NULL;

-- Nombres legibles para las cuentas semilla.
UPDATE usuarios SET nombre = 'Administrador' WHERE email = 'admin@upse.edu.ec'   AND nombre = '';
UPDATE usuarios SET nombre = 'Diana Paz'     WHERE email = 'cliente@upse.edu.ec' AND nombre = '';

CREATE INDEX IF NOT EXISTS idx_usuarios_referido ON usuarios (referido_id);
