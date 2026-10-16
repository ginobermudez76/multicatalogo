// Declaramos el paquete controllers para la lógica de productos.
package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	// Importamos Fiber para manejar el contexto y respuestas HTTP.
	"github.com/gofiber/fiber/v2"
	// Importamos nuestro paquete config para acceder a Redis.
	"multicatalogo-backend/config"
	// Importamos nuestro paquete de modelos para usar la estructura APIError.
	"multicatalogo-backend/models"
	// Importamos el repositorio para acceder a la capa de datos de productos.
	"multicatalogo-backend/repository"
)

// GetProductos es la función controladora encargada de devolver el catálogo completo de artículos.
// Implementa el patrón Cache-Aside con Redis (clave 'productos:all' y TTL de 5 minutos).
func GetProductos(c *fiber.Ctx) error {
	redisClient := config.GetRedis()
	ctx := context.Background()
	cacheKey := "productos:all"

	// 1. Intento de lectura desde Redis (Cache-Aside: Cache Hit)
	if redisClient != nil {
		cachedBytes, err := redisClient.Get(ctx, cacheKey).Bytes()
		if err == nil {
			var productos []models.Producto
			if err := json.Unmarshal(cachedBytes, &productos); err == nil {
				c.Set("X-Cache", "HIT")
				return c.Status(fiber.StatusOK).JSON(productos)
			}
		}
	}

	// 2. Cache Miss: Consultamos la base de datos relacional PostgreSQL
	productos, err := repository.ObtenerProductos()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIError{
			Status:  fiber.StatusInternalServerError,
			Message: "Error al recuperar productos de la base de datos",
			Details: err.Error(),
		})
	}

	// 3. Almacenamos en caché de Redis con un TTL de 5 minutos
	if redisClient != nil {
		if jsonData, err := json.Marshal(productos); err == nil {
			_ = redisClient.Set(ctx, cacheKey, jsonData, 5*time.Minute).Err()
		}
	}

	c.Set("X-Cache", "MISS")
	return c.Status(fiber.StatusOK).JSON(productos)
}

// GetProductoPorID es la función controladora encargada de buscar y retornar un producto por su ID.
func GetProductoPorID(c *fiber.Ctx) error {
	// Extraemos el parámetro 'id' de la URL.
	idParam := c.Params("id")

	// Convertimos el parámetro textual a entero (int).
	id, err := strconv.Atoi(idParam)
	if err != nil {
		// Si la conversión falla (ej. se envió texto en lugar de número), retornamos HTTP 400 con APIError.
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "El parámetro ID debe ser un número entero válido",
			Details: err.Error(),
		})
	}

	// Buscamos el producto en la base de datos a través del repositorio.
	producto, err := repository.ObtenerProductoPorID(id)
	if err != nil {
		if errors.Is(err, repository.ErrProductoNoEncontrado) {
			return c.Status(fiber.StatusNotFound).JSON(models.APIError{
				Status:  fiber.StatusNotFound,
				Message: "Producto no encontrado",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIError{
			Status:  fiber.StatusInternalServerError,
			Message: "Error al consultar producto en la base de datos",
			Details: err.Error(),
		})
	}

	// Si fue encontrado, retornamos HTTP 200 con el producto en formato JSON.
	return c.Status(fiber.StatusOK).JSON(producto)
}