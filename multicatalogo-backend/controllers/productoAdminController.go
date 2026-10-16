package controllers

import (
	"context"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/config"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// ProductoAdminController gestiona el CRUD de productos para admin (Fase 2).
type ProductoAdminController struct {
	repo  repository.DBRepository
	cache config.CacheService
}

// NewProductoAdminController crea una nueva instancia de ProductoAdminController.
func NewProductoAdminController(repo repository.DBRepository, cache config.CacheService) *ProductoAdminController {
	return &ProductoAdminController{repo: repo, cache: cache}
}

// CrearProducto crea un nuevo producto (Fase 2).
func (pc *ProductoAdminController) CrearProducto(c *fiber.Ctx) error {
	var prod models.Producto

	if err := c.BodyParser(&prod); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "Cuerpo de petición inválido",
			Details: err.Error(),
		})
	}

	if strings.TrimSpace(prod.Nombre) == "" || prod.Precio <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "Nombre y precio válido son obligatorios",
		})
	}

	producto, err := pc.repo.CrearProducto(&prod)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIError{
			Status:  fiber.StatusInternalServerError,
			Message: "Error al crear producto",
			Details: err.Error(),
		})
	}

	// Invalidar caché
	if pc.cache != nil {
		pc.cache.Delete(context.Background(), "productos:all")
	}

	return c.Status(fiber.StatusCreated).JSON(producto)
}

// ActualizarProducto modifica un producto existente (Fase 2).
func (pc *ProductoAdminController) ActualizarProducto(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "ID inválido",
		})
	}

	var prod models.Producto
	if err := c.BodyParser(&prod); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "Cuerpo de petición inválido",
			Details: err.Error(),
		})
	}

	if strings.TrimSpace(prod.Nombre) == "" || prod.Precio <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "Nombre y precio válido son obligatorios",
		})
	}

	producto, err := pc.repo.ActualizarProducto(id, &prod)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIError{
			Status:  fiber.StatusInternalServerError,
			Message: "Error al actualizar producto",
			Details: err.Error(),
		})
	}

	// Invalidar caché
	if pc.cache != nil {
		pc.cache.Delete(context.Background(), "productos:all")
		pc.cache.Delete(context.Background(), "producto:"+strconv.Itoa(id))
	}

	return c.Status(fiber.StatusOK).JSON(producto)
}

// EliminarProducto hace borrado lógico de un producto (Fase 2).
func (pc *ProductoAdminController) EliminarProducto(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "ID inválido",
		})
	}

	err = pc.repo.EliminarProducto(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIError{
			Status:  fiber.StatusInternalServerError,
			Message: "Error al eliminar producto",
			Details: err.Error(),
		})
	}

	// Invalidar caché
	if pc.cache != nil {
		pc.cache.Delete(context.Background(), "productos:all")
		pc.cache.Delete(context.Background(), "producto:"+strconv.Itoa(id))
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Producto eliminado correctamente",
	})
}

// ListarProductosAdmin lista productos con búsqueda y paginación (Fase 2 & 7).
func (pc *ProductoAdminController) ListarProductosAdmin(c *fiber.Ctx) error {
	q := c.Query("q", "")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	productos, total, err := pc.repo.ObtenerProductosAdmin(q, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIError{
			Status:  fiber.StatusInternalServerError,
			Message: "Error al obtener productos",
			Details: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(models.PaginaResponse{
		Data:  productos,
		Page:  page,
		Limit: limit,
		Total: total,
	})
}
