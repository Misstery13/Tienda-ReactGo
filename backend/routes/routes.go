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
	Pedidos   *controllers.PedidoController
	Admin     *controllers.AdminController
}

// SetupRoutes recibe la aplicación Fiber, los controladores y el secreto de firma.
func SetupRoutes(app *fiber.App, c Controladores, secretoJWT string) {
	// Middlewares reutilizables para las rutas privadas.
	requiereSesion := middleware.RequireAuth(secretoJWT)
	soloAdmin := middleware.RequireAdmin()

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

	// El árbol global es información de negocio: solo administradores.
	app.Get("/api/red", requiereSesion, soloAdmin, c.Red.GetRed)

	// Cada usuario ve su propia red y su propio ledger de comisiones.
	app.Get("/api/red/mia", requiereSesion, c.Red.MiRed)
	app.Get("/api/comisiones/mias", requiereSesion, c.Red.MisComisiones)

	// Procesamiento transaccional de la compra: cualquier usuario autenticado.
	app.Post("/api/checkout", requiereSesion, c.Checkout.Checkout)

	// Administración del catálogo: solo administradores.
	app.Get("/api/admin/productos", requiereSesion, soloAdmin, c.Productos.ListarAdmin)
	app.Post("/api/productos", requiereSesion, soloAdmin, c.Productos.Crear)
	app.Put("/api/productos/:id", requiereSesion, soloAdmin, c.Productos.Actualizar)
	app.Delete("/api/productos/:id", requiereSesion, soloAdmin, c.Productos.Desactivar)

	// Pedidos: el cliente ve los suyos, el administrador todos.
	app.Get("/api/pedidos/mios", requiereSesion, c.Pedidos.Mios)
	app.Get("/api/pedidos", requiereSesion, soloAdmin, c.Pedidos.Listar)
	app.Get("/api/pedidos/:id", requiereSesion, c.Pedidos.Detalle)
	app.Patch("/api/pedidos/:id/estado", requiereSesion, soloAdmin, c.Pedidos.CambiarEstado)

	// Liquidaciones a distribuidores.
	app.Get("/api/liquidaciones/mias", requiereSesion, c.Admin.MisLiquidaciones)
	app.Get("/api/liquidaciones", requiereSesion, soloAdmin, c.Admin.Liquidaciones)
	app.Post("/api/liquidaciones", requiereSesion, soloAdmin, c.Admin.Liquidar)
	app.Get("/api/comisiones/deuda", requiereSesion, soloAdmin, c.Admin.Deuda)

	// Reportes de negocio: solo administradores.
	app.Get("/api/reportes/resumen", requiereSesion, soloAdmin, c.Admin.Resumen)
	app.Get("/api/reportes/top-productos", requiereSesion, soloAdmin, c.Admin.TopProductos)
	app.Get("/api/reportes/top-referidos", requiereSesion, soloAdmin, c.Admin.TopReferidos)

	// Administración de cuentas: solo administradores.
	app.Get("/api/usuarios", requiereSesion, middleware.RequireAdmin(), c.Usuarios.Listar)
	app.Patch("/api/usuarios/:id", requiereSesion, middleware.RequireAdmin(), c.Usuarios.Actualizar)
}
