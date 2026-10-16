// Middlewares de autenticación y autorización.
// Son la frontera real de seguridad: las guardas de React son solo comodidad visual.
package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"multicatalogo-backend/auth"
	"multicatalogo-backend/models"
)

// Claves con las que el middleware deja la identidad en el contexto de Fiber.
const (
	ClaveUsuarioID = "usuario_id"
	ClaveEmail     = "usuario_email"
	ClaveRol       = "usuario_rol"
)

// RequireAuth exige un JWT válido en la cabecera Authorization.
func RequireAuth(secreto string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		cabecera := c.Get("Authorization")

		// Formato esperado: "Bearer <token>"
		partes := strings.SplitN(cabecera, " ", 2)
		if len(partes) != 2 || !strings.EqualFold(partes[0], "Bearer") {
			return c.Status(fiber.StatusUnauthorized).JSON(models.NuevoAPIError(
				fiber.StatusUnauthorized,
				"Falta el token de autenticación",
				nil,
			))
		}

		claims, err := auth.VerificarToken(strings.TrimSpace(partes[1]), secreto)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(models.NuevoAPIError(
				fiber.StatusUnauthorized,
				"Token inválido o expirado",
				nil,
			))
		}

		// La identidad queda disponible para los controladores.
		c.Locals(ClaveUsuarioID, claims.Subject)
		c.Locals(ClaveEmail, claims.Email)
		c.Locals(ClaveRol, claims.Rol)

		return c.Next()
	}
}

// RequireAdmin se encadena después de RequireAuth y exige el rol admin.
func RequireAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if rol, _ := c.Locals(ClaveRol).(string); rol != "admin" {
			return c.Status(fiber.StatusForbidden).JSON(models.NuevoAPIError(
				fiber.StatusForbidden,
				"Esta sección es solo para administradores",
				nil,
			))
		}
		return c.Next()
	}
}

// EmailDeSesion devuelve el correo que venía firmado en el token.
func EmailDeSesion(c *fiber.Ctx) string {
	email, _ := c.Locals(ClaveEmail).(string)
	return email
}
