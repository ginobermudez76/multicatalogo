// Declaramos el paquete controllers para agrupar las funciones que manejan la lógica de negocio de las rutas.
package controllers

import (
	"errors"
	"strings"

	// Importamos el framework Fiber para tener acceso al contexto (c *fiber.Ctx) de la petición HTTP.
	"github.com/gofiber/fiber/v2"
	// Importamos nuestro paquete de modelos para poder usar LoginRequest y APIError.
	"multicatalogo-backend/models"
	// Importamos el repositorio para delegar la autenticación de usuarios.
	"multicatalogo-backend/repository"
)

// Login es la función controladora que se ejecutará cuando el cliente envíe sus credenciales.
func Login(c *fiber.Ctx) error {
	// Creamos una variable 'req' del tipo LoginRequest para almacenar los datos recibidos.
	var req models.LoginRequest

	// Intentamos parsear (transformar) el cuerpo JSON entrante y guardarlo en la variable 'req'.
	if err := c.BodyParser(&req); err != nil {
		// Si ocurre un error al parsear (ej. JSON mal formado), retornamos HTTP 400 con APIError.
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "Cuerpo de petición inválido",
			Details: err.Error(),
		})
	}

	// Validamos explícitamente que los campos Email y Password no vengan vacíos antes de enviarlos al repositorio.
	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "El email y la contraseña son obligatorios y no pueden estar vacíos",
		})
	}

	// Delegamos la validación de credenciales a la capa del repositorio con PostgreSQL.
	userAuth, err := repository.AuthenticateUser(req.Email, req.Password)
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

	// Si la autenticación es exitosa, retornamos HTTP 200 con el token, correo y rol.
	return c.Status(fiber.StatusOK).JSON(userAuth)
}