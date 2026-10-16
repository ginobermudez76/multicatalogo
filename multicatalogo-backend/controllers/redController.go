// Declaramos el paquete controllers para la lógica de la red multinivel.
package controllers

import (
	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// GetRed es la función controladora encargada de devolver la estructura de árbol de la red de referidos desde PostgreSQL.
func GetRed(c *fiber.Ctx) error {
	red, err := repository.ObtenerRed()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIError{
			Status:  fiber.StatusInternalServerError,
			Message: "Error al recuperar la red de referidos de la base de datos",
			Details: err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(red)
}
