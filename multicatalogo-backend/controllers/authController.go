// Declaramos el paquete controllers para agrupar las funciones que manejan la lógica de negocio de las rutas.
package controllers

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// AuthController gestiona la autenticación con Inyección de Dependencias (Fase 4).
type AuthController struct {
	repo repository.DBRepository
}

// NewAuthController crea una nueva instancia de AuthController inyectándole el repositorio.
func NewAuthController(repo repository.DBRepository) *AuthController {
	return &AuthController{repo: repo}
}

// Login es el método controlador que se ejecuta cuando el cliente envía sus credenciales.
func (ac *AuthController) Login(c *fiber.Ctx) error {
	var req models.LoginRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "Cuerpo de petición inválido",
			Details: err.Error(),
		})
	}

	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "El email y la contraseña son obligatorios y no pueden estar vacíos",
		})
	}

	userAuth, err := ac.repo.AuthenticateUser(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, repository.ErrCredencialesInvalidas) {
			return c.Status(fiber.StatusUnauthorized).JSON(models.APIError{
				Status:  fiber.StatusUnauthorized,
				Message: "Credenciales incorrectas",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIError{
			Status:  fiber.StatusInternalServerError,
			Message: "Error interno al verificar credenciales en la base de datos",
			Details: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(userAuth)
}

// Login función libre para retrocompatibilidad
func Login(c *fiber.Ctx) error {
	ctrl := NewAuthController(repository.NewPostgresRepository(nil))
	return ctrl.Login(c)
}