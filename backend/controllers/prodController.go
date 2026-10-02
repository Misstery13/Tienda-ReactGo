// Seguimos dentro del paquete controllers, ya que maneja la lógica de otra sección del negocio.
package controllers

import (
	"strconv"

	// Importamos Fiber para manejar la respuesta HTTP.
	"github.com/gofiber/fiber/v2"
	// Importamos nuestro paquete de modelos para usar la estructura APIError.
	"multicatalogo-backend/models"
	// El catálogo vive en el repositorio.
	"multicatalogo-backend/repository"
)

// GetProductos devuelve el catálogo completo de artículos.
func GetProductos(c *fiber.Ctx) error {
	// Fiber convierte automáticamente el slice de estructuras a formato JSON y lo envía como respuesta al cliente.
	return c.JSON(repository.ObtenerProductos())
}

// GetProductoPorID devuelve un solo producto a partir del :id de la ruta.
func GetProductoPorID(c *fiber.Ctx) error {
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

	// Un id válido pero inexistente no es un error de formato, sino un recurso no encontrado: HTTP 404.
	producto, encontrado := repository.ObtenerProductoPorID(id)
	if !encontrado {
		return c.Status(fiber.StatusNotFound).JSON(models.NuevoAPIError(
			fiber.StatusNotFound,
			"Producto no encontrado",
			fiber.Map{"id": id},
		))
	}

	return c.JSON(producto)
}
