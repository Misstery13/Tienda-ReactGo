// Declaramos el paquete controllers para agrupar las funciones que manejan la lógica de negocio de las rutas.
package controllers

import (
	"errors"
	"log"
	"strings"

	// Importamos el framework Fiber para tener acceso al contexto (c *fiber.Ctx) de la petición HTTP.
	"github.com/gofiber/fiber/v2"
	// Importamos nuestro paquete de modelos para poder usar la estructura LoginRequest.
	"multicatalogo-backend/models"
	// El repositorio es el que conoce los datos; el controlador solo maneja la petición y la respuesta.
	"multicatalogo-backend/repository"
)

// AuthController recibe el repositorio por inyección de dependencias.
type AuthController struct {
	repo repository.DBRepository
}

// NuevoAuthController construye el controlador con su dependencia.
func NuevoAuthController(repo repository.DBRepository) *AuthController {
	return &AuthController{repo: repo}
}

// Login valida las credenciales recibidas y devuelve el token y el rol del usuario.
func (ac *AuthController) Login(c *fiber.Ctx) error {
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

	// El repositorio consulta PostgreSQL.
	usuario, err := ac.repo.BuscarUsuario(c.Context(), email, password)
	if errors.Is(err, repository.ErrNoEncontrado) {
		return c.Status(fiber.StatusUnauthorized).JSON(models.NuevoAPIError(
			fiber.StatusUnauthorized,
			"Credenciales incorrectas",
			nil,
		))
	}
	if err != nil {
		// Un fallo de base de datos no es culpa del cliente: HTTP 500.
		log.Printf("error en login: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NuevoAPIError(
			fiber.StatusInternalServerError,
			"Error al consultar la base de datos",
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
