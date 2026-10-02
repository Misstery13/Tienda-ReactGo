// Paquete repository: única fuente de datos de la aplicación.
// Por ahora los datos son estáticos (en memoria); en la siguiente fase estas
// funciones se reemplazarán por consultas a PostgreSQL sin tocar los controladores.
package repository

import (
	"strings"

	"multicatalogo-backend/models"
)

// Usuarios registrados del sistema.
var usuarios = []models.Usuario{
	{ID: 1, Email: "admin@upse.edu.ec", Password: "123456", Rol: "admin", Token: "fake-jwt-token-123"},
	{ID: 2, Email: "cliente@upse.edu.ec", Password: "123456", Rol: "cliente", Token: "fake-jwt-token-456"},
}

// Catálogo de productos.
var productos = []models.Producto{
	{ID: 1, Nombre: "Serum Revitalizante", Precio: 45.00, Img: "https://picsum.photos/seed/serum/150"},
	{ID: 2, Nombre: "Crema Hidratante Pro", Precio: 32.50, Img: "https://picsum.photos/seed/crema/150"},
	{ID: 3, Nombre: "Tónico Purificante", Precio: 28.00, Img: "https://picsum.photos/seed/tonico/150"},
	{ID: 4, Nombre: "Mascarilla Nocturna", Precio: 50.00, Img: "https://picsum.photos/seed/mascarilla/150"},
}

// BuscarUsuario devuelve el usuario cuyas credenciales coinciden.
// El segundo valor indica si hubo coincidencia. El correo se compara sin
// distinguir mayúsculas; la contraseña sí es sensible a mayúsculas.
func BuscarUsuario(email, password string) (models.Usuario, bool) {
	for _, u := range usuarios {
		if strings.EqualFold(u.Email, email) && u.Password == password {
			return u, true
		}
	}
	return models.Usuario{}, false
}

// ObtenerProductos devuelve el catálogo completo.
func ObtenerProductos() []models.Producto {
	return productos
}

// ObtenerProductoPorID busca un producto por su identificador.
// El segundo valor indica si el producto existe.
func ObtenerProductoPorID(id int) (models.Producto, bool) {
	for _, p := range productos {
		if p.ID == id {
			return p, true
		}
	}
	return models.Producto{}, false
}
