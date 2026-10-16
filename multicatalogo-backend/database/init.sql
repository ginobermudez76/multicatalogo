-- Script de inicialización de Base de Datos para MultiCatálogo (Fases 2, 3 y 4)
-- Base de datos: multicatalogo

-- 1. Tabla de Usuarios
CREATE TABLE IF NOT EXISTS usuarios (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    rol VARCHAR(50) NOT NULL
);

-- 2. Tabla de Productos
CREATE TABLE IF NOT EXISTS productos (
    id SERIAL PRIMARY KEY,
    nombre VARCHAR(255) NOT NULL,
    descripcion TEXT,
    precio NUMERIC(10, 2) NOT NULL,
    categoria VARCHAR(100),
    img TEXT NOT NULL,
    galeria JSONB DEFAULT '[]'::jsonb
);

-- 3. Tabla de Referidos (Estructura de red multinivel tipo árbol)
CREATE TABLE IF NOT EXISTS referidos (
    id INT PRIMARY KEY,
    nombre VARCHAR(255) NOT NULL,
    nivel INT NOT NULL,
    ventas NUMERIC(10, 2) NOT NULL DEFAULT 0,
    parent_id INT REFERENCES referidos(id) ON DELETE SET NULL
);

-- 4. Tabla de Pedidos (Fase 4 - Transacciones ACID)
CREATE TABLE IF NOT EXISTS pedidos (
    id SERIAL PRIMARY KEY,
    numero VARCHAR(50) UNIQUE NOT NULL,
    cliente_nombre VARCHAR(255) NOT NULL,
    cliente_email VARCHAR(255) NOT NULL,
    cliente_ciudad VARCHAR(255) NOT NULL,
    cliente_direccion VARCHAR(255),
    cliente_telefono VARCHAR(50),
    metodo_pago VARCHAR(100) NOT NULL,
    total NUMERIC(10, 2) NOT NULL,
    fecha TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 5. Tabla de Items de Pedido (Fase 4 - Transacciones ACID)
CREATE TABLE IF NOT EXISTS pedido_items (
    id SERIAL PRIMARY KEY,
    pedido_id INT REFERENCES pedidos(id) ON DELETE CASCADE,
    producto_id INT REFERENCES productos(id) ON DELETE SET NULL,
    nombre VARCHAR(255) NOT NULL,
    cantidad INT NOT NULL,
    precio NUMERIC(10, 2) NOT NULL,
    subtotal NUMERIC(10, 2) NOT NULL
);

-- Poblar tabla de Usuarios
INSERT INTO usuarios (id, email, password, rol) VALUES
(1, 'admin@upse.edu.ec', '123456', 'admin'),
(2, 'cliente@upse.edu.ec', '123456', 'cliente')
ON CONFLICT (id) DO NOTHING;
SELECT setval(pg_get_serial_sequence('usuarios', 'id'), coalesce(max(id), 1)) FROM usuarios;

-- Poblar tabla de Productos
INSERT INTO productos (id, nombre, descripcion, precio, categoria, img, galeria) VALUES
(1, 'Serum Revitalizante', 'Serum concentrado con vitamina C y ácido hialurónico. Ilumina la piel, reduce manchas y aporta hidratación profunda desde la primera aplicación.', 45.00, 'Serum', 'https://picsum.photos/seed/serum/600', '["https://picsum.photos/seed/serum1/600", "https://picsum.photos/seed/serum2/600", "https://picsum.photos/seed/serum3/600"]'::jsonb),
(2, 'Crema Hidratante Pro', 'Crema facial de textura ligera con niacinamida y manteca de karité. Hidratación de 24 horas y barrera cutánea fortalecida para todo tipo de piel.', 32.50, 'Crema', 'https://picsum.photos/seed/crema/600', '["https://picsum.photos/seed/crema1/600", "https://picsum.photos/seed/crema2/600", "https://picsum.photos/seed/crema3/600"]'::jsonb),
(3, 'Tónico Purificante', 'Tónico sin alcohol con hamamelis y agua de rosas. Limpia poros, equilibra el pH y prepara la piel para el resto de la rutina diaria.', 28.00, 'Tónico', 'https://picsum.photos/seed/tonico/600', '["https://picsum.photos/seed/tonico1/600", "https://picsum.photos/seed/tonico2/600", "https://picsum.photos/seed/tonico3/600"]'::jsonb),
(4, 'Mascarilla Nocturna', 'Mascarilla de noche con colágeno y vitamina E. Repara la piel mientras duermes y recupera la luminosidad al despertar.', 50.00, 'Mascarilla', 'https://picsum.photos/seed/mascarilla/600', '["https://picsum.photos/seed/mascarilla1/600", "https://picsum.photos/seed/mascarilla2/600", "https://picsum.photos/seed/mascarilla3/600"]'::jsonb),
(5, 'Serum Anti-Edad Retinol', 'Serum de retinol encapsulado que suaviza líneas de expresión y mejora la firmeza. Uso nocturno recomendado.', 58.00, 'Serum', 'https://picsum.photos/seed/retinol/600', '["https://picsum.photos/seed/retinol1/600", "https://picsum.photos/seed/retinol2/600", "https://picsum.photos/seed/retinol3/600"]'::jsonb),
(6, 'Crema Contorno de Ojos', 'Contorno de ojos con cafeína y péptidos. Reduce bolsas y ojeras, hidrata la zona más delicada del rostro.', 26.00, 'Crema', 'https://picsum.photos/seed/ojos/600', '["https://picsum.photos/seed/ojos1/600", "https://picsum.photos/seed/ojos2/600", "https://picsum.photos/seed/ojos3/600"]'::jsonb),
(7, 'Tónico Exfoliante AHA-BHA', 'Exfoliación química suave con ácidos AHA y BHA. Renueva la textura de la piel y desobstruye los poros.', 34.00, 'Tónico', 'https://picsum.photos/seed/exfoliante/600', '["https://picsum.photos/seed/exfoliante1/600", "https://picsum.photos/seed/exfoliante2/600", "https://picsum.photos/seed/exfoliante3/600"]'::jsonb),
(8, 'Kit Rutina Completa', 'Kit integral con serum, crema, tónico y mascarilla. La rutina perfecta para comenzar: limpieza, tratamiento e hidratación.', 129.00, 'Kit', 'https://picsum.photos/seed/kit/600', '["https://picsum.photos/seed/kit1/600", "https://picsum.photos/seed/kit2/600", "https://picsum.photos/seed/kit3/600"]'::jsonb)
ON CONFLICT (id) DO NOTHING;
SELECT setval(pg_get_serial_sequence('productos', 'id'), coalesce(max(id), 1)) FROM productos;

-- Poblar tabla de Referidos
INSERT INTO referidos (id, nombre, nivel, ventas, parent_id) VALUES
(0, 'Tú', 0, 2400.00, NULL),
(1, 'Ana García', 1, 1200.00, 0),
(2, 'Luis Poveda', 1, 850.00, 0),
(3, 'Marta Sánchez', 1, 430.00, 0),
(4, 'Carlos Ruiz', 2, 500.00, 1),
(5, 'Sofía León', 2, 430.00, 1),
(6, 'Marco Díaz', 2, 380.00, 2),
(7, 'Diana Paz', 3, 300.00, 4)
ON CONFLICT (id) DO NOTHING;
