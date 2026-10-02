// Declaramos el paquete models para agrupar las estructuras de datos de la aplicación.
package models

// LoginRequest define la estructura esperada para el cuerpo de la petición (JSON) al iniciar sesión.
type LoginRequest struct {
	// Email representa el correo del usuario; la etiqueta `json:"email"` indica cómo se mapea el JSON a esta variable.
	Email string `json:"email"`
	// Password representa la contraseña del usuario; se extrae del campo "password" del JSON entrante.
	Password string `json:"password"`
}

// LoginResponse es la respuesta de un inicio de sesión correcto.
// Mantiene los mismos campos JSON que el frontend ya consume: token, email y rol.
type LoginResponse struct {
	Token string `json:"token"`
	Email string `json:"email"`
	Rol   string `json:"rol"`
}

// Usuario representa una cuenta del sistema. Password no se serializa a JSON ("-").
type Usuario struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"-"`
	Rol      string `json:"rol"`
	Token    string `json:"-"`
}

// Producto define la estructura de los datos de un artículo en nuestro catálogo.
type Producto struct {
	// ID es el identificador único numérico del producto.
	ID int `json:"id"`
	// Nombre es la descripción en texto del producto.
	Nombre string `json:"nombre"`
	// Precio es el costo del producto, almacenado como un número con decimales (float64).
	Precio float64 `json:"precio"`
	// Img es la URL o ruta que apunta a la fotografía o imagen del producto.
	Img string `json:"img"`
}

// APIError estandariza el formato de los errores HTTP que devuelve la API.
type APIError struct {
	// Status es el código HTTP asociado al error (400, 401, 404...).
	Status int `json:"status"`
	// Message es la descripción legible del problema.
	Message string `json:"message"`
	// Details es información adicional opcional (campos faltantes, valor recibido, etc.).
	Details interface{} `json:"details,omitempty"`
}

// NuevoAPIError construye un APIError listo para enviarse como respuesta.
func NuevoAPIError(status int, message string, details interface{}) APIError {
	return APIError{Status: status, Message: message, Details: details}
}
