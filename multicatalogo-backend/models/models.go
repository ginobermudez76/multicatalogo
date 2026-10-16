// Declaramos el paquete models para agrupar las estructuras de datos de la aplicación.
package models

// APIError estandariza la estructura de los mensajes de error HTTP en la API.
// Contiene Status (código HTTP numérico), Message (descripción general legible) y Details (información técnica u opcional).
type APIError struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Error implementa la interfaz error estándar de Go para permitir usar APIError como error nativo.
func (e APIError) Error() string {
	return e.Message
}

// LoginRequest define la estructura esperada para el cuerpo de la petición (JSON) al iniciar sesión.
type LoginRequest struct {
	// Email representa el correo del usuario; la etiqueta `json:"email"` indica cómo se mapea el JSON a esta variable.
	Email string `json:"email"`
	// Password representa la contraseña del usuario; se extrae del campo "password" del JSON entrante.
	Password string `json:"password"`
}

// LoginResponse representa la estructura de datos que se retorna tras una autenticación exitosa.
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
	// ID es el identificador único numérico del producto en la base de datos.
	ID int `json:"id"`
	// Nombre es la descripción comercial en texto del producto.
	Nombre string `json:"nombre"`
	// Descripcion es el texto detallado sobre los beneficios y uso del producto.
	Descripcion string `json:"descripcion,omitempty"`
	// Precio es el costo del producto, almacenado como número con decimales (float64).
	Precio float64 `json:"precio"`
	// Categoria clasifica el tipo de producto (Serum, Crema, Tónico, Mascarilla, Kit).
	Categoria string `json:"categoria,omitempty"`
	// Img es la URL de la fotografía principal del producto.
	Img string `json:"img"`
	// Galeria contiene el arreglo de URLs de fotografías adicionales del producto.
	Galeria []string `json:"galeria,omitempty"`
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