package controllers

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// PedidoController gestiona operaciones de pedidos (Fase 3).
type PedidoController struct {
	repo repository.DBRepository
}

// NewPedidoController crea una nueva instancia de PedidoController.
func NewPedidoController(repo repository.DBRepository) *PedidoController {
	return &PedidoController{repo: repo}
}

// MisPedidos obtiene los pedidos del usuario autenticado (Fase 3).
func (pc *PedidoController) MisPedidos(c *fiber.Ctx) error {
	// Obtener usuario_id del token (simulado - en producción usar middleware de JWT)
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.APIError{
			Status:  fiber.StatusUnauthorized,
			Message: "Usuario no autenticado",
		})
	}

	uid, ok := userID.(int)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(models.APIError{
			Status:  fiber.StatusUnauthorized,
			Message: "ID de usuario inválido",
		})
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	pedidos, total, err := pc.repo.ObtenerPedidosUsuario(uid, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIError{
			Status:  fiber.StatusInternalServerError,
			Message: "Error al obtener pedidos",
			Details: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(models.PaginaResponse{
		Data:  pedidos,
		Page:  page,
		Limit: limit,
		Total: total,
	})
}

// ListarPedidosAdmin obtiene todos los pedidos con filtro de estado (Fase 3).
func (pc *PedidoController) ListarPedidosAdmin(c *fiber.Ctx) error {
	estado := c.Query("estado", "")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	pedidos, total, err := pc.repo.ObtenerPedidosAdmin(estado, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIError{
			Status:  fiber.StatusInternalServerError,
			Message: "Error al obtener pedidos",
			Details: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(models.PaginaResponse{
		Data:  pedidos,
		Page:  page,
		Limit: limit,
		Total: total,
	})
}

// ObtenerPedidoDetalle obtiene un pedido con sus ítems (Fase 3).
func (pc *PedidoController) ObtenerPedidoDetalle(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "ID inválido",
		})
	}

	pedido, items, err := pc.repo.ObtenerPedidoPorID(id)
	if err != nil {
		if strings.Contains(err.Error(), "no encontrado") {
			return c.Status(fiber.StatusNotFound).JSON(models.APIError{
				Status:  fiber.StatusNotFound,
				Message: "Pedido no encontrado",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIError{
			Status:  fiber.StatusInternalServerError,
			Message: "Error al obtener pedido",
			Details: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"pedido": pedido,
		"items":  items,
	})
}

// ActualizarEstadoPedido cambia el estado de un pedido (Fase 3).
func (pc *PedidoController) ActualizarEstadoPedido(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "ID inválido",
		})
	}

	var req struct {
		Estado string `json:"estado"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "Cuerpo de petición inválido",
			Details: err.Error(),
		})
	}

	if strings.TrimSpace(req.Estado) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "Estado es obligatorio",
		})
	}

	err = pc.repo.ActualizarEstadoPedido(id, req.Estado)
	if err != nil {
		if strings.Contains(err.Error(), "no encontrado") {
			return c.Status(fiber.StatusNotFound).JSON(models.APIError{
				Status:  fiber.StatusNotFound,
				Message: "Pedido no encontrado",
			})
		}
		if strings.Contains(err.Error(), "transición inválida") {
			return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
				Status:  fiber.StatusBadRequest,
				Message: err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIError{
			Status:  fiber.StatusInternalServerError,
			Message: "Error al actualizar estado",
			Details: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Estado actualizado correctamente",
	})
}
