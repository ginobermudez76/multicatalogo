// Declaramos que este es el paquete principal (main), es decir, el archivo ejecutable que inicia el programa.
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	// Importamos el framework principal Fiber.
	"github.com/gofiber/fiber/v2"
	// Importamos el middleware CORS para gestionar la seguridad entre distintos puertos/dominios.
	"github.com/gofiber/fiber/v2/middleware/cors"
	// Importamos los paquetes de configuración, controladores, repositorios y rutas.
	"multicatalogo-backend/config"
	"multicatalogo-backend/controllers"
	"multicatalogo-backend/repository"
	"multicatalogo-backend/routes"
)

// func main es el punto de entrada de la aplicación en Go. Todo comienza a ejecutarse aquí.
func main() {
	// 1. Inicializamos el pool de conexiones a la base de datos PostgreSQL (Fase 2)
	db, err := config.ConnectDB()
	if err != nil {
		log.Fatalf("Error crítico al inicializar la base de datos: %v", err)
	}
	defer config.CloseDB()

	// 2. Inicializamos el cliente de Redis para almacenamiento en caché (Fase 3)
	redisClient, err := config.ConnectRedis()
	if err != nil {
		log.Printf("Aviso: No se pudo conectar a Redis (%v). El servidor continuará sin caché", err)
	}
	defer config.CloseRedis()

	// 3. Inyección de Dependencias (Fase 4 - Paso 1)
	repo := repository.NewPostgresRepository(db)
	cacheService := config.NewRedisCacheService(redisClient)

	deps := routes.Dependencies{
		AuthCtrl:           controllers.NewAuthController(repo),
		ProdCtrl:           controllers.NewProdController(repo, cacheService),
		RedCtrl:            controllers.NewRedController(repo, cacheService),
		CheckoutCtrl:       controllers.NewCheckoutController(repo, cacheService),
		HealthCtrl:         controllers.NewHealthController(),
		UsuarioCtrl:        controllers.NewUsuarioController(repo),
		ProductoAdminCtrl:  controllers.NewProductoAdminController(repo, cacheService),
		PedidoCtrl:         controllers.NewPedidoController(repo),
		ComisionCtrl:       controllers.NewComisionController(repo, cacheService),
		ReporteCtrl:        controllers.NewReporteController(repo),
	}

	// 4. Instanciamos la aplicación de Fiber
	app := fiber.New()

	// Implementamos el middleware CORS a nivel global
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173, http://127.0.0.1:5173",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-User-Email",
	}))

	// Registramos las rutas inyectando las dependencias estructuradas
	routes.SetupRoutes(app, deps)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3002"
	}

	// Manejo de señales del sistema para apagado controlado (Graceful Shutdown)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Señal de apagado recibida. Cerrando servidor HTTP...")
		_ = app.Shutdown()
	}()

	log.Printf("Servidor backend Go escuchando en http://localhost:%s", port)
	// Ponemos a la aplicación a escuchar peticiones en el puerto configurado.
	if err := app.Listen(":" + port); err != nil {
		log.Printf("Servidor cerrado: %v", err)
	}
}