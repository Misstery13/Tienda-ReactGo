-- Esquema y datos semilla de MultiCatálogo.
-- Docker ejecuta este script la primera vez que se crea el volumen de PostgreSQL.

-- ============ TABLAS ============

CREATE TABLE IF NOT EXISTS usuarios (
    id       SERIAL PRIMARY KEY,
    email    VARCHAR(150) NOT NULL UNIQUE,
    password VARCHAR(100) NOT NULL, -- texto plano temporal; se migrará a hash
    rol      VARCHAR(20)  NOT NULL CHECK (rol IN ('admin', 'cliente'))
);

CREATE TABLE IF NOT EXISTS productos (
    id          SERIAL PRIMARY KEY,
    nombre      VARCHAR(150)   NOT NULL,
    descripcion TEXT           NOT NULL DEFAULT '',
    precio      NUMERIC(10, 2) NOT NULL CHECK (precio >= 0),
    categoria   VARCHAR(60)    NOT NULL DEFAULT '',
    img         TEXT           NOT NULL DEFAULT '',
    galeria     JSONB          NOT NULL DEFAULT '[]'::jsonb
);

-- parent_id apunta a la misma tabla: así se arma el árbol multinivel.
CREATE TABLE IF NOT EXISTS referidos (
    id        SERIAL PRIMARY KEY,
    nombre    VARCHAR(150)   NOT NULL,
    nivel     INT            NOT NULL CHECK (nivel >= 0),
    ventas    NUMERIC(12, 2) NOT NULL DEFAULT 0,
    parent_id INT REFERENCES referidos (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_referidos_parent ON referidos (parent_id);
CREATE INDEX IF NOT EXISTS idx_productos_categoria ON productos (categoria);

-- ============ DATOS SEMILLA ============

INSERT INTO usuarios (id, email, password, rol) VALUES
    (1, 'admin@upse.edu.ec',   '123456', 'admin'),
    (2, 'cliente@upse.edu.ec', '123456', 'cliente')
ON CONFLICT (id) DO NOTHING;

INSERT INTO productos (id, nombre, descripcion, precio, categoria, img, galeria) VALUES
    (1, 'Serum Revitalizante',
     'Serum concentrado con vitamina C y ácido hialurónico. Ilumina la piel, reduce manchas y aporta hidratación profunda desde la primera aplicación.',
     45.00, 'Serum', 'https://picsum.photos/seed/serum/600',
     '["https://picsum.photos/seed/serum1/600","https://picsum.photos/seed/serum2/600","https://picsum.photos/seed/serum3/600"]'::jsonb),
    (2, 'Crema Hidratante Pro',
     'Crema facial de textura ligera con niacinamida y manteca de karité. Hidratación de 24 horas y barrera cutánea fortalecida para todo tipo de piel.',
     32.50, 'Crema', 'https://picsum.photos/seed/crema/600',
     '["https://picsum.photos/seed/crema1/600","https://picsum.photos/seed/crema2/600","https://picsum.photos/seed/crema3/600"]'::jsonb),
    (3, 'Tónico Purificante',
     'Tónico sin alcohol con hamamelis y agua de rosas. Limpia poros, equilibra el pH y prepara la piel para el resto de la rutina diaria.',
     28.00, 'Tónico', 'https://picsum.photos/seed/tonico/600',
     '["https://picsum.photos/seed/tonico1/600","https://picsum.photos/seed/tonico2/600","https://picsum.photos/seed/tonico3/600"]'::jsonb),
    (4, 'Mascarilla Nocturna',
     'Mascarilla de noche con colágeno y vitamina E. Repara la piel mientras duermes y recupera la luminosidad al despertar.',
     50.00, 'Mascarilla', 'https://picsum.photos/seed/mascarilla/600',
     '["https://picsum.photos/seed/mascarilla1/600","https://picsum.photos/seed/mascarilla2/600","https://picsum.photos/seed/mascarilla3/600"]'::jsonb),
    (5, 'Serum Anti-Edad Retinol',
     'Serum de retinol encapsulado que suaviza líneas de expresión y mejora la firmeza. Uso nocturno recomendado.',
     58.00, 'Serum', 'https://picsum.photos/seed/retinol/600',
     '["https://picsum.photos/seed/retinol1/600","https://picsum.photos/seed/retinol2/600","https://picsum.photos/seed/retinol3/600"]'::jsonb),
    (6, 'Crema Contorno de Ojos',
     'Contorno de ojos con cafeína y péptidos. Reduce bolsas y ojeras, hidrata la zona más delicada del rostro.',
     26.00, 'Crema', 'https://picsum.photos/seed/ojos/600',
     '["https://picsum.photos/seed/ojos1/600","https://picsum.photos/seed/ojos2/600","https://picsum.photos/seed/ojos3/600"]'::jsonb),
    (7, 'Tónico Exfoliante AHA-BHA',
     'Exfoliación química suave con ácidos AHA y BHA. Renueva la textura de la piel y desobstruye los poros.',
     34.00, 'Tónico', 'https://picsum.photos/seed/exfoliante/600',
     '["https://picsum.photos/seed/exfoliante1/600","https://picsum.photos/seed/exfoliante2/600","https://picsum.photos/seed/exfoliante3/600"]'::jsonb),
    (8, 'Kit Rutina Completa',
     'Kit integral con serum, crema, tónico y mascarilla. La rutina perfecta para comenzar: limpieza, tratamiento e hidratación.',
     129.00, 'Kit', 'https://picsum.photos/seed/kit/600',
     '["https://picsum.photos/seed/kit1/600","https://picsum.photos/seed/kit2/600","https://picsum.photos/seed/kit3/600"]'::jsonb)
ON CONFLICT (id) DO NOTHING;

-- La raíz del árbol (nivel 0, sin padre) es el usuario dueño de la red.
INSERT INTO referidos (id, nombre, nivel, ventas, parent_id) VALUES
    (1, 'Tú',            0, 2400, NULL),
    (2, 'Ana García',    1, 1200, 1),
    (3, 'Luis Poveda',   1,  850, 1),
    (4, 'Marta Sánchez', 1,  430, 1),
    (5, 'Carlos Ruiz',   2,  500, 2),
    (6, 'Sofía León',    2,  430, 2),
    (7, 'Marco Díaz',    2,  380, 3),
    (8, 'Diana Paz',     3,  300, 5)
ON CONFLICT (id) DO NOTHING;

-- Las secuencias deben continuar después de los ids insertados a mano.
SELECT setval('usuarios_id_seq',  (SELECT MAX(id) FROM usuarios));
SELECT setval('productos_id_seq', (SELECT MAX(id) FROM productos));
SELECT setval('referidos_id_seq', (SELECT MAX(id) FROM referidos));
