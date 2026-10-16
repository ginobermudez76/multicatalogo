# Mapa Arquitectónico del Proyecto (PROJECT_MAP)

> **Stack Principal:**
> - **Frontend:** React 19 + TypeScript + Vite + Tailwind CSS + React Router v7 (`http://localhost:5173`)
> - **Backend:** Go 1.25 + Fiber v2 + PostgreSQL (pgx/v5) + Redis (go-redis/v9) (`http://localhost:3000`)
> - **Infraestructura:** Docker Compose (PostgreSQL puerto `5434`, Redis puerto `6379`)

---

## 1. Backend (`multicatalogo-backend/`)
Core transaccional SaaS, arquitectura en capas e inyección de dependencias (DI).

- `main.go`: Punto de entrada, conexión a DB/Redis, inyección de dependencias y arranque del servidor Fiber (:3000).
- `docker-compose.yml`: Servicios de PostgreSQL (`5434`) y Redis (`6379`).
- `config/`:
  - `database.go`: Pool de conexiones PostgreSQL (`pgxpool`) y variables de entorno.
  - `redis.go`: Cliente Redis oficial y helpers de conexión/ping.
  - `cache.go`: Interfaz `CacheService` desacoplada (Get, Set, Delete).
- `models/models.go`: DTOs y entidades (`Usuario`, `Producto`, `Referido`, `CheckoutRequest`, `PedidoConfirmado`, `APIError`).
- `repository/repository.go`: Interfaz `DBRepository` e implementación PostgreSQL (consultas SQL nativas, transacciones ACID `CrearPedido` y comisiones).
- `controllers/`:
  - `authController.go`: Autenticación con JWT simulado y roles (`admin`/`cliente`).
  - `prodController.go`: Catálogo y detalle de producto con estrategia Cache-Aside.
  - `redController.go`: Árbol jerárquico multinivel recursivo con caché Redis.
  - `checkoutController.go`: Transacción ACID de compras y Goroutine asíncrona de comisiones.
  - `healthController.go`: Verificación de salud del backend (`/api/health`).
- `routes/`:
  - `routes.go`: Definición de endpoints HTTP y cableado de dependencias (`Dependencies`).
  - `routes_test.go`: Suite de pruebas unitarias y de integración (13 tests HTTP).
- `database/init.sql`: Script DDL con esquemas relacionales, llaves foráneas e índices.

---

## 2. Frontend (`multicatalogo-frontend/`)
SPA interactiva para catálogo multinivel y comercio electrónico.

- `src/main.tsx` & `src/App.tsx`: Inicialización de React Router, proveedores de contexto y definición de rutas.
- `src/components/`:
  - `Storefront.tsx`: Vista inmersiva a pantalla completa (`/tienda`) con hero y secciones.
  - `Catalogo.tsx`: Catálogo interactivo con búsqueda en tiempo real y filtros por categoría.
  - `DetalleProducto.tsx`: Vista detallada con galería de imágenes / lightbox modal (`/producto/:id`).
  - `Carrito.tsx`: Carrito de compras con persistencia `localStorage` por usuario.
  - `Checkout.tsx`: Formulario de pago, validación y envío de orden a `/api/checkout`.
  - `Confirmacion.tsx`: Pantalla de éxito con resumen estructurado del pedido.
  - `MiRed.tsx`: Visualización jerárquica del árbol multinivel de afiliados y comisiones.
  - `Dashboard.tsx`: KPIs de ventas, comisiones y referidos calculados en tiempo real.
  - `Login.tsx`: Autenticación y redirección condicionada por rol (`admin` / `cliente`).
  - `Layout.tsx`, `Navbar.tsx`, `Sidebar.tsx`: Estructura responsive con navegación por perfil de rol.
  - `ProtectedRoute.tsx`: Guardas de rutas por autenticación y roles.
- `src/context/`:
  - `AuthContext.tsx`: Estado global de autenticación, rol (`admin` / `cliente`) y token.
  - `CartContext.tsx`: Estado global de carrito persistente indexado por correo del usuario.
- `src/services/api.ts`: Cliente HTTP para el consumo de endpoints del backend.
- `src/data/`: Datos locales de fallback para productos y red.

---

## 3. Documentación y Especificaciones (`docs/`)
- `docs/Guia_Practica_Tema5_leccion.md`: Guía metodológica paso a paso de la Práctica 03 (Unidad 1, Tema 5).
