// Declaramos el paquete models para agrupar las estructuras de datos de la aplicación.
package models

import "time"

// APIError estandariza la estructura de los mensajes de error HTTP en la API.
type APIError struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Error implementa la interfaz error estándar de Go.
func (e APIError) Error() string {
	return e.Message
}

// LoginRequest define la estructura esperada para el cuerpo de la petición (JSON) al iniciar sesión.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse representa la respuesta retornada tras una autenticación exitosa.
type LoginResponse struct {
	Token string `json:"token"`
	Email string `json:"email"`
	Rol   string `json:"rol"`
}

// Usuario define la entidad de usuario mapeada con la tabla 'usuarios' en PostgreSQL.
type Usuario struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"-"`
	Rol      string `json:"rol"`
}

// Producto define la estructura de los datos de un artículo en nuestro catálogo relacional.
type Producto struct {
	ID          int      `json:"id"`
	Nombre      string   `json:"nombre"`
	Descripcion string   `json:"descripcion,omitempty"`
	Precio      float64  `json:"precio"`
	Categoria   string   `json:"categoria,omitempty"`
	Img         string   `json:"img"`
	Galeria     []string `json:"galeria,omitempty"`
}

// Referido representa un nodo en el árbol de la red multinivel de referidos.
type Referido struct {
	ID       int        `json:"id"`
	Nombre   string     `json:"nombre"`
	Nivel    int        `json:"nivel"`
	Ventas   float64    `json:"ventas"`
	ParentID *int       `json:"parent_id,omitempty"`
	Hijos    []Referido `json:"hijos,omitempty"`
}

// ItemCarrito representa un ítem comprado dentro del carrito de compras.
type ItemCarrito struct {
	ID       int     `json:"id,omitempty"`
	Nombre   string  `json:"nombre"`
	Cantidad int     `json:"cantidad"`
	Precio   float64 `json:"precio"`
}

// ClienteCheckout contiene los datos del comprador enviados desde Checkout.tsx.
type ClienteCheckout struct {
	Nombre    string `json:"nombre"`
	Email     string `json:"email"`
	Direccion string `json:"direccion,omitempty"`
	Ciudad    string `json:"ciudad,omitempty"`
	Telefono  string `json:"telefono,omitempty"`
}

// CheckoutRequest es el cuerpo JSON que recibe el endpoint POST /api/checkout.
type CheckoutRequest struct {
	Cliente    ClienteCheckout `json:"cliente"`
	MetodoPago string          `json:"metodoPago"`
	Items      []ItemCarrito   `json:"items"`
	Total      float64         `json:"total"`
}

// PedidoConfirmado representa la respuesta estructurada tras procesar la orden exitosamente.
// Su forma coincide con la estructura requerida por el componente Confirmacion.tsx.
type PedidoConfirmado struct {
	Numero  string          `json:"numero"`
	Items   []ItemCarrito   `json:"items"`
	Total   float64         `json:"total"`
	Cliente ClienteCheckout `json:"cliente"`
	Fecha   string          `json:"fecha"`
}

// Pedido representa la fila en la tabla relacional 'pedidos' de PostgreSQL.
type Pedido struct {
	ID               int       `json:"id"`
	Numero           string    `json:"numero"`
	ClienteNombre    string    `json:"cliente_nombre"`
	ClienteEmail     string    `json:"cliente_email"`
	ClienteCiudad    string    `json:"cliente_ciudad"`
	ClienteDireccion string    `json:"cliente_direccion"`
	ClienteTelefono  string    `json:"cliente_telefono"`
	MetodoPago       string    `json:"metodo_pago"`
	Total            float64   `json:"total"`
	Fecha            time.Time `json:"fecha"`
}