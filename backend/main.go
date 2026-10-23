// Declaramos que este es el paquete principal (main), es decir, el archivo ejecutable que inicia el programa.
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	// Importamos el framework principal Fiber.
	"github.com/gofiber/fiber/v2"
	// Importamos el middleware CORS para gestionar la seguridad entre distintos puertos/dominios.
	"github.com/gofiber/fiber/v2/middleware/cors"
	// Configuración de infraestructura (variables de entorno y conexiones).
	"multicatalogo-backend/config"
	// Controladores y capa de datos que se inyectan al arrancar.
	"multicatalogo-backend/controllers"
	"multicatalogo-backend/repository"
	// Importamos nuestro propio paquete de rutas para delegarle la configuración de los endpoints.
	"multicatalogo-backend/routes"
)

// func main es el punto de entrada de la aplicación en Go. Todo comienza a ejecutarse aquí.
func main() {
	// Variables de entorno (.env) antes de cualquier conexión.
	config.CargarEnv()

	// Sin base de datos la API no puede responder nada útil: se aborta el arranque.
	pool, err := config.ConectarDB()
	if err != nil {
		log.Fatalf("Error de inicio: %v", err)
	}
	// El pool se cierra al terminar main.
	defer pool.Close()

	// Redis es opcional: si no responde, la API sigue sirviendo desde PostgreSQL.
	clienteRedis := config.ConectarRedis()
	if clienteRedis != nil {
		defer clienteRedis.Close()
	}

	// ===== Inyección de dependencias =====
	// Se construyen aquí, una sola vez, y se pasan hacia abajo.
	var cache repository.CacheRepository = repository.SinCache{}
	if clienteRedis != nil {
		cache = repository.NuevoRedisCache(clienteRedis)
	}

	baseDatos := repository.NuevoPostgresRepository(pool)          // habla con PostgreSQL
	conCache := repository.NuevoCachedRepository(baseDatos, cache) // le agrega Cache-Aside

	// El secreto de firma nunca va en el código: sale del entorno.
	secretoJWT := os.Getenv("JWT_SECRET")
	if secretoJWT == "" {
		log.Fatal("Error de inicio: falta la variable JWT_SECRET")
	}

	controladores := routes.Controladores{
		Health:    controllers.NuevoHealthController(conCache, cache),
		Auth:      controllers.NuevoAuthController(conCache, secretoJWT),
		Productos: controllers.NuevoProductoController(conCache, cache),
		Red:       controllers.NuevoRedController(conCache),
		Checkout:  controllers.NuevoCheckoutController(conCache, cache),
		Usuarios:  controllers.NuevoUsuarioController(conCache, secretoJWT),
		Pedidos:   controllers.NuevoPedidoController(conCache),
		Admin:     controllers.NuevoAdminController(conCache),
	}

	// Instanciamos una nueva aplicación de Fiber y la guardamos en la variable 'app'.
	app := fiber.New()

	// Implementamos el middleware CORS a nivel global usando app.Use() para interceptar todas las peticiones entrantes.
	// Los orígenes permitidos se configuran por entorno (CORS_ORIGINS).
	origenes := os.Getenv("CORS_ORIGINS")
	if origenes == "" {
		origenes = "http://localhost:5173, http://127.0.0.1:5173, http://localhost:5174"
	}

	app.Use(cors.New(cors.Config{
		// Solo el frontend declarado puede llamar a esta API desde el navegador.
		AllowOrigins: origenes,
		// Authorization es necesaria para enviar el token JWT.
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Las rutas reciben los controladores y el middleware de autenticación.
	routes.SetupRoutes(app, controladores, secretoJWT)

	// Ctrl+C debe cerrar el pool de conexiones antes de salir.
	cerrar := make(chan os.Signal, 1)
	signal.Notify(cerrar, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-cerrar
		log.Println("Apagando el servidor...")
		if err := app.Shutdown(); err != nil {
			log.Printf("error al apagar: %v", err)
		}
	}()

	// Ponemos a la aplicación a escuchar peticiones en el puerto configurado.
	puerto := os.Getenv("PORT")
	if puerto == "" {
		puerto = os.Getenv("APP_PORT")
	}
	if puerto == "" {
		puerto = "3000"
	}
	if err := app.Listen(":" + puerto); err != nil {
		log.Printf("servidor detenido: %v", err)
	}
}

// // main.go
// package main

// import (
// 	"github.com/gofiber/fiber/v2"
// 	"github.com/gofiber/fiber/v2/middleware/cors"
// )

// // Estructuras de datos
// type LoginRequest struct {
// 	Email    string `json:"email"`
// 	Password string `json:"password"`
// }

// type Producto struct {
// 	ID     int     `json:"id"`
// 	Nombre string  `json:"nombre"`
// 	Precio float64 `json:"precio"`
// 	Img    string  `json:"img"`
// }

// func main() {
// 	app := fiber.New()

// 	// Middleware CORS para permitir peticiones desde React (Vite corre en el puerto 5173)
// 	app.Use(cors.New(cors.Config{
// 		AllowOrigins: "http://localhost:5173",
// 		AllowHeaders: "Origin, Content-Type, Accept",
// 	}))

// 	// Endpoint para Login
// 	app.Post("/api/login", func(c *fiber.Ctx) error {
// 		var req LoginRequest
// 		if err := c.BodyParser(&req); err != nil {
// 			return c.Status(400).JSON(fiber.Map{"error": "Cuerpo de petición inválido"})
// 		}

// 		if req.Email == "admin@upse.edu.ec" && req.Password == "123456" {
// 			return c.JSON(fiber.Map{"token": "fake-jwt-token-123", "email": req.Email})
// 		}
// 		return c.Status(401).JSON(fiber.Map{"error": "Credenciales incorrectas"})
// 	})

// 	// Endpoint para obtener Productos
// 	app.Get("/api/productos", func(c *fiber.Ctx) error {
// 		productos := []Producto{
// 			{ID: 1, Nombre: "Serum Revitalizante", Precio: 45.00, Img: "https://picsum.photos/seed/serum/150"},
// 			{ID: 2, Nombre: "Crema Hidratante Pro", Precio: 32.50, Img: "https://picsum.photos/seed/crema/150"},
// 			{ID: 3, Nombre: "Tónico Purificante", Precio: 28.00, Img: "https://picsum.photos/seed/tonico/150"},
// 			{ID: 4, Nombre: "Mascarilla Nocturna", Precio: 50.00, Img: "https://picsum.photos/seed/mascarilla/150"},
// 		}
// 		return c.JSON(productos)
// 	})

// 	app.Listen(":3000")
// }
