// Declaramos el paquete routes, cuya única responsabilidad es registrar los endpoints de la API.
package routes

import (
	// Importamos Fiber para poder recibir y configurar la instancia principal de la aplicación.
	"github.com/gofiber/fiber/v2"
	// Importamos nuestro paquete de controladores, donde residen las funciones que se ejecutarán en cada ruta.
	"multicatalogo-backend/controllers"
)

// SetupRoutes es una función que recibe un puntero a la aplicación Fiber (*fiber.App) para inyectarle las rutas.
func SetupRoutes(app *fiber.App) {
	// Endpoint de comprobación de salud del servidor (Health Check)
	app.Get("/api/health", controllers.HealthCheck)

	// Definimos una ruta HTTP POST en "/api/login" y la enlazamos a la función Login.
	app.Post("/api/login", controllers.Login)

	// Definimos rutas HTTP para el catálogo de productos relacional.
	app.Get("/api/productos", controllers.GetProductos)
	app.Get("/api/productos/:id", controllers.GetProductoPorID)

	// Definimos la ruta HTTP para la red multinivel de referidos desde PostgreSQL.
	app.Get("/api/red", controllers.GetRed)
}