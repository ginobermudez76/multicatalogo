// Declaramos el paquete controllers para la lógica de productos.
package controllers

import (
	"errors"
	"strconv"

	// Importamos Fiber para manejar el contexto y respuestas HTTP.
	"github.com/gofiber/fiber/v2"
	// Importamos nuestro paquete de modelos para usar la estructura APIError.
	"multicatalogo-backend/models"
	// Importamos el repositorio para acceder a la capa de datos de productos.
	"multicatalogo-backend/repository"
)

// GetProductos es la función controladora encargada de devolver el catálogo completo de artículos desde PostgreSQL.
func GetProductos(c *fiber.Ctx) error {
	productos, err := repository.ObtenerProductos()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIError{
			Status:  fiber.StatusInternalServerError,
			Message: "Error al recuperar productos de la base de datos",
			Details: err.Error(),
		})
	}
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