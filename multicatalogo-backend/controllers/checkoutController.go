// Declaramos el paquete controllers para la gestión de pagos y transacciones del sistema.
package controllers

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/config"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// CheckoutController implementa el microservicio de finalización de compra con Inyección de Dependencias.
type CheckoutController struct {
	repo  repository.DBRepository
	cache config.CacheService
}

// NewCheckoutController es el constructor que inyecta las dependencias de base de datos y caché.
func NewCheckoutController(repo repository.DBRepository, cache config.CacheService) *CheckoutController {
	return &CheckoutController{
		repo:  repo,
		cache: cache,
	}
}

// Checkout procesa la compra mediante transacciones ACID y cálculo de comisiones asíncrono con Goroutines.
func (cc *CheckoutController) Checkout(c *fiber.Ctx) error {
	var req models.CheckoutRequest

	// 1. Parseo del cuerpo JSON recibido desde Checkout.tsx
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "Cuerpo de petición inválido para el checkout",
			Details: err.Error(),
		})
	}

	// 2. Validaciones de entrada
	if strings.TrimSpace(req.Cliente.Nombre) == "" || strings.TrimSpace(req.Cliente.Email) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "El nombre y el correo electrónico del cliente son obligatorios",
		})
	}

	if len(req.Items) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "El carrito de compras no contiene productos para procesar",
		})
	}

	if req.Total <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "El monto total de la orden debe ser mayor a cero",
		})
	}

	// 3. Ejecución de la Transacción ACID en PostgreSQL (Paso 3)
	pedidoConfirmado, err := cc.repo.CrearPedido(&req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIError{
			Status:  fiber.StatusInternalServerError,
			Message: "Error en la transacción ACID al procesar la orden (Rollback ejecutado)",
			Details: err.Error(),
		})
	}

	// 4. Concurrencia y Misión Crítica: Goroutine para calcular comisiones en segundo plano (Paso 4)
	go func(total float64, email string) {
		log.Printf("[Goroutine] Iniciando cálculo asíncrono de comisiones para la orden por $%.2f...", total)
		if err := cc.repo.ActualizarComisionesRed(total, email); err != nil {
			log.Printf("[Goroutine] Error al actualizar comisiones en PostgreSQL: %v", err)
			return
		}
		log.Printf("[Goroutine] Comisiones actualizadas exitosamente en la base de datos.")

		// 5. Invalidación de Caché en Redis (Paso 5)
		if cc.cache != nil {
			ctx := context.Background()
			_ = cc.cache.Delete(ctx, fmt.Sprintf("red:tree:%s", email))
			_ = cc.cache.Delete(ctx, "red:tree:admin@upse.edu.ec")
			_ = cc.cache.Delete(ctx, "productos:all")
			log.Printf("[Goroutine] Caché invalidada en Redis para asegurar datos frescos en el Dashboard.")
		}
	}(pedidoConfirmado.Total, req.Cliente.Email)

	// 6. Retorno inmediato al cliente (sin esperar a la Goroutine) con HTTP 201 Created
	return c.Status(fiber.StatusCreated).JSON(pedidoConfirmado)
}
