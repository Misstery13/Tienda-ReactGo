// Consulta y gestión de pedidos.
package controllers

import (
	"errors"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"multicatalogo-backend/domain"
	"multicatalogo-backend/middleware"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// PedidoController recibe el repositorio por inyección.
type PedidoController struct {
	repo repository.DBRepository
}

// NuevoPedidoController construye el controlador.
func NuevoPedidoController(repo repository.DBRepository) *PedidoController {
	return &PedidoController{repo: repo}
}

// paginacion lee ?page y ?limit con valores por defecto seguros.
func paginacion(c *fiber.Ctx) (int, int) {
	pagina := max(c.QueryInt("page", 1), 1)
	limite := c.QueryInt("limit", 10)
	if limite < 1 || limite > 100 {
		limite = 10
	}
	return pagina, limite
}

// Mios devuelve los pedidos del usuario del token.
func (pc *PedidoController) Mios(c *fiber.Ctx) error {
	usuarioID := middleware.UsuarioIDDeSesion(c)
	pagina, limite := paginacion(c)

	resultado, err := pc.repo.ListarPedidosDeUsuario(c.Context(), usuarioID, pagina, limite)
	if err != nil {
		log.Printf("error listando mis pedidos: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NuevoAPIError(
			fiber.StatusInternalServerError, "Error al consultar tus pedidos", nil,
		))
	}

	return c.JSON(resultado)
}

// Listar devuelve todos los pedidos con filtro opcional ?estado=. Solo admin.
func (pc *PedidoController) Listar(c *fiber.Ctx) error {
	estado := c.Query("estado")
	if estado != "" && !domain.EstadoValido(estado) {
		return c.Status(fiber.StatusBadRequest).JSON(models.NuevoAPIError(
			fiber.StatusBadRequest, "El estado indicado no existe",
			fiber.Map{"valor_recibido": estado},
		))
	}

	pagina, limite := paginacion(c)

	resultado, err := pc.repo.ListarPedidos(c.Context(), estado, pagina, limite)
	if err != nil {
		log.Printf("error listando pedidos: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NuevoAPIError(
			fiber.StatusInternalServerError, "Error al consultar los pedidos", nil,
		))
	}

	return c.JSON(resultado)
}

// Detalle entrega un pedido con sus artículos.
// Un cliente solo puede ver los suyos; el administrador, cualquiera.
func (pc *PedidoController) Detalle(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NuevoAPIError(
			fiber.StatusBadRequest, "El id del pedido debe ser un número entero",
			fiber.Map{"valor_recibido": c.Params("id")},
		))
	}

	// El filtro por dueño solo aplica si quien pregunta no es administrador.
	filtroUsuario := 0
	if middleware.RolDeSesion(c) != "admin" {
		filtroUsuario = middleware.UsuarioIDDeSesion(c)
	}

	pedido, err := pc.repo.ObtenerPedido(c.Context(), id, filtroUsuario)
	switch {
	case errors.Is(err, repository.ErrNoEncontrado):
		return c.Status(fiber.StatusNotFound).JSON(models.NuevoAPIError(
			fiber.StatusNotFound, "Pedido no encontrado", fiber.Map{"id": id},
		))
	case errors.Is(err, repository.ErrPedidoAjeno):
		// Se responde 404 y no 403 para no revelar que el pedido existe.
		return c.Status(fiber.StatusNotFound).JSON(models.NuevoAPIError(
			fiber.StatusNotFound, "Pedido no encontrado", fiber.Map{"id": id},
		))
	case err != nil:
		log.Printf("error consultando pedido %d: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NuevoAPIError(
			fiber.StatusInternalServerError, "Error al consultar el pedido", nil,
		))
	}

	return c.JSON(pedido)
}

// CambiarEstado avanza o cancela un pedido. Solo admin.
func (pc *PedidoController) CambiarEstado(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NuevoAPIError(
			fiber.StatusBadRequest, "El id del pedido debe ser un número entero",
			fiber.Map{"valor_recibido": c.Params("id")},
		))
	}

	var req models.CambiarEstadoRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NuevoAPIError(
			fiber.StatusBadRequest, "Cuerpo de petición inválido", err.Error(),
		))
	}

	pedido, err := pc.repo.CambiarEstadoPedido(c.Context(), id, req.Estado)
	switch {
	case errors.Is(err, repository.ErrEstadoInvalido), errors.Is(err, repository.ErrTransicionInvalida):
		return c.Status(fiber.StatusBadRequest).JSON(models.NuevoAPIError(
			fiber.StatusBadRequest, err.Error(), nil,
		))
	case errors.Is(err, repository.ErrNoEncontrado):
		return c.Status(fiber.StatusNotFound).JSON(models.NuevoAPIError(
			fiber.StatusNotFound, "Pedido no encontrado", fiber.Map{"id": id},
		))
	case err != nil:
		log.Printf("error cambiando el estado del pedido %d: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NuevoAPIError(
			fiber.StatusInternalServerError, "No se pudo cambiar el estado", nil,
		))
	}

	return c.JSON(pedido)
}
