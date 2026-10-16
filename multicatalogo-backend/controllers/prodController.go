// Declaramos el paquete controllers para la lógica de productos.
package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/config"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// ProdController gestiona las operaciones del catálogo de productos con Inyección de Dependencias.
type ProdController struct {
	repo  repository.DBRepository
	cache config.CacheService
}

// NewProdController es el constructor que inyecta el repositorio de base de datos y el servicio de caché.
func NewProdController(repo repository.DBRepository, cache config.CacheService) *ProdController {
	return &ProdController{
		repo:  repo,
		cache: cache,
	}
}

// GetProductos devuelve el catálogo completo de artículos implementando el patrón Cache-Aside con Redis.
func (pc *ProdController) GetProductos(c *fiber.Ctx) error {
	ctx := context.Background()
	cacheKey := "productos:all"

	// 1. Intento de lectura desde Redis (Cache Hit)
	if pc.cache != nil {
		cachedBytes, err := pc.cache.Get(ctx, cacheKey)
		if err == nil && len(cachedBytes) > 0 {
			var productos []models.Producto
			if err := json.Unmarshal(cachedBytes, &productos); err == nil {
				c.Set("X-Cache", "HIT")
				return c.Status(fiber.StatusOK).JSON(productos)
			}
		}
	}

	// 2. Cache Miss: Consultamos la base de datos relacional PostgreSQL
	productos, err := pc.repo.ObtenerProductos()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIError{
			Status:  fiber.StatusInternalServerError,
			Message: "Error al recuperar productos de la base de datos",
			Details: err.Error(),
		})
	}

	// 3. Almacenamos en caché de Redis con un TTL de 5 minutos
	if pc.cache != nil {
		if jsonData, err := json.Marshal(productos); err == nil {
			_ = pc.cache.Set(ctx, cacheKey, jsonData, 5*time.Minute)
		}
	}

	c.Set("X-Cache", "MISS")
	return c.Status(fiber.StatusOK).JSON(productos)
}

// GetProductoPorID busca y retorna un producto por su clave primaria numérica.
func (pc *ProdController) GetProductoPorID(c *fiber.Ctx) error {
	idParam := c.Params("id")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "El parámetro ID debe ser un número entero válido",
			Details: err.Error(),
		})
	}

	producto, err := pc.repo.ObtenerProductoPorID(id)
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

	return c.Status(fiber.StatusOK).JSON(producto)
}

// Funciones libres para retrocompatibilidad
func GetProductos(c *fiber.Ctx) error {
	ctrl := NewProdController(repository.NewPostgresRepository(config.GetDB()), config.NewRedisCacheService(config.GetRedis()))
	return ctrl.GetProductos(c)
}

func GetProductoPorID(c *fiber.Ctx) error {
	ctrl := NewProdController(repository.NewPostgresRepository(config.GetDB()), config.NewRedisCacheService(config.GetRedis()))
	return ctrl.GetProductoPorID(c)
}