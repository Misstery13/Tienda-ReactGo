// El paquete controllers también expone el estado del servicio.
package controllers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"

	"multicatalogo-backend/repository"
)

// HealthController verifica de verdad sus dependencias.
type HealthController struct {
	repo  repository.DBRepository
	cache repository.CacheRepository
}

// NuevoHealthController construye el controlador con sus dependencias.
func NuevoHealthController(repo repository.DBRepository, cache repository.CacheRepository) *HealthController {
	return &HealthController{repo: repo, cache: cache}
}

// HealthCheck informa el estado real del servidor y de sus dependencias.
// Sin base de datos responde 503: decir "ok" cuando PostgreSQL está caído
// es peor que no tener health check.
func (hc *HealthController) HealthCheck(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 3*time.Second)
	defer cancel()

	estadoBD := "ok"
	if err := hc.repo.Ping(ctx); err != nil {
		estadoBD = "caída"
	}

	estadoCache := "ok"
	if !hc.cache.Activa() {
		estadoCache = "no disponible"
	}

	// Sin base de datos el servicio no puede operar.
	if estadoBD != "ok" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status":     "unavailable",
			"message":    "Servidor Go/Fiber sin acceso a PostgreSQL",
			"postgresql": estadoBD,
			"redis":      estadoCache,
		})
	}

	// Sin caché funciona, pero más lento: se informa como degradado.
	estado, mensaje := "ok", "Servidor Go/Fiber operativo"
	if estadoCache != "ok" {
		estado, mensaje = "degraded", "Servidor operativo sin caché: todas las lecturas van a PostgreSQL"
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":     estado,
		"message":    mensaje,
		"postgresql": estadoBD,
		"redis":      estadoCache,
	})
}
