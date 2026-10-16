// Declaramos el paquete repository, encargado de la persistencia y recuperación de datos relacionales en PostgreSQL.
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"multicatalogo-backend/config"
	"multicatalogo-backend/models"
)

var (
	// ErrCredencialesInvalidas indica que el usuario no existe o la contraseña no coincide.
	ErrCredencialesInvalidas = errors.New("credenciales incorrectas")
	// ErrProductoNoEncontrado indica que no existe un producto con el ID solicitado.
	ErrProductoNoEncontrado = errors.New("producto no encontrado")
)

// DBRepository define la interfaz para la persistencia relacional (Inyección de Dependencias - Fase 4).
type DBRepository interface {
	BuscarUsuarioPorCredenciales(email, password string) (*models.Usuario, error)
	AuthenticateUser(email, password string) (*models.LoginResponse, error)
	ObtenerProductos() ([]models.Producto, error)
	ObtenerProductoPorID(id int) (*models.Producto, error)
	ObtenerRed() (*models.Referido, error)
	CrearPedido(req *models.CheckoutRequest) (*models.PedidoConfirmado, error)
	ActualizarComisionesRed(total float64, clienteEmail string) error
}

// PostgresRepository implementa DBRepository sobre una base de datos PostgreSQL real.
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository crea una nueva instancia de PostgresRepository inyectándole la conexión a la base de datos.
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// BuscarUsuarioPorCredenciales ejecuta SELECT * FROM usuarios WHERE email = $1 AND password = $2
func (r *PostgresRepository) BuscarUsuarioPorCredenciales(email, password string) (*models.Usuario, error) {
	if r.db == nil {
		return nil, errors.New("conexión a la base de datos no inicializada")
	}

	query := `SELECT id, email, password, rol FROM usuarios WHERE email = $1 AND password = $2`
	var user models.Usuario
	err := r.db.QueryRow(query, email, password).Scan(&user.ID, &user.Email, &user.Password, &user.Rol)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCredencialesInvalidas
		}
		return nil, fmt.Errorf("error al consultar usuario: %w", err)
	}

	return &user, nil
}

