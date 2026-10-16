// Declaramos el paquete repository, encargado de la persistencia y recuperación de datos relacionales en PostgreSQL.
package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"multicatalogo-backend/config"
	"multicatalogo-backend/models"
)

var (
	// ErrCredencialesInvalidas indica que el usuario no existe o la contraseña no coincide.
	ErrCredencialesInvalidas = errors.New("credenciales incorrectas")
	// ErrProductoNoEncontrado indica que no existe un producto con el ID solicitado.
	ErrProductoNoEncontrado = errors.New("producto no encontrado")
)

// BuscarUsuarioPorCredenciales ejecuta SELECT * FROM usuarios WHERE email = $1 AND password = $2
func BuscarUsuarioPorCredenciales(email, password string) (*models.Usuario, error) {
	db := config.GetDB()
	if db == nil {
		return nil, errors.New("conexión a la base de datos no inicializada")
	}

	query := `SELECT id, email, password, rol FROM usuarios WHERE email = $1 AND password = $2`
	var user models.Usuario
	err := db.QueryRow(query, email, password).Scan(&user.ID, &user.Email, &user.Password, &user.Rol)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCredencialesInvalidas
		}
		return nil, fmt.Errorf("error al consultar usuario: %w", err)
	}

	return &user, nil
}

// AuthenticateUser valida credenciales contra PostgreSQL y genera la respuesta de autenticación.
func AuthenticateUser(email, password string) (*models.LoginResponse, error) {
	user, err := BuscarUsuarioPorCredenciales(email, password)
	if err != nil {
		return nil, err
	}

	token := fmt.Sprintf("jwt-token-%d-%s", user.ID, user.Rol)
	return &models.LoginResponse{
		Token: token,
		Email: user.Email,
		Rol:   user.Rol,
	}, nil
}

// ObtenerProductos ejecuta un SELECT para recuperar todos los productos desde la base de datos PostgreSQL.
func ObtenerProductos() ([]models.Producto, error) {
	db := config.GetDB()
	if db == nil {
		return nil, errors.New("conexión a la base de datos no inicializada")
	}

	query := `SELECT id, nombre, COALESCE(descripcion, ''), precio, COALESCE(categoria, ''), img, COALESCE(galeria, '[]'::jsonb) FROM productos ORDER BY id ASC`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error al consultar productos: %w", err)
	}
	defer rows.Close()

	productos := make([]models.Producto, 0)
	for rows.Next() {
		var (
			p          models.Producto
			galeriaRaw []byte
		)
		if err := rows.Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.Precio, &p.Categoria, &p.Img, &galeriaRaw); err != nil {
			return nil, fmt.Errorf("error al escanear producto: %w", err)
		}

		if len(galeriaRaw) > 0 {
			_ = json.Unmarshal(galeriaRaw, &p.Galeria)
		}
		if p.Galeria == nil {
			p.Galeria = []string{}
		}

		productos = append(productos, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error durante iteración de productos: %w", err)
	}

	return productos, nil
}

// GetAllProductos es un alias para ObtenerProductos para mantener compatibilidad con controladores existentes.
func GetAllProductos() []models.Producto {
	prods, err := ObtenerProductos()
	if err != nil {
		return []models.Producto{}
	}
	return prods
}

// ObtenerProductoPorID ejecuta un SELECT con parámetro $1 para buscar un producto por su clave primaria.
func ObtenerProductoPorID(id int) (*models.Producto, error) {
	db := config.GetDB()
	if db == nil {
		return nil, errors.New("conexión a la base de datos no inicializada")
	}

	query := `SELECT id, nombre, COALESCE(descripcion, ''), precio, COALESCE(categoria, ''), img, COALESCE(galeria, '[]'::jsonb) FROM productos WHERE id = $1`
	var (
		p          models.Producto
		galeriaRaw []byte
	)

	err := db.QueryRow(query, id).Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.Precio, &p.Categoria, &p.Img, &galeriaRaw)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrProductoNoEncontrado
		}
		return nil, fmt.Errorf("error al consultar producto por ID: %w", err)
	}

	if len(galeriaRaw) > 0 {
		_ = json.Unmarshal(galeriaRaw, &p.Galeria)
	}
	if p.Galeria == nil {
		p.Galeria = []string{}
	}

	return &p, nil
}

// GetProductoByID es un alias para ObtenerProductoPorID para mantener compatibilidad existente.
func GetProductoByID(id int) (*models.Producto, error) {
	return ObtenerProductoPorID(id)
}

// ObtenerRed construye la jerarquía multinivel consultando la tabla 'referidos' en PostgreSQL.
func ObtenerRed() (*models.Referido, error) {
	db := config.GetDB()
	if db == nil {
		return nil, errors.New("conexión a la base de datos no inicializada")
	}

	// Ordenamos descendente por nivel (hojas primero) para construir el árbol recursivo de abajo hacia arriba
	query := `SELECT id, nombre, nivel, ventas, parent_id FROM referidos ORDER BY nivel DESC, id ASC`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error al consultar referidos: %w", err)
	}
	defer rows.Close()

	type rawReferido struct {
		id       int
		nombre   string
		nivel    int
		ventas   float64
		parentID *int
	}

	var list []rawReferido
	nodeMap := make(map[int]*models.Referido)

	for rows.Next() {
		var item rawReferido
		if err := rows.Scan(&item.id, &item.nombre, &item.nivel, &item.ventas, &item.parentID); err != nil {
			return nil, fmt.Errorf("error al escanear referido: %w", err)
		}
		list = append(list, item)
		nodeMap[item.id] = &models.Referido{
			ID:       item.id,
			Nombre:   item.nombre,
			Nivel:    item.nivel,
			Ventas:   item.ventas,
			ParentID: item.parentID,
			Hijos:    []models.Referido{},
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error durante iteración de referidos: %w", err)
	}

	if len(list) == 0 {
		return nil, errors.New("no se encontraron datos de la red de referidos")
	}

	var root *models.Referido

	// Al procesar nivel DESC, cuando agregamos un hijo a parent.Hijos, ese hijo ya tiene todos sus nietos cargados
	for _, item := range list {
		if item.parentID != nil {
			parentPtr, exists := nodeMap[*item.parentID]
			if exists {
				childCopy := *nodeMap[item.id]
				parentPtr.Hijos = append(parentPtr.Hijos, childCopy)
			}
		} else {
			root = nodeMap[item.id]
		}
	}

	if root == nil {
		for _, item := range list {
			if root == nil || item.nivel < root.Nivel {
				root = nodeMap[item.id]
			}
		}
	}

	return root, nil
}

// GetRed es un alias para ObtenerRed.
func GetRed() (*models.Referido, error) {
	return ObtenerRed()
}
