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
	BuscarUsuario(ctx context.Context, email, password string) (models.Usuario, error)
	ObtenerProductos(ctx context.Context) ([]models.Producto, error)
	ObtenerProductoPorID(ctx context.Context, id int) (models.Producto, error)
	ObtenerRed(ctx context.Context) (*models.Referido, error)
	CrearPedido(ctx context.Context, req models.CheckoutRequest) (models.Pedido, error)
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
