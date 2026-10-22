# Mapa Arquitectónico del Proyecto (PROJECT_MAP)

> **Stack Principal:**
> - **Frontend:** React 19 + TypeScript + Vite + Tailwind CSS + React Router v7 (`http://localhost:5173`)
> - **Backend:** Go 1.25 + Fiber v2 + PostgreSQL (pgx/v5) + Redis (go-redis/v9) (`http://localhost:3000`)
> - **Infraestructura:** Docker Compose (PostgreSQL puerto `5434`, Redis puerto `6379`)

---

## 1. Backend (`multicatalogo-backend/`)
Core transaccional SaaS, arquitectura en capas, middleware de seguridad e inyección de dependencias (DI).

- `main.go`: Punto de entrada, conexión a DB/Redis, inyección de dependencias y arranque del servidor Fiber (:3000).
- `docker-compose.yml`: Servicios de PostgreSQL (`5434`) y Redis (`6379`).
- `config/`:
  - `database.go`: Pool de conexiones PostgreSQL (`pgxpool`) y variables de entorno.
  - `redis.go`: Cliente Redis oficial y helpers de conexión/ping.
  - `cache.go`: Interfaz `CacheService` desacoplada (Get, Set, Delete).
- `models/models.go`: DTOs y entidades (`Usuario`, `RegisterRequest`, `Producto`, `Referido`, `CheckoutRequest`, `PedidoConfirmado`, `Pedido`, `Comision`, `Liquidacion`, `ReporteResumen`, `APIError`).
- `middleware/`:
  - `auth.go`: `RequireAuth` (autenticación obligatoria) y `RequireAdmin` (control de acceso por roles).
- `repository/repository.go`: Interfaz `DBRepository` e implementación PostgreSQL (consultas SQL nativas, transacciones ACID `CrearPedido`, `RegistrarUsuario`, comisiones, liquidaciones con bloqueo `FOR UPDATE` y reportes agregados).
- `controllers/`:
  - `authController.go`: Autenticación con JWT y control de cuenta activa (`403 Forbidden` si está inactiva).
  - `usuarioController.go`: Registro con código de referido opcional y vinculación a la red multinivel (`POST /api/register`).
  - `prodController.go`: Catálogo público y detalle de producto con estrategia Cache-Aside.
  - `productoAdminController.go`: CRUD de productos para admin (POST/PUT/DELETE-lógico, búsqueda `?q=` ILIKE e invalidación de caché).
  - `pedidoController.go`: Consulta de pedidos propios (`/api/pedidos/mios`), listado admin y transiciones de estados con rollback de stock/comisiones.
  - `redController.go`: Árbol jerárquico multinivel recursivo con caché Redis.
  - `comisionController.go`: Subárbol propio (`/api/red/mia`), ledger auditable (`/api/comisiones/mias`) y liquidaciones.
  - `reporteController.go`: Métricas agregadas en tiempo real (resumen, top-productos, top-referidos, comisiones-pendientes).
  - `checkoutController.go`: Transacción ACID de compras y Goroutine asíncrona de comisiones.
  - `healthController.go`: Verificación de salud del backend (`/api/health`).
- `routes/`:
  - `routes.go`: Definición de endpoints HTTP y cableado de dependencias (`Dependencies`).
  - `routes_test.go`: Suite de pruebas unitarias y de integración (13 tests HTTP).
- `database/`:
  - `init.sql`: Script DDL base con esquemas relacionales, llaves foráneas e índices.
  - `004_backoffice.sql`: Migración idempotente del backoffice (usuarios, comisiones, liquidaciones, códigos de referidos).

---

## 2. Frontend (`multicatalogo-frontend/`)
SPA interactiva para catálogo multinivel y back-office operacional completo.

- `src/main.tsx` & `src/App.tsx`: Inicialización de React Router, proveedores de contexto y definición de rutas públicas, protegidas y administrativas.
- `src/components/`:
  - `Storefront.tsx`: Vista inmersiva a pantalla completa (`/tienda`) con hero y secciones.
  - `Catalogo.tsx`: Catálogo interactivo con búsqueda en tiempo real y filtros por categoría.
  - `DetalleProducto.tsx`: Vista detallada con galería de imágenes / lightbox modal (`/producto/:id`).
  - `Carrito.tsx`: Carrito de compras con persistencia `localStorage` por usuario.
  - `Checkout.tsx`: Formulario de pago, validación y envío de orden a `/api/checkout`.
  - `Confirmacion.tsx`: Pantalla de éxito con resumen estructurado del pedido.
  - `Login.tsx`: Autenticación, redirección condicionada por rol y enlace a registro.
  - `Registro.tsx`: Formulario público de alta con prefill de código de referido (`?ref=`).
  - `MisPedidos.tsx`: Historial de compras del cliente con detalle de ítems y estado en tiempo real.
  - `AdminProductos.tsx`: Panel administrativo para crear, editar, buscar y desactivar productos.
  - `AdminPedidos.tsx`: Gestión y transición de estados de órdenes de compra.
  - `MiRed.tsx`: Visualización jerárquica de la red propia, link copiable de referido (`?ref=CODIGO`) y ledger de comisiones.
  - `Dashboard.tsx`: KPIs de negocio conectados con `/api/reportes/resumen` de la base de datos.
  - `PwaBanner.tsx`: Detección de conectividad offline y modal nativo de instalación PWA (`beforeinstallprompt`).
  - `Layout.tsx`, `Navbar.tsx`, `Sidebar.tsx`: Estructura responsive con navegación por perfil de rol (`admin`/`cliente`).
  - `ProtectedRoute.tsx`: Guardas de rutas por autenticación y roles.
- `src/context/`:
  - `AuthContext.tsx`: Estado global de autenticación, rol (`admin` / `cliente`) y token.
  - `CartContext.tsx`: Estado global de carrito persistente indexado por correo del usuario.
- `src/services/api.ts`: Cliente HTTP centralizado con token Bearer automático para todos los endpoints del backend.
- `src/data/`: Datos locales y fallbacks tipados para productos y red MLM.

---

## 3. Documentación y Especificaciones (`docs/`)
- `docs/Guia_Practica_Tema5_leccion.md`: Guía metodológica paso a paso de la Práctica 03 (Unidad 1, Tema 5).
- `docs/ANALISIS_FUNCIONAL_U2.md`: Análisis funcional del backoffice y requerimientos MLM.
- `docs/PLAN_IMPLEMENTACION_U2.md`: Plan de implementación y checklist técnico detallado de Fases 0 a 7.
