# Lista de Verificación de Requerimientos (REQUIREMENTS_CHECKLIST)

## Estado General
- **Hito actual:** Back-office MultiCatálogo Completo (Unidad 1 + Unidad 2).
- **Última tarea completada:** Integración y verificación completa de Fases 0 a 7 del [PLAN_IMPLEMENTACION_U2.md](file:///c:/Users/Gino/Proyectos/mi-proyecto-spa/docs/PLAN_IMPLEMENTACION_U2.md): Registro con código de referido (`/registro`), CRUD Admin de Productos (`/admin/productos`), Gestión de Pedidos (`/admin/pedidos`), Historial de Compras (`/mis-pedidos`), Mi Red MLM con enlace copiable y Ledger auditable (`/mi-red`), Dashboard con reportes agregados SQL nativos (`/api/reportes/resumen`) y cliente API con Bearer Token automático.
- **Siguiente tarea pendiente:** Despliegue, empaquetado de producción o inicio de módulos de Unidad 3 (SaaS multitenant / subdominios).

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
- [x] Enlace copiable de referido propio con código (`/registro?ref=CODIGO`).
- [x] Ledger auditable de comisiones devengadas con estados (`pendiente`, `pagada`, `revertida`).
- [x] Integración con endpoints `GET /api/red/mia` y `GET /api/comisiones/mias`.

### Hito F7: Dashboard con KPIs Dinámicos
- [x] Indicadores en tiempo real: total ventas acumuladas, número de pedidos, comisiones devengadas y ticket promedio.
- [x] Integración con endpoint de agregación SQL `GET /api/reportes/resumen`.

### Hito F8: Autenticación, Registro y Navegación por Rol
- [x] Login con credenciales para roles `admin` y `cliente`.
- [x] Registro público `/registro` con validación de contraseñas y código de referido opcional.
- [x] Historial de compras para clientes en `/mis-pedidos`.
- [x] Panel administrativo de productos en `/admin/productos` (crear, editar, stock, desactivar).
- [x] Panel administrativo de pedidos en `/admin/pedidos` (estados y trazabilidad).
- [x] Rutas protegidas (`ProtectedRoute` y `AdminRoute`) restringiendo vistas según perfil de rol.
- [x] Navbar y Sidebar adaptativos mostrando solo opciones autorizadas para cada rol.

---

## 2. Unidad 2 — Backend (Go + Fiber + PostgreSQL + Redis)

### Fase 0: Migración DDL Backoffice
- [x] Script `database/004_backoffice.sql` idempotente para tablas `comisiones`, `liquidaciones`, campos `activo`, `stock`, `codigo` de referidos y estados de pedidos.

### Fase 1: Registro y Vinculación Referido Real
- [x] `POST /api/register` con bcrypt, asignación de rol `cliente` y creación de nodo de red en transacción ACID.
- [x] Control de duplicados (409) y código de referido inválido (400).
- [x] Rechazo de cuentas desactivadas con HTTP 403 Forbidden en `/api/login`.

### Fase 2: CRUD de Productos (Admin)
- [x] `POST /api/admin/productos`, `PUT /api/admin/productos/:id`, `DELETE /api/admin/productos/:id` (borrado lógico).
- [x] Listado admin con paginación y búsqueda server-side `?q=` ILIKE.
- [x] Invalidación de caché Redis `productos:all` en cada mutación.

### Fase 3: Gestión y Ciclo de Vida de Pedidos
- [x] `GET /api/pedidos/mios` filtrado por usuario autenticado.
- [x] `GET /api/pedidos` (admin) con paginación y filtro `?estado=`.
- [x] `PATCH /api/pedidos/:id/estado` con validación de transiciones de estado.
- [x] Cancelación atómica: restauración de stock en `productos` y reversión de comisiones en el ledger.

### Fase 4: Ledger de Comisiones y Red Propia
- [x] Inserción de comisiones por niveles (10%, 5%, 2%) en la transacción de `CrearPedido`.
- [x] Subárbol propio en `GET /api/red/mia` partiendo del `referido_id` del usuario.
- [x] Historial auditable en `GET /api/comisiones/mias`.

### Fase 5: Liquidaciones a Distribuidores
- [x] `POST /api/liquidaciones` con bloqueo transaccional `FOR UPDATE`, suma de pendientes y actualización a `pagada`.
- [x] Consultas de historial en `GET /api/liquidaciones` y `GET /api/liquidaciones/mias`.

### Fase 6: Reportes y Métricas en Tiempo Real
- [x] `GET /api/reportes/resumen` con agregaciones nativas (`SUM`, `COUNT`, `AVG`).
- [x] `GET /api/reportes/top-productos`, `GET /api/reportes/top-referidos`, `GET /api/reportes/comisiones-pendientes`.

### Fase 7: Seguridad y Dureza Operacional
- [x] Middleware `RequireAuth` y `RequireAdmin` en todas las rutas operativas.
- [x] Formato estándar de respuestas y manejo de errores con `APIError`.
- [x] Suite de pruebas unitarias e integración en verde.
