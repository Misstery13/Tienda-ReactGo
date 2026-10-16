// Microservicio de checkout: recibe el carrito, lo procesa en una transacción
// y lanza el cálculo de comisiones en segundo plano.
package controllers

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"multicatalogo-backend/middleware"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// CheckoutController recibe sus dos dependencias por inyección.
type CheckoutController struct {
	repo  repository.DBRepository
	cache repository.CacheRepository
}

// NuevoCheckoutController construye el controlador con sus dependencias.
func NuevoCheckoutController(repo repository.DBRepository, cache repository.CacheRepository) *CheckoutController {
	return &CheckoutController{repo: repo, cache: cache}
}

// Checkout procesa la compra enviada por el frontend.
func (cc *CheckoutController) Checkout(c *fiber.Ctx) error {
	var req models.CheckoutRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NuevoAPIError(
			fiber.StatusBadRequest,
			"Cuerpo de petición inválido",
			err.Error(),
		))
	}

	// La identidad la pone el token, no el cuerpo de la petición: así nadie
	// puede atribuir una compra (ni sus comisiones) a otra cuenta.
	req.Email = middleware.EmailDeSesion(c)

	// Validación de los datos de envío antes de abrir la transacción.
	faltantes := []string{}
	campos := map[string]string{
		"nombre":     req.Nombre,
		"email":      req.Email,
		"direccion":  req.Direccion,
		"ciudad":     req.Ciudad,
		"telefono":   req.Telefono,
		"metodoPago": req.MetodoPago,
	}
	for campo, valor := range campos {
		if strings.TrimSpace(valor) == "" {
			faltantes = append(faltantes, campo)
		}
	}
	if len(faltantes) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(models.NuevoAPIError(
			fiber.StatusBadRequest,
			"Faltan datos obligatorios del pedido",
			fiber.Map{"campos_faltantes": faltantes},
		))
	}

	// Toda la escritura ocurre dentro de una transacción ACID.
	pedido, err := cc.repo.CrearPedido(c.Context(), req)

	switch {
	case errors.Is(err, repository.ErrCarritoVacio),
		errors.Is(err, repository.ErrProductoInexistente),
		errors.Is(err, repository.ErrCantidadInvalida),
		errors.Is(err, repository.ErrStockInsuficiente):
		// Datos del cliente incorrectos: la transacción ya se revirtió.
		return c.Status(fiber.StatusBadRequest).JSON(models.NuevoAPIError(
			fiber.StatusBadRequest,
			err.Error(),
			nil,
		))
	case err != nil:
		// Fallo del sistema: el ROLLBACK dejó la base sin rastro del pedido.
		log.Printf("checkout revertido: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NuevoAPIError(
			fiber.StatusInternalServerError,
			"No se pudo procesar el pedido; no se realizó ningún cargo",
			nil,
		))
	}

	// El COMMIT fue exitoso: las comisiones se calculan en segundo plano
	// para no hacer esperar al cliente en la pantalla de pago.
	go cc.procesarComisiones(pedido)

	return c.Status(fiber.StatusCreated).JSON(pedido)
}

// procesarComisiones corre en una goroutine, fuera del ciclo de la petición.
// Por eso usa su propio contexto: el de Fiber se recicla al responder.
func (cc *CheckoutController) procesarComisiones(pedido models.Pedido) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := cc.repo.AcreditarComisiones(ctx, pedido.Cliente.Email, pedido.Total); err != nil {
		log.Printf("error acreditando comisiones del pedido %s: %v", pedido.Numero, err)
		return
	}

	// Los KPIs cambiaron: la caché del árbol queda obsoleta.
	// Se borran todas las claves porque cualquier usuario ve la misma red.
	cc.cache.EliminarPatron(ctx, repository.ClaveRedPrefix+"*")

	log.Printf("comisiones acreditadas para el pedido %s (total %.2f)", pedido.Numero, pedido.Total)
}
