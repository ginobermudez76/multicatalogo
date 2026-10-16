// Declaramos el paquete config para la inicialización y gestión de clientes de infraestructura.
package config

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	// Importamos la librería oficial de Redis v9 para Go.
	"github.com/redis/go-redis/v9"
)

// RedisClient es la instancia global del cliente de Redis.
var RedisClient *redis.Client

// ConnectRedis inicializa la conexión con Redis leyendo las variables de entorno.
func ConnectRedis() (*redis.Client, error) {
	host := getEnv("REDIS_HOST", "localhost")
	port := getEnv("REDIS_PORT", "6379")
	password := getEnv("REDIS_PASSWORD", "")
	dbStr := getEnv("REDIS_DB", "0")

	dbNum, err := strconv.Atoi(dbStr)
	if err != nil {
		dbNum = 0
	}

	addr := fmt.Sprintf("%s:%s", host, port)

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       dbNum,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Verificamos conectividad con Ping
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("error al conectar con Redis en %s: %w", addr, err)
	}

	log.Printf("Conexión exitosa a Redis en %s (DB: %d)", addr, dbNum)
	RedisClient = client
	return RedisClient, nil
}

// GetRedis retorna la instancia actual del cliente de Redis.
func GetRedis() *redis.Client {
	return RedisClient
}

// CloseRedis cierra la conexión con Redis al finalizar el ciclo de vida del servidor.
func CloseRedis() error {
	if RedisClient != nil {
		log.Println("Cerrando cliente de Redis...")
		return RedisClient.Close()
	}
	return nil
}
