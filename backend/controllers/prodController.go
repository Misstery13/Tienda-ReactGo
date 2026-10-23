// Seguimos dentro del paquete controllers, ya que maneja la lógica de otra sección del negocio.
package controllers

import (
	"errors"
	"log"
	"strconv"

	// Importamos Fiber para manejar la respuesta HTTP.
	"github.com/gofiber/fiber/v2"
	// Importamos nuestro paquete de modelos para usar la estructura APIError.
	"multicatalogo-backend/models"
	// El catálogo vive en el repositorio.
	"multicatalogo-backend/repository"
)

// ProductoController recibe el repositorio y la caché por inyección.
type ProductoController struct {
	repo  repository.DBRepository
	cache repository.CacheRepository
}

// NuevoProductoController construye el controlador con sus dependencias.
func NuevoProductoController(repo repository.DBRepository, cache repository.CacheRepository) *ProductoController {
	return &ProductoController{repo: repo, cache: cache}
}

// invalidarCatalogo borra la caché del catálogo tras cada escritura.
// Sin esto el cliente seguiría viendo la lista vieja hasta que expire el TTL.
func (pc *ProductoController) invalidarCatalogo(c *fiber.Ctx) {
	pc.cache.Eliminar(c.Context(), repository.ClaveProductos)
}

// errorDeValidacion traduce los errores de negocio del catálogo a HTTP 400.
func errorDeValidacion(err error) bool {
	return errors.Is(err, repository.ErrNombreVacio) ||
		errors.Is(err, repository.ErrPrecioInvalido) ||
		errors.Is(err, repository.ErrCategoriaVacia) ||
		errors.Is(err, repository.ErrStockNegativo)
}

// Crear da de alta un producto. Solo administradores.
func (pc *ProductoController) Crear(c *fiber.Ctx) error {
	var req models.ProductoRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NuevoAPIError(
			fiber.StatusBadRequest, "Cuerpo de petición inválido", err.Error(),
		))
	}

	producto, err := pc.repo.CrearProducto(c.Context(), req)
	if errorDeValidacion(err) {
		return c.Status(fiber.StatusBadRequest).JSON(models.NuevoAPIError(
			fiber.StatusBadRequest, err.Error(), nil,
		))
	}
	if err != nil {
		log.Printf("error creando producto: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NuevoAPIError(
			fiber.StatusInternalServerError, "No se pudo crear el producto", nil,
		))
	}

	pc.invalidarCatalogo(c)
	return c.Status(fiber.StatusCreated).JSON(producto)
}

// Actualizar edita un producto existente. Solo administradores.
func (pc *ProductoController) Actualizar(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NuevoAPIError(
			fiber.StatusBadRequest, "El id del producto debe ser un número entero",
			fiber.Map{"valor_recibido": c.Params("id")},
		))
	}

	var req models.ProductoRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NuevoAPIError(
			fiber.StatusBadRequest, "Cuerpo de petición inválido", err.Error(),
		))
	}

	producto, err := pc.repo.ActualizarProducto(c.Context(), id, req)
	switch {
	case errorDeValidacion(err):
		return c.Status(fiber.StatusBadRequest).JSON(models.NuevoAPIError(
			fiber.StatusBadRequest, err.Error(), nil,
		))
	case errors.Is(err, repository.ErrNoEncontrado):
		return c.Status(fiber.StatusNotFound).JSON(models.NuevoAPIError(
			fiber.StatusNotFound, "Producto no encontrado", fiber.Map{"id": id},
		))
	case err != nil:
		log.Printf("error actualizando producto %d: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NuevoAPIError(
			fiber.StatusInternalServerError, "No se pudo actualizar el producto", nil,
		))
	}

	pc.invalidarCatalogo(c)
	return c.JSON(producto)
}

// Desactivar aplica el borrado lógico. Solo administradores.
func (pc *ProductoController) Desactivar(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NuevoAPIError(
			fiber.StatusBadRequest, "El id del producto debe ser un número entero",
			fiber.Map{"valor_recibido": c.Params("id")},
		))
	}

	err = pc.repo.DesactivarProducto(c.Context(), id)
	if errors.Is(err, repository.ErrNoEncontrado) {
		return c.Status(fiber.StatusNotFound).JSON(models.NuevoAPIError(
			fiber.StatusNotFound, "Producto no encontrado", fiber.Map{"id": id},
		))
	}
	if err != nil {
		log.Printf("error desactivando producto %d: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NuevoAPIError(
			fiber.StatusInternalServerError, "No se pudo desactivar el producto", nil,
		))
	}

	pc.invalidarCatalogo(c)
	return c.JSON(fiber.Map{"id": id, "activo": false, "message": "Producto desactivado"})
}

// ListarAdmin incluye los productos inactivos y acepta búsqueda por ?q=.
func (pc *ProductoController) ListarAdmin(c *fiber.Ctx) error {
	pagina := max(c.QueryInt("page", 1), 1)
	limite := c.QueryInt("limit", 20)
	if limite < 1 || limite > 100 {
		limite = 20
	}

	resultado, err := pc.repo.ListarProductosAdmin(c.Context(), c.Query("q"), pagina, limite)
	if err != nil {
		log.Printf("error listando productos (admin): %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NuevoAPIError(
			fiber.StatusInternalServerError, "Error al consultar el catálogo", nil,
		))
	}

	return c.JSON(resultado)
}

// GetProductos devuelve el catálogo completo de artículos.
func (pc *ProductoController) GetProductos(c *fiber.Ctx) error {
	productos, err := pc.repo.ObtenerProductos(c.Context())
	if err != nil {
		log.Printf("error listando productos: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NuevoAPIError(
			fiber.StatusInternalServerError,
			"Error al consultar el catálogo",
			nil,
		))
	}

	// Fiber convierte automáticamente el slice de estructuras a formato JSON y lo envía como respuesta al cliente.
	return c.JSON(productos)
}

// GetProductoPorID devuelve un solo producto a partir del :id de la ruta.
func (pc *ProductoController) GetProductoPorID(c *fiber.Ctx) error {
	// El parámetro de la URL llega siempre como texto: hay que convertirlo y controlar el error.
	idParam := c.Params("id")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		// El cliente envió texto en lugar de un número: HTTP 400.
		return c.Status(fiber.StatusBadRequest).JSON(models.NuevoAPIError(
			fiber.StatusBadRequest,
			"El id del producto debe ser un número entero",
			fiber.Map{"valor_recibido": idParam},
		))
	}

	producto, err := pc.repo.ObtenerProductoPorID(c.Context(), id)
	if errors.Is(err, repository.ErrNoEncontrado) {
		// Un id válido pero inexistente no es un error de formato, sino un recurso no encontrado: HTTP 404.
		return c.Status(fiber.StatusNotFound).JSON(models.NuevoAPIError(
			fiber.StatusNotFound,
			"Producto no encontrado",
			fiber.Map{"id": id},
		))
	}
	if err != nil {
		log.Printf("error consultando producto %d: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NuevoAPIError(
			fiber.StatusInternalServerError,
			"Error al consultar el producto",
			nil,
		))
	}

	return c.JSON(producto)
}
