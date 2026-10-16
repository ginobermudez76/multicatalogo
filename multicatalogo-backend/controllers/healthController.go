// Declaramos el paquete controllers para la lógica de verificación del servidor.
package controllers

import "github.com/gofiber/fiber/v2"

// HealthController gestiona el endpoint de verificación de salud del sistema.
type HealthController struct{}

// NewHealthController constructor de HealthController.
func NewHealthController() *HealthController {
	return &HealthController{}
}

// HealthCheck verifica el estado operativo del servidor Go/Fiber.
func (hc *HealthController) HealthCheck(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "ok",
		"message": "Servidor Go/Fiber operativo",
	})
}

// HealthCheck función libre para retrocompatibilidad
func HealthCheck(c *fiber.Ctx) error {
	return NewHealthController().HealthCheck(c)
}
