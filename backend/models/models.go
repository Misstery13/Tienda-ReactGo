// Declaramos el paquete models para agrupar las estructuras de datos de la aplicación.
package models

import "os"

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
	ID         int    `json:"id"`
	Email      string `json:"email"`
	Password   string `json:"-"`
	Rol        string `json:"rol"`
	Token      string `json:"-"`
	Nombre     string `json:"nombre"`
	Activo     bool   `json:"activo"`
	ReferidoID *int   `json:"referido_id,omitempty"`
	CreadoEn   string `json:"creado_en,omitempty"`
}

// RegistroRequest es el cuerpo de POST /api/register.
// CodigoReferido es opcional: si viene, el nuevo usuario cuelga de ese nodo.
type RegistroRequest struct {
	Nombre         string `json:"nombre"`
	Email          string `json:"email"`
	Password       string `json:"password"`
	CodigoReferido string `json:"codigo_referido"`
}

// ActualizarUsuarioRequest permite al administrador cambiar rol y estado.
// Son punteros para distinguir "no enviado" de "enviado en falso".
type ActualizarUsuarioRequest struct {
	Rol    *string `json:"rol"`
	Activo *bool   `json:"activo"`
}

// Pagina envuelve cualquier listado paginado.
type Pagina[T any] struct {
	Datos  []T `json:"datos"`
	Total  int `json:"total"`
	Pagina int `json:"pagina"`
	Limite int `json:"limite"`
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
	// Descripcion es el texto comercial del producto.
	Descripcion string `json:"descripcion"`
	// Categoria agrupa los productos para los filtros del catálogo.
	Categoria string `json:"categoria"`
	// Galeria son las imágenes adicionales que se muestran en el detalle.
	Galeria []string `json:"galeria"`
	// Stock disponible; el checkout lo descuenta.
	Stock int `json:"stock"`
	// Activo en falso oculta el producto del catálogo sin borrar la fila.
	Activo bool `json:"activo"`
}

// ProductoRequest es el cuerpo para crear o editar un producto.
// Los punteros permiten editar solo algunos campos (PUT parcial).
type ProductoRequest struct {
	Nombre      *string   `json:"nombre"`
	Descripcion *string   `json:"descripcion"`
	Precio      *float64  `json:"precio"`
	Categoria   *string   `json:"categoria"`
	Img         *string   `json:"img"`
	Galeria     *[]string `json:"galeria"`
	Stock       *int      `json:"stock"`
	Activo      *bool     `json:"activo"`
}

// Referido es un nodo del árbol multinivel. Hijos contiene los referidos del
// siguiente nivel, lo que permite enviar la jerarquía completa en un solo JSON.
type Referido struct {
	ID     int         `json:"id"`
	Nombre string      `json:"nombre"`
	Nivel  int         `json:"nivel"`
	Ventas float64     `json:"ventas"`
	Hijos  []*Referido `json:"hijos,omitempty"`
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
// Details solo viaja al cliente en desarrollo (APP_ENV=dev): en producción
// podría filtrar detalles internos de la base de datos.
func NuevoAPIError(status int, message string, details interface{}) APIError {
	if os.Getenv("APP_ENV") != "dev" {
		details = nil
	}
	return APIError{Status: status, Message: message, Details: details}
}

// ============ FASE 4: CHECKOUT ============

// ItemCarrito es cada línea que envía el carrito del frontend.
// Solo viajan el id y la cantidad: el precio se toma de la base de datos
// para que el cliente no pueda manipularlo.
type ItemCarrito struct {
	ProductoID int `json:"producto_id"`
	Cantidad   int `json:"cantidad"`
}

// CheckoutRequest es el cuerpo que manda Checkout.tsx al confirmar la compra.
type CheckoutRequest struct {
	Nombre     string        `json:"nombre"`
	Email      string        `json:"email"`
	Direccion  string        `json:"direccion"`
	Ciudad     string        `json:"ciudad"`
	Telefono   string        `json:"telefono"`
	MetodoPago string        `json:"metodoPago"`
	Items      []ItemCarrito `json:"items"`
}

// ClientePedido son los datos del comprador que se muestran en la confirmación.
type ClientePedido struct {
	Nombre string `json:"nombre"`
	Email  string `json:"email"`
	Ciudad string `json:"ciudad"`
}

// LineaPedido es un item ya resuelto con el precio real de la base.
type LineaPedido struct {
	ProductoID int     `json:"producto_id"`
	Nombre     string  `json:"nombre"`
	Cantidad   int     `json:"cantidad"`
	Precio     float64 `json:"precio"`
	Subtotal   float64 `json:"subtotal"`
}

// Pedido es la respuesta de /api/checkout; su forma coincide con la que
// Confirmacion.tsx ya espera.
type Pedido struct {
	ID      int           `json:"id"`
	Numero  string        `json:"numero"`
	Cliente ClientePedido `json:"cliente"`
	Items   []LineaPedido `json:"items"`
	Total   float64       `json:"total"`
	Fecha   string        `json:"fecha"`
	Estado  string        `json:"estado,omitempty"`
}

// ResumenPedido es cada fila de los listados de pedidos (sin el detalle).
type ResumenPedido struct {
	ID        int     `json:"id"`
	Numero    string  `json:"numero"`
	Cliente   string  `json:"cliente"`
	Email     string  `json:"email"`
	Total     float64 `json:"total"`
	Estado    string  `json:"estado"`
	Fecha     string  `json:"fecha"`
	Articulos int     `json:"articulos"`
}

// CambiarEstadoRequest es el cuerpo de PATCH /api/pedidos/:id/estado.
type CambiarEstadoRequest struct {
	Estado string `json:"estado"`
}
