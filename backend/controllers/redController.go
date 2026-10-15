// El paquete controllers también expone la red multinivel de referidos.
package controllers

import (
	"errors"
	"log"

	"github.com/gofiber/fiber/v2"

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
// El email identifica de quién es la red y forma la clave de caché red:tree:{email}.
func (rc *RedController) GetRed(c *fiber.Ctx) error {
	emailUsuario := c.Query("email", "default")

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
