// Declaramos el paquete controllers para la lógica de la red multinivel.
package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/config"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// GetRed es la función controladora encargada de devolver la estructura de árbol de la red de referidos.
// Implementa el patrón Cache-Aside con Redis bajo la clave 'red:tree:{email_usuario}' con un TTL de 10 minutos.
func GetRed(c *fiber.Ctx) error {
	// Determinamos el email del usuario para particionar la clave de caché
	email := c.Query("email")
	if email == "" {
		email = c.Get("X-User-Email")
	}
	if email == "" {
		email = "admin@upse.edu.ec"
	}

	cacheKey := fmt.Sprintf("red:tree:%s", email)
	redisClient := config.GetRedis()
	ctx := context.Background()

	// 1. Intento de lectura desde Redis (Cache Hit)
	if redisClient != nil {
		cachedBytes, err := redisClient.Get(ctx, cacheKey).Bytes()
		if err == nil {
			var red models.Referido
			if err := json.Unmarshal(cachedBytes, &red); err == nil {
				c.Set("X-Cache", "HIT")
				return c.Status(fiber.StatusOK).JSON(red)
			}
		}
	}

	// 2. Cache Miss: Reconstrucción costosa del árbol relacional desde PostgreSQL
	red, err := repository.ObtenerRed()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIError{
			Status:  fiber.StatusInternalServerError,
			Message: "Error al recuperar la red de referidos de la base de datos",
			Details: err.Error(),
		})
	}

	// 3. Almacenamos el árbol resultante en Redis con TTL de 10 minutos
	if redisClient != nil {
		if jsonData, err := json.Marshal(red); err == nil {
			_ = redisClient.Set(ctx, cacheKey, jsonData, 10*time.Minute).Err()
		}
	}

	c.Set("X-Cache", "MISS")
	return c.Status(fiber.StatusOK).JSON(red)
}
