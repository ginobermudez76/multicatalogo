package middleware

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// RequireAuth verifica que el usuario esté autenticado (middleware simple).
// En producción, esto debería validar un JWT real.
func RequireAuth(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  fiber.StatusUnauthorized,
			"message": "Token de autenticación requerido",
		})
	}

	// Extraer token del header "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  fiber.StatusUnauthorized,
			"message": "Formato de token inválido",
		})
	}

	token := parts[1]

	// Parse simple del token simulado: "jwt-token-<id>-<rol>"
	if !strings.HasPrefix(token, "jwt-token-") {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  fiber.StatusUnauthorized,
			"message": "Token inválido",
		})
	}

	// Extraer ID y rol del token (simulado)
	tokenParts := strings.Split(token, "-")
	if len(tokenParts) < 4 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  fiber.StatusUnauthorized,
			"message": "Token inválido",
		})
	}

	// Convertir ID a entero
	userID, err := strconv.Atoi(tokenParts[2])
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  fiber.StatusUnauthorized,
			"message": "Token inválido",
		})
	}

	// En un sistema real, aquí validaríamos el JWT con una clave secreta
	// Por ahora, extraemos el user_id del token para uso en controladores
	c.Locals("user_id", userID)
	c.Locals("user_rol", tokenParts[3])

	return c.Next()
}

// RequireAdmin verifica que el usuario tenga rol admin.
func RequireAdmin(c *fiber.Ctx) error {
	rol := c.Locals("user_rol")
	if rol == nil || rol != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status":  fiber.StatusForbidden,
			"message": "Se requiere rol de administrador",
		})
	}

	return c.Next()
}
