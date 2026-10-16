-- Migración 003: seguridad (Unidad 2, Tema 4).
-- Convierte las contraseñas en hashes bcrypt y agrega el índice que faltaba.

-- pgcrypto aporta crypt() y gen_salt('bf'), compatibles con bcrypt de Go.
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Las contraseñas en texto plano se reemplazan por su hash.
-- El prefijo $2 identifica un hash bcrypt: así la migración es idempotente.
UPDATE usuarios
SET password = crypt(password, gen_salt('bf', 10))
WHERE password NOT LIKE '$2%';

-- El hash ocupa 60 caracteres; la columna debe admitirlos.
ALTER TABLE usuarios ALTER COLUMN password TYPE VARCHAR(100);

-- Índice para buscar los pedidos de un cliente.
CREATE INDEX IF NOT EXISTS idx_pedidos_cliente_email ON pedidos (cliente_email);
