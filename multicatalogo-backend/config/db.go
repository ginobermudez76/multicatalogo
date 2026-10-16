// Declaramos el paquete config para gestionar la configuración y conexiones externas de la aplicación.
package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	// Importamos el driver pgx para database/sql estándar de Go.
	_ "github.com/jackc/pgx/v5/stdlib"
	// Importamos godotenv para cargar variables de entorno desde el archivo .env.
	"github.com/joho/godotenv"
)

// DB es la instancia global del pool de conexiones a la base de datos PostgreSQL.
var DB *sql.DB

// ConnectDB inicializa el pool de conexiones hacia PostgreSQL usando variables de entorno.
func ConnectDB() (*sql.DB, error) {
	// Intentamos cargar el archivo .env si existe en el directorio actual o en la raíz del backend
	if err := godotenv.Load(".env", "../.env"); err != nil {
		log.Println("Aviso: No se encontró archivo .env local, usando variables de entorno del sistema")
	}

	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5434")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "multicatalogo")
	sslmode := getEnv("DB_SSLMODE", "disable")

	// Construimos el Connection String (DSN) para PostgreSQL
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, password, host, port, dbname, sslmode)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("error al abrir la conexión con PostgreSQL: %w", err)
	}

	// Configuramos los parámetros del pool de conexiones
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verificamos que la conexión esté activa mediante un Ping
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error al conectar con PostgreSQL (Ping fallido): %w", err)
	}

	log.Printf("Conexión exitosa a PostgreSQL en %s:%s (BD: %s)", host, port, dbname)
	DB = db
	return DB, nil
}

// GetDB retorna la instancia actual de la base de datos.
func GetDB() *sql.DB {
	return DB
}

// CloseDB cierra el pool de conexiones al finalizar la ejecución del servidor.
func CloseDB() error {
	if DB != nil {
		log.Println("Cerrando pool de conexiones de PostgreSQL...")
		return DB.Close()
	}
	return nil
}

// getEnv obtiene una variable de entorno o devuelve un valor por defecto si está vacía.
func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
