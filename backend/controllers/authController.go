// Declaramos el paquete controllers para agrupar las funciones que manejan la lógica de negocio de las rutas.
package controllers

import (
	"strings"

	// Importamos el framework Fiber para tener acceso al contexto (c *fiber.Ctx) de la petición HTTP.
	"github.com/gofiber/fiber/v2"
	// Importamos nuestro paquete de modelos para poder usar la estructura LoginRequest.
	"multicatalogo-backend/models"
	// El repositorio es el que conoce los datos; el controlador solo maneja la petición y la respuesta.
	"multicatalogo-backend/repository"
)

// Login valida las credenciales recibidas y devuelve el token y el rol del usuario.
func Login(c *fiber.Ctx) error {
	// Creamos una variable 'req' del tipo LoginRequest (ubicada en nuestro paquete models) para almacenar los datos.
	var req models.LoginRequest

	// Intentamos parsear (transformar) el cuerpo JSON entrante y guardarlo en la variable 'req'.
	if err := c.BodyParser(&req); err != nil {
		// JSON mal formado: HTTP 400 con el formato estándar de error.
		return c.Status(fiber.StatusBadRequest).JSON(models.NuevoAPIError(
			fiber.StatusBadRequest,
			"Cuerpo de petición inválido",
			err.Error(),
		))
	}

	// Validación de entrada: ningún campo obligatorio puede llegar vacío.
	email := strings.TrimSpace(req.Email)
	password := strings.TrimSpace(req.Password)

	camposFaltantes := []string{}
	if email == "" {
		camposFaltantes = append(camposFaltantes, "email")
	}
	if password == "" {
		camposFaltantes = append(camposFaltantes, "password")
	}

	if len(camposFaltantes) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(models.NuevoAPIError(
			fiber.StatusBadRequest,
			"Los campos email y password son obligatorios",
			fiber.Map{"campos_faltantes": camposFaltantes},
		))
	}

	// El repositorio resuelve si las credenciales corresponden a un usuario registrado.
	usuario, encontrado := repository.BuscarUsuario(email, password)
	if !encontrado {
		return c.Status(fiber.StatusUnauthorized).JSON(models.NuevoAPIError(
			fiber.StatusUnauthorized,
			"Credenciales incorrectas",
			nil,
		))
	}

	// Respuesta exitosa: mismo contrato JSON que el frontend ya consume (token, email, rol).
	return c.JSON(models.LoginResponse{
		Token: usuario.Token,
		Email: usuario.Email,
		Rol:   usuario.Rol,
	})
}
