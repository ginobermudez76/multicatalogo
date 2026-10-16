-- Migración Back-office MultiCatálogo (Fases 0-7)
-- Base de datos: multicatalogo
-- Aplicar: psql -h localhost -p 5434 -U postgres -d multicatalogo -f database/004_backoffice.sql

-- 1. ALTER usuarios: agregar nombre, activo, referido_id, creado_en
ALTER TABLE usuarios
ADD COLUMN IF NOT EXISTS nombre VARCHAR(255),
ADD COLUMN IF NOT EXISTS activo BOOLEAN DEFAULT true,
ADD COLUMN IF NOT EXISTS referido_id INTEGER REFERENCES referidos(id) ON DELETE SET NULL,
ADD COLUMN IF NOT EXISTS creado_en TIMESTAMP DEFAULT NOW();

-- 2. ALTER referidos: agregar codigo VARCHAR UNIQUE
ALTER TABLE referidos
ADD COLUMN IF NOT EXISTS codigo VARCHAR(50) UNIQUE;

-- Generar códigos para el seed existente
UPDATE referidos SET codigo = 'TU' WHERE id = 0 AND codigo IS NULL;
UPDATE referidos SET codigo = 'ANA' WHERE id = 1 AND codigo IS NULL;
UPDATE referidos SET codigo = 'LUIS' WHERE id = 2 AND codigo IS NULL;
UPDATE referidos SET codigo = 'MARTA' WHERE id = 3 AND codigo IS NULL;
UPDATE referidos SET codigo = 'CARLOS' WHERE id = 4 AND codigo IS NULL;
UPDATE referidos SET codigo = 'SOFIA' WHERE id = 5 AND codigo IS NULL;
UPDATE referidos SET codigo = 'MARCO' WHERE id = 6 AND codigo IS NULL;
UPDATE referidos SET codigo = 'DIANA' WHERE id = 7 AND codigo IS NULL;

-- 3. ALTER productos: agregar activo (borrado lógico)
ALTER TABLE productos
ADD COLUMN IF NOT EXISTS activo BOOLEAN DEFAULT true,
ADD COLUMN IF NOT EXISTS stock INT DEFAULT 0;

-- Inicializar stock para productos existentes
UPDATE productos SET stock = 50 WHERE stock = 0;

-- 4. ALTER pedidos: agregar estado, usuario_id
ALTER TABLE pedidos
ADD COLUMN IF NOT EXISTS estado VARCHAR(50) DEFAULT 'pendiente',
ADD COLUMN IF NOT EXISTS usuario_id INTEGER REFERENCES usuarios(id) ON DELETE SET NULL;

-- Enlazar el seed: usuarios.referido_id
UPDATE usuarios SET referido_id = 0 WHERE id = 1; -- admin -> Tú
UPDATE usuarios SET referido_id = 2 WHERE id = 2; -- cliente -> Luis Poveda

-- 5. Tabla comisiones (ledger de comisiones)
CREATE TABLE IF NOT EXISTS comisiones (
    id SERIAL PRIMARY KEY,
    pedido_id INT REFERENCES pedidos(id) ON DELETE CASCADE,
    referido_id INT REFERENCES referidos(id) ON DELETE CASCADE,
    nivel INT NOT NULL,
    porcentaje NUMERIC(5, 2) NOT NULL,
    monto NUMERIC(10, 2) NOT NULL,
    estado VARCHAR(50) DEFAULT 'pendiente',
    liquidacion_id INT,
    creado_en TIMESTAMP DEFAULT NOW()
);

-- Índices para comisiones
CREATE INDEX IF NOT EXISTS idx_comisiones_referido_pedido ON comisiones(referido_id, pedido_id);
CREATE INDEX IF NOT EXISTS idx_comisiones_estado ON comisiones(estado);
CREATE INDEX IF NOT EXISTS idx_comisiones_liquidacion ON comisiones(liquidacion_id);

-- 6. Tabla liquidaciones
CREATE TABLE IF NOT EXISTS liquidaciones (
    id SERIAL PRIMARY KEY,
    referido_id INT REFERENCES referidos(id) ON DELETE CASCADE,
    total NUMERIC(10, 2) NOT NULL,
    estado VARCHAR(50) DEFAULT 'pendiente',
    creado_en TIMESTAMP DEFAULT NOW(),
    pagada_en TIMESTAMP
);

-- Índices para liquidaciones
CREATE INDEX IF NOT EXISTS idx_liquidaciones_referido ON liquidaciones(referido_id);
CREATE INDEX IF NOT EXISTS idx_liquidaciones_estado ON liquidaciones(estado);

-- 7. Foreign key constraint para comisiones.liquidacion_id
ALTER TABLE comisiones
ADD CONSTRAINT fk_comisiones_liquidacion
FOREIGN KEY (liquidacion_id) REFERENCES liquidaciones(id) ON DELETE SET NULL;

-- 8. Índices adicionales para performance
CREATE INDEX IF NOT EXISTS idx_pedidos_usuario ON pedidos(usuario_id);
CREATE INDEX IF NOT EXISTS idx_pedidos_estado ON pedidos(estado);
CREATE INDEX IF NOT EXISTS idx_pedidos_fecha ON pedidos(fecha);
CREATE INDEX IF NOT EXISTS idx_usuarios_email ON usuarios(email);
CREATE INDEX IF NOT EXISTS idx_usuarios_referido ON usuarios(referido_id);
