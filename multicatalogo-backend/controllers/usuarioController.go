package controllers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// UsuarioController gestiona operaciones de usuarios (Fase 1).
type UsuarioController struct {
	repo repository.DBRepository
}

// NewUsuarioController crea una nueva instancia de UsuarioController.
func NewUsuarioController(repo repository.DBRepository) *UsuarioController {
	return &UsuarioController{repo: repo}
}

// Register registra un nuevo usuario con código de referido opcional (Fase 1).
func (uc *UsuarioController) Register(c *fiber.Ctx) error {
	var req models.RegisterRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "Cuerpo de petición inválido",
			Details: err.Error(),
		})
	}

	if strings.TrimSpace(req.Nombre) == "" || strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "Nombre, email y contraseña son obligatorios",
		})
	}

	usuario, err := uc.repo.RegistrarUsuario(&req)
	if err != nil {
		if strings.Contains(err.Error(), "email ya está registrado") {
			return c.Status(fiber.StatusConflict).JSON(models.APIError{
				Status:  fiber.StatusConflict,
				Message: "El email ya está registrado",
			})
		}
		if strings.Contains(err.Error(), "código de referido inválido") {
			return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
				Status:  fiber.StatusBadRequest,
				Message: "Código de referido inválido",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIError{
			Status:  fiber.StatusInternalServerError,
			Message: "Error al registrar usuario",
			Details: err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(usuario)
}
