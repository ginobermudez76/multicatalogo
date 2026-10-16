// Declaramos el paquete routes, encargado de registrar los endpoints de la API con Inyección de Dependencias.
package routes

import (
	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/config"
	"multicatalogo-backend/controllers"
	"multicatalogo-backend/repository"
)

// Dependencies agrupa las instancias de controladores inyectadas al arrancar el servidor.
type Dependencies struct {
	AuthCtrl     *controllers.AuthController
	ProdCtrl     *controllers.ProdController
	RedCtrl      *controllers.RedController
	CheckoutCtrl *controllers.CheckoutController
	HealthCtrl   *controllers.HealthController
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
			AuthCtrl:     controllers.NewAuthController(repo),
			ProdCtrl:     controllers.NewProdController(repo, cache),
			RedCtrl:      controllers.NewRedController(repo, cache),
			CheckoutCtrl: controllers.NewCheckoutController(repo, cache),
			HealthCtrl:   controllers.NewHealthController(),
		}
	}

	// 1. Endpoint de verificación de salud
	app.Get("/api/health", d.HealthCtrl.HealthCheck)

	// 2. Endpoints de autenticación
	app.Post("/api/login", d.AuthCtrl.Login)

	// 3. Endpoints del catálogo de productos (Cache-Aside en Redis)
	app.Get("/api/productos", d.ProdCtrl.GetProductos)
	app.Get("/api/productos/:id", d.ProdCtrl.GetProductoPorID)

	// 4. Endpoint de red multinivel de referidos (Cache-Aside en Redis)
	app.Get("/api/red", d.RedCtrl.GetRed)

	// 5. Endpoint de checkout y compra (Transacciones ACID en PostgreSQL y Goroutines)
	app.Post("/api/checkout", d.CheckoutCtrl.Checkout)
}