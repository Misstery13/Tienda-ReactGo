// Declaramos el paquete controllers para agrupar las funciones que manejan la lógica de negocio de las rutas.
package controllers

import (
	"errors"
	"log"
	"strings"

	// Importamos el framework Fiber para tener acceso al contexto (c *fiber.Ctx) de la petición HTTP.
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"

	"multicatalogo-backend/auth"
	// Importamos nuestro paquete de modelos para poder usar la estructura LoginRequest.
	"multicatalogo-backend/models"
	// El repositorio es el que conoce los datos; el controlador solo maneja la petición y la respuesta.
	"multicatalogo-backend/repository"
)

// AuthController recibe el repositorio y el secreto de firma por inyección.
type AuthController struct {
	repo    repository.DBRepository
	secreto string
}

// NuevoAuthController construye el controlador con sus dependencias.
func NuevoAuthController(repo repository.DBRepository, secreto string) *AuthController {
	return &AuthController{repo: repo, secreto: secreto}
}

// Login valida las credenciales recibidas y devuelve un JWT firmado y el rol del usuario.
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

	usuario, err := ac.repo.BuscarUsuarioPorEmail(c.Context(), email)
	if err != nil && !errors.Is(err, repository.ErrNoEncontrado) {
		log.Printf("error en login: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NuevoAPIError(
			fiber.StatusInternalServerError,
			"Error al consultar la base de datos",
			nil,
		))
	}

	// Mismo mensaje si el usuario no existe o si la contraseña es incorrecta:
	// así no se revela qué correos están registrados.
	credencialesInvalidas := errors.Is(err, repository.ErrNoEncontrado)
	if !credencialesInvalidas {
		// bcrypt compara la contraseña contra el hash guardado en PostgreSQL.
		if bcrypt.CompareHashAndPassword([]byte(usuario.Password), []byte(password)) != nil {
			credencialesInvalidas = true
		}
	}

	if credencialesInvalidas {
		return c.Status(fiber.StatusUnauthorized).JSON(models.NuevoAPIError(
			fiber.StatusUnauthorized,
			"Credenciales incorrectas",
			nil,
		))
	}

	// Una cuenta desactivada no puede iniciar sesión aunque la contraseña sea correcta.
	if !usuario.Activo {
		return c.Status(fiber.StatusForbidden).JSON(models.NuevoAPIError(
			fiber.StatusForbidden,
			"La cuenta está desactivada; contacta al administrador",
			nil,
		))
	}

	// Token firmado: a diferencia del anterior, no se puede inventar a mano.
	token, err := auth.GenerarToken(usuario, ac.secreto)
	if err != nil {
		log.Printf("error firmando el token: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NuevoAPIError(
			fiber.StatusInternalServerError,
			"No se pudo iniciar la sesión",
			nil,
		))
	}

	// Respuesta exitosa: mismo contrato JSON que el frontend ya consume (token, email, rol).
	return c.JSON(models.LoginResponse{
		Token: token,
		Email: usuario.Email,
		Rol:   usuario.Rol,
	})
}
