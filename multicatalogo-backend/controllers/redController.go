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

// RedController gestiona las operaciones de la red multinivel con Inyección de Dependencias.
type RedController struct {
	repo  repository.DBRepository
	cache config.CacheService
}

// NewRedController crea una nueva instancia inyectándole el repositorio y la caché.
func NewRedController(repo repository.DBRepository, cache config.CacheService) *RedController {
	return &RedController{
		repo:  repo,
		cache: cache,
	}
}

// GetRed devuelve la estructura de árbol de la red de referidos con optimización de caché Redis.
func (rc *RedController) GetRed(c *fiber.Ctx) error {
	email := c.Query("email")
	if email == "" {
		email = c.Get("X-User-Email")
	}
	if email == "" {
		email = "admin@upse.edu.ec"
	}

	cacheKey := fmt.Sprintf("red:tree:%s", email)
	ctx := context.Background()

	// 1. Intento de lectura desde Redis (Cache Hit)
	if rc.cache != nil {
		cachedBytes, err := rc.cache.Get(ctx, cacheKey)
		if err == nil && len(cachedBytes) > 0 {
			var red models.Referido
			if err := json.Unmarshal(cachedBytes, &red); err == nil {
				c.Set("X-Cache", "HIT")
				return c.Status(fiber.StatusOK).JSON(red)
			}
		}
	}

	// 2. Cache Miss: Reconstrucción desde PostgreSQL
	red, err := rc.repo.ObtenerRed()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIError{
			Status:  fiber.StatusInternalServerError,
			Message: "Error al recuperar la red de referidos de la base de datos",
			Details: err.Error(),
		})
	}

	// 3. Almacenamiento en caché de Redis con TTL de 10 minutos
	if rc.cache != nil {
		if jsonData, err := json.Marshal(red); err == nil {
			_ = rc.cache.Set(ctx, cacheKey, jsonData, 10*time.Minute)
		}
	}

	c.Set("X-Cache", "MISS")
	return c.Status(fiber.StatusOK).JSON(red)
}

// GetRed función libre para retrocompatibilidad
func GetRed(c *fiber.Ctx) error {
	ctrl := NewRedController(repository.NewPostgresRepository(config.GetDB()), config.NewRedisCacheService(config.GetRedis()))
	return ctrl.GetRed(c)
}
