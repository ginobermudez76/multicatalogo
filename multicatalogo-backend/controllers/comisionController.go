package controllers

import (
	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/config"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// ComisionController gestiona comisiones y liquidaciones (Fase 4 & 5).
type ComisionController struct {
	repo  repository.DBRepository
	cache config.CacheService
}

// NewComisionController crea una nueva instancia de ComisionController.
func NewComisionController(repo repository.DBRepository, cache config.CacheService) *ComisionController {
	return &ComisionController{repo: repo, cache: cache}
}

// MiRed obtiene el subárbol de la red del usuario autenticado (Fase 4).
func (cc *ComisionController) MiRed(c *fiber.Ctx) error {
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

	red, err := cc.repo.ObtenerRedUsuario(uid)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIError{
			Status:  fiber.StatusInternalServerError,
			Message: "Error al obtener red",
			Details: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(red)
}

// MisComisiones obtiene el ledger de comisiones del usuario (Fase 4).
func (cc *ComisionController) MisComisiones(c *fiber.Ctx) error {
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

	comisiones, err := cc.repo.ObtenerComisionesUsuario(uid)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIError{
			Status:  fiber.StatusInternalServerError,
			Message: "Error al obtener comisiones",
			Details: err.Error(),
		})
	}

	// Calcular totales
	var totalPendiente, totalPagado float64
	for _, com := range comisiones {
		if com.Estado == "pendiente" {
			totalPendiente += com.Monto
		} else if com.Estado == "pagada" {
			totalPagado += com.Monto
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"comisiones":      comisiones,
		"total_pendiente": totalPendiente,
		"total_pagado":    totalPagado,
	})
}

// CrearLiquidacion crea una liquidación de comisiones (Fase 5).
func (cc *ComisionController) CrearLiquidacion(c *fiber.Ctx) error {
	var req struct {
		ReferidoID int `json:"referido_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "Cuerpo de petición inválido",
			Details: err.Error(),
		})
	}

	if req.ReferidoID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "ReferidoID es obligatorio",
		})
	}

	liquidacion, err := cc.repo.CrearLiquidacion(req.ReferidoID)
	if err != nil {
		if err.Error() == "no hay comisiones pendientes para liquidar" {
			return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
				Status:  fiber.StatusBadRequest,
				Message: err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIError{
			Status:  fiber.StatusInternalServerError,
			Message: "Error al crear liquidación",
			Details: err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(liquidacion)
}

// ListarLiquidaciones obtiene todas las liquidaciones (Fase 5).
func (cc *ComisionController) ListarLiquidaciones(c *fiber.Ctx) error {
	liquidaciones, err := cc.repo.ObtenerLiquidaciones()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIError{
			Status:  fiber.StatusInternalServerError,
			Message: "Error al obtener liquidaciones",
			Details: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(liquidaciones)
}

// MisLiquidaciones obtiene liquidaciones del usuario autenticado (Fase 5).
func (cc *ComisionController) MisLiquidaciones(c *fiber.Ctx) error {
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

	liquidaciones, err := cc.repo.ObtenerLiquidacionesUsuario(uid)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIError{
			Status:  fiber.StatusInternalServerError,
			Message: "Error al obtener liquidaciones",
			Details: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(liquidaciones)
}
