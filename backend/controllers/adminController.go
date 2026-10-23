// Liquidaciones a distribuidores y reportes de negocio.
package controllers

import (
	"context"
	"errors"
	"log"

	"github.com/gofiber/fiber/v2"

	"multicatalogo-backend/middleware"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// AdminController agrupa liquidaciones y reportes.
type AdminController struct {
	repo repository.DBRepository
}

// NuevoAdminController construye el controlador.
func NuevoAdminController(repo repository.DBRepository) *AdminController {
	return &AdminController{repo: repo}
}

// Deuda lista cuánto se le debe a cada distribuidor. Solo admin.
func (ac *AdminController) Deuda(c *fiber.Ctx) error {
	deudas, err := ac.repo.DeudaPorDistribuidor(c.Context())
	if err != nil {
		log.Printf("error consultando la deuda: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NuevoAPIError(
			fiber.StatusInternalServerError, "Error al consultar la deuda", nil,
		))
	}
	return c.JSON(deudas)
}

// Liquidar paga las comisiones pendientes de un distribuidor. Solo admin.
func (ac *AdminController) Liquidar(c *fiber.Ctx) error {
	var req models.LiquidarRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NuevoAPIError(
			fiber.StatusBadRequest, "Cuerpo de petición inválido", err.Error(),
		))
	}

	if req.ReferidoID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(models.NuevoAPIError(
			fiber.StatusBadRequest, "Falta el referido a liquidar", nil,
		))
	}

	liq, err := ac.repo.Liquidar(c.Context(), req.ReferidoID)
	if errors.Is(err, repository.ErrSinDeuda) {
		// Liquidar dos veces no es un error del sistema: simplemente no hay deuda.
		return c.Status(fiber.StatusBadRequest).JSON(models.NuevoAPIError(
			fiber.StatusBadRequest, err.Error(), fiber.Map{"referido_id": req.ReferidoID},
		))
	}
	if err != nil {
		log.Printf("error liquidando al referido %d: %v", req.ReferidoID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NuevoAPIError(
			fiber.StatusInternalServerError, "No se pudo liquidar", nil,
		))
	}

	return c.Status(fiber.StatusCreated).JSON(liq)
}

// Liquidaciones devuelve el historial completo. Solo admin.
func (ac *AdminController) Liquidaciones(c *fiber.Ctx) error {
	lista, err := ac.repo.ListarLiquidaciones(c.Context(), 0)
	if err != nil {
		log.Printf("error listando liquidaciones: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NuevoAPIError(
			fiber.StatusInternalServerError, "Error al consultar las liquidaciones", nil,
		))
	}
	return c.JSON(lista)
}

// MisLiquidaciones devuelve los cobros del distribuidor del token.
func (ac *AdminController) MisLiquidaciones(c *fiber.Ctx) error {
	referidoID, err := ac.repo.ReferidoDeUsuario(c.Context(), middleware.UsuarioIDDeSesion(c))
	if err != nil {
		log.Printf("error buscando el nodo del usuario: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NuevoAPIError(
			fiber.StatusInternalServerError, "Error al consultar tus cobros", nil,
		))
	}

	// Sin nodo en la red no hay cobros, pero tampoco es un error.
	if referidoID == 0 {
		return c.JSON([]models.Liquidacion{})
	}

	lista, err := ac.repo.ListarLiquidaciones(c.Context(), referidoID)
	if err != nil {
		log.Printf("error listando mis liquidaciones: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NuevoAPIError(
			fiber.StatusInternalServerError, "Error al consultar tus cobros", nil,
		))
	}
	return c.JSON(lista)
}

// Resumen entrega los KPIs del periodo. Solo admin.
func (ac *AdminController) Resumen(c *fiber.Ctx) error {
	resumen, err := ac.repo.ResumenDeVentas(c.Context(), c.Query("desde"), c.Query("hasta"))
	if err != nil {
		log.Printf("error calculando el resumen: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NuevoAPIError(
			fiber.StatusInternalServerError, "Error al calcular el resumen", nil,
		))
	}
	return c.JSON(resumen)
}

// TopProductos entrega el ranking de productos por venta. Solo admin.
func (ac *AdminController) TopProductos(c *fiber.Ctx) error {
	return ac.ranking(c, ac.repo.TopProductos)
}

// TopReferidos entrega el ranking de distribuidores por comisión. Solo admin.
func (ac *AdminController) TopReferidos(c *fiber.Ctx) error {
	return ac.ranking(c, ac.repo.TopReferidos)
}

// ranking comparte el manejo de ?limit= y de errores de ambos rankings.
func (ac *AdminController) ranking(c *fiber.Ctx, consulta func(ctx context.Context, limite int) ([]models.FilaReporte, error)) error {
	limite := c.QueryInt("limit", 5)
	if limite < 1 || limite > 50 {
		limite = 5
	}

	filas, err := consulta(c.Context(), limite)
	if err != nil {
		log.Printf("error consultando el ranking: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NuevoAPIError(
			fiber.StatusInternalServerError, "Error al consultar el reporte", nil,
		))
	}
	return c.JSON(filas)
}
