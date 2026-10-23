// Declaramos el paquete routes, encargado de registrar los endpoints de la API con Inyección de Dependencias.
package routes

import (
	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/config"
	"multicatalogo-backend/controllers"
	"multicatalogo-backend/middleware"
	"multicatalogo-backend/repository"
)

// Dependencies agrupa las instancias de controladores inyectadas al arrancar el servidor.
type Dependencies struct {
	AuthCtrl           *controllers.AuthController
	ProdCtrl           *controllers.ProdController
	RedCtrl            *controllers.RedController
	CheckoutCtrl       *controllers.CheckoutController
	HealthCtrl         *controllers.HealthController
	UsuarioCtrl        *controllers.UsuarioController
	ProductoAdminCtrl  *controllers.ProductoAdminController
	PedidoCtrl         *controllers.PedidoController
	ComisionCtrl       *controllers.ComisionController
	ReporteCtrl        *controllers.ReporteController
}

// SetupRoutes registra todos los endpoints de la API en la aplicación Fiber.
func SetupRoutes(app *fiber.App, deps ...Dependencies) {
	var d Dependencies
	if len(deps) > 0 {
		d = deps[0]
	} else {
		// Instanciación automática si no se proporcionan dependencias explícitas
		db := config.GetDB()
		redisClient := config.GetRedis()
		repo := repository.NewPostgresRepository(db)
		cache := config.NewRedisCacheService(redisClient)

		d = Dependencies{
			AuthCtrl:           controllers.NewAuthController(repo),
			ProdCtrl:           controllers.NewProdController(repo, cache),
			RedCtrl:            controllers.NewRedController(repo, cache),
			CheckoutCtrl:       controllers.NewCheckoutController(repo, cache),
			HealthCtrl:         controllers.NewHealthController(),
			UsuarioCtrl:        controllers.NewUsuarioController(repo),
			ProductoAdminCtrl:  controllers.NewProductoAdminController(repo, cache),
			PedidoCtrl:         controllers.NewPedidoController(repo),
			ComisionCtrl:       controllers.NewComisionController(repo, cache),
			ReporteCtrl:        controllers.NewReporteController(repo),
		}
	}

	// 1. Endpoint de verificación de salud
	app.Get("/api/health", d.HealthCtrl.HealthCheck)

	// 2. Endpoints de autenticación
	app.Post("/api/login", d.AuthCtrl.Login)
	app.Post("/api/register", d.UsuarioCtrl.Register)

	// 3. Endpoints del catálogo de productos (Cache-Aside en Redis)
	app.Get("/api/productos", d.ProdCtrl.GetProductos)
	app.Get("/api/productos/:id", d.ProdCtrl.GetProductoPorID)

	// 4. Endpoint de red multinivel de referidos (Cache-Aside en Redis)
	app.Get("/api/red", d.RedCtrl.GetRed)

	// 5. Endpoint de checkout y compra (Transacciones ACID en PostgreSQL y Goroutines)
	app.Post("/api/checkout", d.CheckoutCtrl.Checkout)

	// 6. Endpoints de administración de productos (Admin) - Fase 2
	adminProductos := app.Group("/api/admin/productos", middleware.RequireAuth, middleware.RequireAdmin)
	adminProductos.Post("", d.ProductoAdminCtrl.CrearProducto)
	adminProductos.Put("/:id", d.ProductoAdminCtrl.ActualizarProducto)
	adminProductos.Delete("/:id", d.ProductoAdminCtrl.EliminarProducto)
	adminProductos.Get("", d.ProductoAdminCtrl.ListarProductosAdmin)

	// 7. Endpoints de pedidos - Fase 3
	app.Get("/api/pedidos/mios", middleware.RequireAuth, d.PedidoCtrl.MisPedidos)
	app.Get("/api/pedidos", middleware.RequireAuth, middleware.RequireAdmin, d.PedidoCtrl.ListarPedidosAdmin)
	app.Get("/api/pedidos/:id", middleware.RequireAuth, d.PedidoCtrl.ObtenerPedidoDetalle)
	app.Patch("/api/pedidos/:id/estado", middleware.RequireAuth, middleware.RequireAdmin, d.PedidoCtrl.ActualizarEstadoPedido)

	// 8. Endpoints de comisiones y red propia - Fase 4
	app.Get("/api/red/mia", middleware.RequireAuth, d.ComisionCtrl.MiRed)
	app.Get("/api/comisiones/mias", middleware.RequireAuth, d.ComisionCtrl.MisComisiones)

	// 9. Endpoints de liquidaciones - Fase 5
	app.Post("/api/liquidaciones", middleware.RequireAuth, middleware.RequireAdmin, d.ComisionCtrl.CrearLiquidacion)
	app.Get("/api/liquidaciones", middleware.RequireAuth, middleware.RequireAdmin, d.ComisionCtrl.ListarLiquidaciones)
	app.Get("/api/liquidaciones/deuda", middleware.RequireAuth, middleware.RequireAdmin, d.ComisionCtrl.DeudaDistribuidores)
	app.Get("/api/liquidaciones/mias", middleware.RequireAuth, d.ComisionCtrl.MisLiquidaciones)

	// 10. Endpoints de gestión de usuarios (Admin)
	adminUsuarios := app.Group("/api/usuarios", middleware.RequireAuth, middleware.RequireAdmin)
	adminUsuarios.Get("", d.UsuarioCtrl.ListarUsuarios)
	adminUsuarios.Patch("/:id", d.UsuarioCtrl.ActualizarUsuario)

	// 11. Endpoints de reportes - Fase 6
	app.Get("/api/reportes/resumen", middleware.RequireAuth, middleware.RequireAdmin, d.ReporteCtrl.Resumen)
	app.Get("/api/reportes/top-productos", middleware.RequireAuth, middleware.RequireAdmin, d.ReporteCtrl.TopProductos)
	app.Get("/api/reportes/top-referidos", middleware.RequireAuth, middleware.RequireAdmin, d.ReporteCtrl.TopReferidos)
	app.Get("/api/reportes/comisiones-pendientes", middleware.RequireAuth, middleware.RequireAdmin, d.ReporteCtrl.ComisionesPendientes)
}