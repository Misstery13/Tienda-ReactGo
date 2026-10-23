// Declaramos el paquete routes, cuya única responsabilidad es registrar los endpoints de la API.
package routes

import (
	// Importamos Fiber para poder recibir y configurar la instancia principal de la aplicación.
	"github.com/gofiber/fiber/v2"
	// Importamos nuestro paquete de controladores, donde residen las funciones que se ejecutarán en cada ruta.
	"multicatalogo-backend/controllers"
	// El middleware protege las rutas privadas.
	"multicatalogo-backend/middleware"
)

// Controladores agrupa las dependencias que main.go construye e inyecta.
type Controladores struct {
	Health    *controllers.HealthController
	Auth      *controllers.AuthController
	Productos *controllers.ProductoController
	Red       *controllers.RedController
	Checkout  *controllers.CheckoutController
	Usuarios  *controllers.UsuarioController
}

// SetupRoutes recibe la aplicación Fiber, los controladores y el secreto de firma.
func SetupRoutes(app *fiber.App, c Controladores, secretoJWT string) {
	// Middleware reutilizable para las rutas que exigen sesión iniciada.
	requiereSesion := middleware.RequireAuth(secretoJWT)

	// Estado del servicio.
	app.Get("/api/health", c.Health.HealthCheck)

	// Definimos una ruta HTTP POST en "/api/login" y la enlazamos a la función Login del paquete controllers.
	app.Post("/api/login", c.Auth.Login)

	// Registro público de clientes (acepta ?ref=CODIGO de invitación).
	app.Post("/api/register", c.Usuarios.Registrar)

	// Definimos una ruta HTTP GET en "/api/productos" y la enlazamos a la función GetProductos del paquete controllers.
	app.Get("/api/productos", c.Productos.GetProductos)

	// Detalle de un producto por su identificador.
	app.Get("/api/productos/:id", c.Productos.GetProductoPorID)

	// ===== Rutas protegidas: exigen un JWT válido =====

	// El árbol multinivel es información de negocio: solo administradores.
	app.Get("/api/red", requiereSesion, middleware.RequireAdmin(), c.Red.GetRed)

	// Procesamiento transaccional de la compra: cualquier usuario autenticado.
	app.Post("/api/checkout", requiereSesion, c.Checkout.Checkout)

	// Administración de cuentas: solo administradores.
	app.Get("/api/usuarios", requiereSesion, middleware.RequireAdmin(), c.Usuarios.Listar)
	app.Patch("/api/usuarios/:id", requiereSesion, middleware.RequireAdmin(), c.Usuarios.Actualizar)
}
