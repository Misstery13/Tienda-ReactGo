// Declaramos el paquete routes, cuya única responsabilidad es registrar los endpoints de la API.
package routes

import (
	// Importamos Fiber para poder recibir y configurar la instancia principal de la aplicación.
	"github.com/gofiber/fiber/v2"
	// Importamos nuestro paquete de controladores, donde residen las funciones que se ejecutarán en cada ruta.
	"multicatalogo-backend/controllers"
)

// Controladores agrupa las dependencias que main.go construye e inyecta.
type Controladores struct {
	Health    *controllers.HealthController
	Auth      *controllers.AuthController
	Productos *controllers.ProductoController
	Red       *controllers.RedController
	Checkout  *controllers.CheckoutController
}

// SetupRoutes recibe la aplicación Fiber y los controladores ya construidos.
func SetupRoutes(app *fiber.App, c Controladores) {
	// Estado del servicio.
	app.Get("/api/health", c.Health.HealthCheck)

	// Definimos una ruta HTTP POST en "/api/login" y la enlazamos a la función Login del paquete controllers.
	app.Post("/api/login", c.Auth.Login)

	// Definimos una ruta HTTP GET en "/api/productos" y la enlazamos a la función GetProductos del paquete controllers.
	app.Get("/api/productos", c.Productos.GetProductos)

	// Detalle de un producto por su identificador.
	app.Get("/api/productos/:id", c.Productos.GetProductoPorID)

	// Árbol multinivel de referidos.
	app.Get("/api/red", c.Red.GetRed)

	// Procesamiento transaccional de la compra.
	app.Post("/api/checkout", c.Checkout.Checkout)
}
