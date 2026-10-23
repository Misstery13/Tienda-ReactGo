// El paquete controllers también expone la red multinivel de referidos.
package controllers

import (
	"errors"
	"log"

	"github.com/gofiber/fiber/v2"

	"multicatalogo-backend/middleware"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// RedController recibe por inyección el repositorio con caché.
type RedController struct {
	repo *repository.CachedRepository
}

// NuevoRedController construye el controlador con su dependencia.
func NuevoRedController(repo *repository.CachedRepository) *RedController {
	return &RedController{repo: repo}
}

// GetRed devuelve el árbol completo de referidos con su jerarquía.
// El email sale del token firmado, nunca de un parámetro que el cliente controle:
// de lo contrario cualquiera podría pedir la red de otra cuenta.
func (rc *RedController) GetRed(c *fiber.Ctx) error {
	emailUsuario := middleware.EmailDeSesion(c)

	red, err := rc.repo.ObtenerRedDe(c.Context(), emailUsuario)
	if errors.Is(err, repository.ErrNoEncontrado) {
		return c.Status(fiber.StatusNotFound).JSON(models.NuevoAPIError(
			fiber.StatusNotFound,
			"La red de referidos está vacía",
			nil,
		))
	}
	if err != nil {
		log.Printf("error construyendo la red: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NuevoAPIError(
			fiber.StatusInternalServerError,
			"Error al consultar la red de referidos",
			nil,
		))
	}

	return c.JSON(red)
}

// MiRed devuelve el subárbol del usuario del token, no el árbol global.
// Cada distribuidor ve solo la red que cuelga de él.
func (rc *RedController) MiRed(c *fiber.Ctx) error {
	usuarioID := middleware.UsuarioIDDeSesion(c)

	red, err := rc.repo.ObtenerRedDeUsuario(c.Context(), usuarioID)
	if errors.Is(err, repository.ErrNoEncontrado) {
		return c.Status(fiber.StatusNotFound).JSON(models.NuevoAPIError(
			fiber.StatusNotFound,
			"Tu cuenta todavía no tiene un nodo en la red",
			nil,
		))
	}
	if err != nil {
		log.Printf("error construyendo la red del usuario: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NuevoAPIError(
			fiber.StatusInternalServerError,
			"Error al consultar tu red",
			nil,
		))
	}

	return c.JSON(red)
}

// MisComisiones devuelve el ledger del usuario con sus totales por estado.
func (rc *RedController) MisComisiones(c *fiber.Ctx) error {
	usuarioID := middleware.UsuarioIDDeSesion(c)

	resumen, err := rc.repo.ListarComisionesDeUsuario(c.Context(), usuarioID)
	if err != nil {
		log.Printf("error consultando comisiones: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NuevoAPIError(
			fiber.StatusInternalServerError,
			"Error al consultar tus comisiones",
			nil,
		))
	}

	return c.JSON(resumen)
}
