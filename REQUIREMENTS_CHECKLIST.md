# Lista de Verificación de Requerimientos (REQUIREMENTS_CHECKLIST)

## Estado General
- **Hito actual:** Unidad 2 — Backend Transaccional, Caché y Persistencia Completados (Fase 1 a 4).
- **Última tarea completada:** Fase 4 Backend — Motor transaccional ACID en `/api/checkout`, Goroutines de comisiones multinivel, DI completa y 13/13 tests en verde.
- **Siguiente tarea pendiente:** Optimización de directivas de bajo consumo de tokens MCP y sincronización continua Frontend-Backend en desarrollo activo.

---

## 1. Unidad 1 — Frontend (React 19 + TypeScript + Tailwind)

### Hito F1: Storefront Inmersivo
- [x] Vista hero full-screen en `/tienda` con diseño inmersivo y llamada a la acción.
- [x] Secciones temáticas (destacados, categorías, beneficios del negocio).
- [x] Navegación fluida hacia el catálogo o detalle de producto.

### Hito F2 & F4: Detalle de Producto y Lightbox
- [x] Ruta dinámica `/producto/:id` con extracción de parámetros (`useParams`).
- [x] Visualización de galería de imágenes y lightbox modal interactivo accesible.
- [x] Selector de cantidad y adición directa al carrito de compras.

### Hito F3: Catálogo con Filtros y Búsqueda
- [x] Búsqueda por texto en tiempo real sobre el catálogo de productos.
- [x] Filtros por categoría y ordenamiento optimizados con `useMemo`.
- [x] Manejo de estados de carga (loading) y mensaje cuando no hay resultados.

### Hito F5: Flujo Transaccional de Carrito y Checkout
- [x] Carrito de compras con persistencia aislada por usuario (`localStorage` indexado por email).
- [x] Control de cantidades (incrementar, decrementar, eliminar ítem).
- [x] Formulario de checkout con validación de campos requeridos (nombre, email, dirección, método de pago).
- [x] Integración con endpoint `POST /api/checkout`.
- [x] Vista de confirmación del pedido (`/confirmacion`) con resumen completo de la orden.

### Hito F6: Red Multinivel (MLM) y Comisiones
- [x] Renderizado de estructura jerárquica de referidos por niveles (Nivel 1, Nivel 2, Nivel 3).
- [x] Cálculo derivado de ventas acumuladas y comisiones asignadas.
- [x] Integración con endpoint `GET /api/red`.

### Hito F7: Dashboard con KPIs Dinámicos
- [x] Indicadores en tiempo real: total ventas acumuladas, comisiones ganadas y miembros en red.
- [x] Sincronización con el estado de la red y compras efectuadas.

### Hito F8: Autenticación y Navegación por Rol
- [x] Login con credenciales para roles `admin` y `cliente`.
- [x] Rutas protegidas (`ProtectedRoute`) restringiendo vistas según perfil de rol.
- [x] Navbar y Sidebar adaptativos mostrando solo opciones autorizadas para cada rol.

---

## 2. Unidad 2 — Backend (Go + Fiber + PostgreSQL + Redis)

### Fase 1: Arquitectura Base y API Robusta
- [x] Separación de capas estricta: `routes`, `controllers`, `models`, `repository`.
- [x] Estandarización de errores con estructura uniforme `APIError` (`status`, `message`, `details`).
- [x] Validación estricta de payloads (campos vacíos, IDs inválidos).
- [x] Endpoint de verificación de estado `GET /api/health`.

### Fase 2: Persistencia Relacional con PostgreSQL
- [x] Contenedor Docker para PostgreSQL en puerto `5434`.
- [x] Esquema DDL relacional (`database/init.sql`) para tablas `usuarios`, `productos`, `referidos`.
- [x] Conexión mediante pool robusto `pgxpool`.
- [x] Migración de datos en memoria a consultas SQL relacionales nativas.

### Fase 3: Optimización con Caché Redis (Cache-Aside)
- [x] Contenedor Docker para Redis en puerto `6379`.
- [x] Patrón Cache-Aside en listado y detalle de catálogo (`productos:all`, `producto:<id>`).
- [x] Patrón Cache-Aside para árbol multinivel (`red:tree:<email>`).
- [x] Cabeceras de trazabilidad `X-Cache: HIT` y `X-Cache: MISS`.

### Fase 4: Motor Transaccional ACID y Concurrencia
- [x] Esquema relacional para `pedidos` y `pedido_items` con claves foráneas e integridad referencial.
- [x] Inyección de Dependencias (DI) completa en todos los controladores y rutas (`Dependencies`).
- [x] Endpoint `POST /api/checkout` con transacción ACID (`BEGIN` / `COMMIT` / `ROLLBACK`).
- [x] Concurrencia con Goroutines para cálculo y dispersión asíncrona de comisiones sin bloquear HTTP 201.
- [x] Invalidación atómica de caché en Redis (`productos:all`, `red:tree:*`) tras cada compra.
- [x] Suite de pruebas automatizadas en `routes/routes_test.go` (13/13 tests en verde).