// AuthenticateUser valida credenciales contra PostgreSQL y genera la respuesta de autenticación.
func (r *PostgresRepository) AuthenticateUser(email, password string) (*models.LoginResponse, error) {
	user, err := r.BuscarUsuarioPorCredenciales(email, password)
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
func (r *PostgresRepository) ObtenerProductos() ([]models.Producto, error) {
	if r.db == nil {
		return nil, errors.New("conexión a la base de datos no inicializada")
	}

	query := `SELECT id, nombre, COALESCE(descripcion, ''), precio, COALESCE(categoria, ''), img, COALESCE(galeria, '[]'::jsonb) FROM productos ORDER BY id ASC`
	rows, err := r.db.Query(query)
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

// ObtenerProductoPorID ejecuta un SELECT con parámetro $1 para buscar un producto por su clave primaria.
func (r *PostgresRepository) ObtenerProductoPorID(id int) (*models.Producto, error) {
	if r.db == nil {
		return nil, errors.New("conexión a la base de datos no inicializada")
	}

	query := `SELECT id, nombre, COALESCE(descripcion, ''), precio, COALESCE(categoria, ''), img, COALESCE(galeria, '[]'::jsonb) FROM productos WHERE id = $1`
	var (
		p          models.Producto
		galeriaRaw []byte
	)

	err := r.db.QueryRow(query, id).Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.Precio, &p.Categoria, &p.Img, &galeriaRaw)
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

// ObtenerRed construye la jerarquía multinivel consultando la tabla 'referidos' en PostgreSQL.
func (r *PostgresRepository) ObtenerRed() (*models.Referido, error) {
	if r.db == nil {
		return nil, errors.New("conexión a la base de datos no inicializada")
	}

	query := `SELECT id, nombre, nivel, ventas, parent_id FROM referidos ORDER BY nivel DESC, id ASC`
	rows, err := r.db.Query(query)
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

// CrearPedido implementa una Transacción ACID completa en PostgreSQL (Fase 4 - Paso 3).
// Inserta el pedido principal en 'pedidos', sus ítems en 'pedido_items' y asegura ROLLBACK ante cualquier falla.
func (r *PostgresRepository) CrearPedido(req *models.CheckoutRequest) (*models.PedidoConfirmado, error) {
	if r.db == nil {
		return nil, errors.New("conexión a base de datos no disponible")
	}

	ctx := context.Background()

	// 1. Iniciar transacción SQL explícita (BEGIN)
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("error al iniciar transacción de compra: %w", err)
	}

	// 2. Mecanismo de seguridad ACID: defer tx.Rollback() revierte todo si no se hace Commit()
	defer tx.Rollback()

	// Generamos el número de pedido con prefijo MC-
	numero := fmt.Sprintf("MC-%d", time.Now().Unix()%1000000)

	// 3. Insertar el registro principal del pedido en la tabla 'pedidos'
	queryPedido := `
		INSERT INTO pedidos (numero, cliente_nombre, cliente_email, cliente_ciudad, cliente_direccion, cliente_telefono, metodo_pago, total)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, fecha
	`
	var pedidoID int
	var fecha time.Time
	err = tx.QueryRowContext(ctx, queryPedido,
		numero,
		req.Cliente.Nombre,
		req.Cliente.Email,
		req.Cliente.Ciudad,
		req.Cliente.Direccion,
		req.Cliente.Telefono,
		req.MetodoPago,
		req.Total,
	).Scan(&pedidoID, &fecha)
	if err != nil {
		return nil, fmt.Errorf("error al insertar pedido principal: %w", err)
	}

	// 4. Insertar cada ítem del carrito en la tabla 'pedido_items'
	queryItem := `
		INSERT INTO pedido_items (pedido_id, producto_id, nombre, cantidad, precio, subtotal)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	for _, item := range req.Items {
		var prodID *int
		if item.ID > 0 {
			prodID = &item.ID
		}
		subtotal := item.Precio * float64(item.Cantidad)
		_, err = tx.ExecContext(ctx, queryItem, pedidoID, prodID, item.Nombre, item.Cantidad, item.Precio, subtotal)
		if err != nil {
			return nil, fmt.Errorf("error al insertar detalle de ítem (%s): %w", item.Nombre, err)
		}
	}

	// 5. Confirmación (COMMIT) definitiva de la transacción ACID
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("error al ejecutar COMMIT de la transacción: %w", err)
	}

	fechaFormateada := fecha.Format("02/01/2006, 15:04:05")

	return &models.PedidoConfirmado{
		Numero:  numero,
		Items:   req.Items,
		Total:   req.Total,
		Cliente: req.Cliente,
		Fecha:   fechaFormateada,
	}, nil
}

// ActualizarComisionesRed actualiza las ventas y comisiones en la red multinivel en segundo plano (Fase 4 - Paso 4).
func (r *PostgresRepository) ActualizarComisionesRed(total float64, clienteEmail string) error {
	if r.db == nil {
		return errors.New("conexión a base de datos no disponible")
	}

	// Atribuimos la venta al primer nodo de referidos de nivel 1 (ej. Ana García) para incrementar comisiones y ventas de la red
	query := `UPDATE referidos SET ventas = ventas + $1 WHERE id = 1`
	_, err := r.db.Exec(query, total)
	if err != nil {
		return fmt.Errorf("error al actualizar ventas en tabla referidos: %w", err)
	}

	return nil
}

// =========================================================================
// Funciones globales / wrappers para mantener compatibilidad con versiones previas
// =========================================================================

func getDefaultRepo() *PostgresRepository {
	return NewPostgresRepository(config.GetDB())
}

func BuscarUsuarioPorCredenciales(email, password string) (*models.Usuario, error) {
	return getDefaultRepo().BuscarUsuarioPorCredenciales(email, password)
}

func AuthenticateUser(email, password string) (*models.LoginResponse, error) {
	return getDefaultRepo().AuthenticateUser(email, password)
}

func ObtenerProductos() ([]models.Producto, error) {
	return getDefaultRepo().ObtenerProductos()
}

func GetAllProductos() []models.Producto {
	prods, err := ObtenerProductos()
	if err != nil {
		return []models.Producto{}
	}
	return prods
}

func ObtenerProductoPorID(id int) (*models.Producto, error) {
	return getDefaultRepo().ObtenerProductoPorID(id)
}

func GetProductoByID(id int) (*models.Producto, error) {
	return ObtenerProductoPorID(id)
}

func ObtenerRed() (*models.Referido, error) {
	return getDefaultRepo().ObtenerRed()
}

func GetRed() (*models.Referido, error) {
	return ObtenerRed()
}
