// El paquete controllers también expone el estado del servicio.
package controllers

import (
	"github.com/gofiber/fiber/v2"
)

// HealthCheck responde si el servidor está operativo.
// Sirve para comprobar desde el navegador o Postman que la API está levantada.
func HealthCheck(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "ok",
		"message": "Servidor Go/Fiber operativo",
	})
}
