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
	// Importamos nuestro paquete config para inicializar la conexión con PostgreSQL.
	"multicatalogo-backend/config"
	// Importamos nuestro propio paquete de rutas para delegarle la configuración de los endpoints.
	"multicatalogo-backend/routes"
)

// func main es el punto de entrada de la aplicación en Go. Todo comienza a ejecutarse aquí.
func main() {
	// Inicializamos el pool de conexiones a la base de datos PostgreSQL (Fase 2)
	if _, err := config.ConnectDB(); err != nil {
		log.Fatalf("Error crítico al inicializar la base de datos: %v", err)
	}
	// Aseguramos el cierre ordenado de la base de datos al finalizar la aplicación
	defer config.CloseDB()

	// Instanciamos una nueva aplicación de Fiber y la guardamos en la variable 'app'.
	app := fiber.New()

	// Implementamos el middleware CORS a nivel global usando app.Use() para interceptar todas las peticiones entrantes.
	app.Use(cors.New(cors.Config{
		// Configuramos el CORS para permitir peticiones provenientes del frontend local en Vite (puerto 5173).
		AllowOrigins: "http://localhost:5173, http://127.0.0.1:5173",
		// Declaramos de forma explícita qué cabeceras (Headers) se permitirán en la comunicación.
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Llamamos a la función SetupRoutes de nuestro paquete 'routes', enviándole la instancia de nuestra 'app'.
	routes.SetupRoutes(app)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
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