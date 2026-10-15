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

// ProductoController recibe el repositorio por inyección de dependencias.
type ProductoController struct {
	repo repository.DBRepository
}

// NuevoProductoController construye el controlador con su dependencia.
func NuevoProductoController(repo repository.DBRepository) *ProductoController {
	return &ProductoController{repo: repo}
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
