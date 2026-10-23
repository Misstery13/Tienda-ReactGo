// Interfaces de la capa de datos.
// Los controladores dependen de estas abstracciones, no de PostgreSQL ni de Redis,
// así se pueden sustituir por dobles de prueba sin tocar la lógica HTTP.
package repository

import (
	"context"
	"time"

	"multicatalogo-backend/models"
)

// DBRepository es todo lo que la aplicación necesita de un almacén de datos.
type DBRepository interface {
	BuscarUsuarioPorEmail(ctx context.Context, email string) (models.Usuario, error)
	CrearUsuario(ctx context.Context, datos models.RegistroRequest, hash string) (models.Usuario, error)
	ListarUsuarios(ctx context.Context, pagina, limite int) (models.Pagina[models.Usuario], error)
	ActualizarUsuario(ctx context.Context, id int, datos models.ActualizarUsuarioRequest) (models.Usuario, error)
	Ping(ctx context.Context) error
	ObtenerProductos(ctx context.Context) ([]models.Producto, error)
	ObtenerProductoPorID(ctx context.Context, id int) (models.Producto, error)
	ObtenerRed(ctx context.Context) (*models.Referido, error)
	CrearProducto(ctx context.Context, datos models.ProductoRequest) (models.Producto, error)
	ActualizarProducto(ctx context.Context, id int, datos models.ProductoRequest) (models.Producto, error)
	DesactivarProducto(ctx context.Context, id int) error
	ListarProductosAdmin(ctx context.Context, q string, pagina, limite int) (models.Pagina[models.Producto], error)
	CrearPedido(ctx context.Context, req models.CheckoutRequest) (models.Pedido, error)
	ListarPedidosDeUsuario(ctx context.Context, usuarioID, pagina, limite int) (models.Pagina[models.ResumenPedido], error)
	ListarPedidos(ctx context.Context, estado string, pagina, limite int) (models.Pagina[models.ResumenPedido], error)
	ObtenerPedido(ctx context.Context, id, usuarioID int) (models.Pedido, error)
	CambiarEstadoPedido(ctx context.Context, id int, nuevo string) (models.Pedido, error)
	AcreditarComisiones(ctx context.Context, email string, total float64) error
}

// CacheRepository abstrae el almacén de caché.
type CacheRepository interface {
	Obtener(ctx context.Context, clave string, destino any) bool
	Guardar(ctx context.Context, clave string, valor any, ttl time.Duration)
	Eliminar(ctx context.Context, claves ...string)
	EliminarPatron(ctx context.Context, patron string)
	Activa() bool
}
