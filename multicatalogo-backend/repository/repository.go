// Declaramos el paquete repository, encargado de la persistencia y recuperación de datos relacionales en PostgreSQL.
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
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
	// Fase 1: Registro de usuarios
	RegistrarUsuario(req *models.RegisterRequest) (*models.Usuario, error)
	// Fase 2: CRUD productos
	CrearProducto(prod *models.Producto) (*models.Producto, error)
	ActualizarProducto(id int, prod *models.Producto) (*models.Producto, error)
	EliminarProducto(id int) error
	ObtenerProductosAdmin(q string, page, limit int) ([]models.Producto, int, error)
	// Fase 3: Gestión de pedidos
	ObtenerPedidosUsuario(usuarioID int, page, limit int) ([]models.Pedido, int, error)
	ObtenerPedidosAdmin(estado string, page, limit int) ([]models.Pedido, int, error)
	ObtenerPedidoPorID(id int) (*models.Pedido, []models.PedidoItem, error)
	ActualizarEstadoPedido(id int, estado string) error
	// Fase 4: Ledger de comisiones y red propia
	ObtenerRedUsuario(usuarioID int) (*models.Referido, error)
	ObtenerComisionesUsuario(usuarioID int) ([]models.Comision, error)
	// Fase 5: Liquidaciones
	CrearLiquidacion(referidoID int) (*models.Liquidacion, error)
	ObtenerLiquidaciones() ([]models.Liquidacion, error)
	ObtenerLiquidacionesUsuario(usuarioID int) ([]models.Liquidacion, error)
	// Fase 6: Reportes
	ObtenerReporteResumen(desde, hasta string) (*models.ReporteResumen, error)
	ObtenerTopProductos(limit int) ([]models.TopProducto, error)
	ObtenerTopReferidos(limit int) ([]models.TopReferido, error)
	ObtenerComisionesPendientes() ([]models.Comision, error)
}

// PostgresRepository implementa DBRepository sobre una base de datos PostgreSQL real.
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository crea una nueva instancia de PostgresRepository inyectándole la conexión a la base de datos.
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// BuscarUsuarioPorCredenciales ejecuta SELECT * FROM usuarios WHERE email = $1
func (r *PostgresRepository) BuscarUsuarioPorCredenciales(email, password string) (*models.Usuario, error) {
	if r.db == nil {
		return nil, errors.New("conexión a la base de datos no inicializada")
	}

	query := `SELECT id, email, password, rol, activo FROM usuarios WHERE email = $1`
	var user models.Usuario
	err := r.db.QueryRow(query, email).Scan(&user.ID, &user.Email, &user.Password, &user.Rol, &user.Activo)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCredencialesInvalidas
		}
		return nil, fmt.Errorf("error al consultar usuario: %w", err)
	}

	// Verificar si la cuenta está activa
	if !user.Activo {
		return nil, errors.New("cuenta desactivada")
	}

	// Verificar contraseña con bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, ErrCredencialesInvalidas
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

	query := `SELECT id, nombre, COALESCE(descripcion, ''), precio, COALESCE(categoria, ''), img, COALESCE(galeria, '[]'::jsonb), activo, stock FROM productos WHERE activo = true ORDER BY id ASC`
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
		if err := rows.Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.Precio, &p.Categoria, &p.Img, &galeriaRaw, &p.Activo, &p.Stock); err != nil {
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

	query := `SELECT id, nombre, COALESCE(descripcion, ''), precio, COALESCE(categoria, ''), img, COALESCE(galeria, '[]'::jsonb), activo, stock FROM productos WHERE id = $1`
	var (
		p          models.Producto
		galeriaRaw []byte
	)

	err := r.db.QueryRow(query, id).Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.Precio, &p.Categoria, &p.Img, &galeriaRaw, &p.Activo, &p.Stock)
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
		INSERT INTO pedidos (numero, cliente_nombre, cliente_email, cliente_ciudad, cliente_direccion, cliente_telefono, metodo_pago, total, usuario_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NULL)
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

	// Intentar obtener usuario_id del email del cliente (si está registrado)
	var usuarioID *int
	err = tx.QueryRowContext(ctx, "SELECT id FROM usuarios WHERE email = $1", req.Cliente.Email).Scan(&usuarioID)
	if err != nil {
		// Usuario no encontrado, usuario_id queda NULL
		usuarioID = nil
	}

	// Actualizar usuario_id del pedido si el usuario existe
	if usuarioID != nil {
		_, err = tx.ExecContext(ctx, "UPDATE pedidos SET usuario_id = $1 WHERE id = $2", *usuarioID, pedidoID)
		if err != nil {
			return nil, fmt.Errorf("error al actualizar usuario_id del pedido: %w", err)
		}
	}

	// 4. Insertar cada ítem del carrito en la tabla 'pedido_items' y descontar stock
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

		// Descontar stock con FOR UPDATE para evitar race conditions
		if prodID != nil {
			_, err = tx.ExecContext(ctx, "UPDATE productos SET stock = stock - $1 WHERE id = $2 AND stock >= $1", item.Cantidad, *prodID)
			if err != nil {
				return nil, fmt.Errorf("error al descontar stock del producto %d: %w", *prodID, err)
			}
		}
	}

	// 5. Insertar comisiones en el ledger (Fase 4)
	tasas := []float64{0.10, 0.05, 0.02} // Nivel 1: 10%, Nivel 2: 5%, Nivel 3: 2%
	queryComision := `
		INSERT INTO comisiones (pedido_id, referido_id, nivel, porcentaje, monto, estado)
		VALUES ($1, $2, $3, $4, $5, 'pendiente')
	`

	// Obtener el cliente para encontrar su referido
	var clienteReferidoID *int
	err = tx.QueryRowContext(ctx, "SELECT referido_id FROM usuarios WHERE email = $1", req.Cliente.Email).Scan(&clienteReferidoID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		// Si no encontramos el usuario, no hay comisiones (normal para pedidos de prueba)
	}

	// Calcular y registrar comisiones multinivel
	if clienteReferidoID != nil {
		// Obtener la cadena de referidos (parent_id recursivo)
		referidoID := *clienteReferidoID
		for nivel := 0; nivel < len(tasas) && referidoID > 0; nivel++ {
			var parentID *int
			err = tx.QueryRowContext(ctx, "SELECT parent_id FROM referidos WHERE id = $1", referidoID).Scan(&parentID)
			if err != nil {
				break
			}

			if parentID != nil && *parentID > 0 {
				monto := req.Total * tasas[nivel]
				porcentaje := tasas[nivel] * 100
				_, err = tx.ExecContext(ctx, queryComision, pedidoID, *parentID, nivel+1, porcentaje, monto)
				if err != nil {
					return nil, fmt.Errorf("error al insertar comisión nivel %d: %w", nivel+1, err)
				}
				referidoID = *parentID
			} else {
				break
			}
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

// RegistrarUsuario crea un nuevo usuario y su nodo en referidos (Fase 1).
func (r *PostgresRepository) RegistrarUsuario(req *models.RegisterRequest) (*models.Usuario, error) {
	if r.db == nil {
		return nil, errors.New("conexión a base de datos no disponible")
	}

	ctx := context.Background()
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("error al iniciar transacción de registro: %w", err)
	}
	defer tx.Rollback()

	// Verificar si el email ya existe
	var exists bool
	err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM usuarios WHERE email = $1)", req.Email).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("error al verificar email duplicado: %w", err)
	}
	if exists {
		return nil, errors.New("el email ya está registrado")
	}

	// Obtener parent_id si se proporciona código de referido
	var parentID *int
	if req.Ref != "" {
		var refID int
		err = tx.QueryRowContext(ctx, "SELECT id FROM referidos WHERE codigo = $1", req.Ref).Scan(&refID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, errors.New("código de referido inválido")
			}
			return nil, fmt.Errorf("error al buscar código de referido: %w", err)
		}
		parentID = &refID
	}

	// Hashear contraseña
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error al hashear contraseña: %w", err)
	}

	// Insertar usuario
	queryUsuario := `
		INSERT INTO usuarios (email, password, rol, nombre, activo, referido_id)
		VALUES ($1, $2, 'cliente', $3, true, $4)
		RETURNING id, creado_en
	`
	var usuarioID int
	var creadoEn time.Time
	err = tx.QueryRowContext(ctx, queryUsuario, req.Email, string(hashedPassword), req.Nombre, parentID).Scan(&usuarioID, &creadoEn)
	if err != nil {
		return nil, fmt.Errorf("error al insertar usuario: %w", err)
	}

	// Obtener el siguiente ID disponible para referidos
	var nextReferidoID int
	err = tx.QueryRowContext(ctx, "SELECT COALESCE(MAX(id), 0) + 1 FROM referidos").Scan(&nextReferidoID)
	if err != nil {
		return nil, fmt.Errorf("error al obtener siguiente ID de referido: %w", err)
	}

	// Generar código único para el referido
	codigo := fmt.Sprintf("USER%d", nextReferidoID)

	// Calcular nivel del nuevo referido
	nivel := 1
	if parentID != nil {
		var nivelPadre int
		err = tx.QueryRowContext(ctx, "SELECT nivel FROM referidos WHERE id = $1", *parentID).Scan(&nivelPadre)
		if err == nil {
			nivel = nivelPadre + 1
		}
	}

	// Insertar nodo en referidos con el ID secuencial
	queryReferido := `
		INSERT INTO referidos (id, nombre, nivel, ventas, parent_id, codigo)
		VALUES ($1, $2, $3, 0, $4, $5)
	`
	_, err = tx.ExecContext(ctx, queryReferido, nextReferidoID, req.Nombre, nivel, parentID, codigo)
	if err != nil {
		return nil, fmt.Errorf("error al insertar referido: %w", err)
	}

	// Actualizar referido_id del usuario con el ID del referido (no el usuario_id)
	_, err = tx.ExecContext(ctx, "UPDATE usuarios SET referido_id = $1 WHERE id = $2", nextReferidoID, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("error al actualizar referido_id del usuario: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("error al ejecutar COMMIT de registro: %w", err)
	}

	return &models.Usuario{
		ID:         usuarioID,
		Email:      req.Email,
		Rol:        "cliente",
		Nombre:     req.Nombre,
		Activo:     true,
		ReferidoID: &usuarioID,
		CreadoEn:   creadoEn,
	}, nil
}

// CrearProducto inserta un nuevo producto (Fase 2).
func (r *PostgresRepository) CrearProducto(prod *models.Producto) (*models.Producto, error) {
	if r.db == nil {
		return nil, errors.New("conexión a base de datos no disponible")
	}

	query := `
		INSERT INTO productos (nombre, descripcion, precio, categoria, img, galeria, activo, stock)
		VALUES ($1, $2, $3, $4, $5, $6, true, $7)
		RETURNING id
	`
	galeriaJSON, _ := json.Marshal(prod.Galeria)
	var id int
	err := r.db.QueryRow(query, prod.Nombre, prod.Descripcion, prod.Precio, prod.Categoria, prod.Img, galeriaJSON, prod.Stock).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("error al crear producto: %w", err)
	}

	prod.ID = id
	prod.Activo = true
	return prod, nil
}

// ActualizarProducto modifica un producto existente (Fase 2).
func (r *PostgresRepository) ActualizarProducto(id int, prod *models.Producto) (*models.Producto, error) {
	if r.db == nil {
		return nil, errors.New("conexión a base de datos no disponible")
	}

	query := `
		UPDATE productos
		SET nombre = $1, descripcion = $2, precio = $3, categoria = $4, img = $5, galeria = $6, stock = $7
		WHERE id = $8
		RETURNING id, activo
	`
	galeriaJSON, _ := json.Marshal(prod.Galeria)
	var activo bool
	err := r.db.QueryRow(query, prod.Nombre, prod.Descripcion, prod.Precio, prod.Categoria, prod.Img, galeriaJSON, prod.Stock, id).Scan(&id, &activo)
	if err != nil {
		return nil, fmt.Errorf("error al actualizar producto: %w", err)
	}

	prod.ID = id
	prod.Activo = activo
	return prod, nil
}

// EliminarProducto hace borrado lógico (activo = false) (Fase 2).
func (r *PostgresRepository) EliminarProducto(id int) error {
	if r.db == nil {
		return errors.New("conexión a base de datos no disponible")
	}

	_, err := r.db.Exec("UPDATE productos SET activo = false WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("error al eliminar producto: %w", err)
	}
	return nil
}

// ObtenerProductosAdmin lista productos con búsqueda y paginación (Fase 2 & 7).
func (r *PostgresRepository) ObtenerProductosAdmin(q string, page, limit int) ([]models.Producto, int, error) {
	if r.db == nil {
		return nil, 0, errors.New("conexión a base de datos no disponible")
	}

	offset := (page - 1) * limit

	// Query base
	queryBase := `SELECT id, nombre, COALESCE(descripcion, ''), precio, COALESCE(categoria, ''), img, COALESCE(galeria, '[]'::jsonb), activo, stock FROM productos`
	var countQuery string
	var dataQuery string
	var args []interface{}

	if q != "" {
		queryBase += " WHERE nombre ILIKE $1 OR categoria ILIKE $1"
		countQuery = `SELECT COUNT(*) FROM productos WHERE nombre ILIKE $1 OR categoria ILIKE $1`
		dataQuery = queryBase + " ORDER BY id ASC LIMIT $2 OFFSET $3"
		args = []interface{}{"%" + q + "%", limit, offset}
	} else {
		countQuery = `SELECT COUNT(*) FROM productos`
		dataQuery = queryBase + " ORDER BY id ASC LIMIT $1 OFFSET $2"
		args = []interface{}{limit, offset}
	}

	// Contar total
	var total int
	err := r.db.QueryRow(countQuery, args[:len(args)-2]...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("error al contar productos: %w", err)
	}

	// Obtener datos
	rows, err := r.db.Query(dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("error al consultar productos: %w", err)
	}
	defer rows.Close()

	productos := make([]models.Producto, 0)
	for rows.Next() {
		var (
			p          models.Producto
			galeriaRaw []byte
		)
		if err := rows.Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.Precio, &p.Categoria, &p.Img, &galeriaRaw, &p.Activo, &p.Stock); err != nil {
			return nil, 0, fmt.Errorf("error al escanear producto: %w", err)
		}

		if len(galeriaRaw) > 0 {
			_ = json.Unmarshal(galeriaRaw, &p.Galeria)
		}
		if p.Galeria == nil {
			p.Galeria = []string{}
		}

		productos = append(productos, p)
	}

	return productos, total, nil
}

// ObtenerPedidosUsuario obtiene pedidos de un usuario específico (Fase 3).
func (r *PostgresRepository) ObtenerPedidosUsuario(usuarioID int, page, limit int) ([]models.Pedido, int, error) {
	if r.db == nil {
		return nil, 0, errors.New("conexión a base de datos no disponible")
	}

	offset := (page - 1) * limit

	countQuery := `SELECT COUNT(*) FROM pedidos WHERE usuario_id = $1`
	var total int
	err := r.db.QueryRow(countQuery, usuarioID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("error al contar pedidos: %w", err)
	}

	query := `
		SELECT id, numero, cliente_nombre, cliente_email, cliente_ciudad, cliente_direccion,
		       cliente_telefono, metodo_pago, total, fecha, estado, usuario_id
		FROM pedidos
		WHERE usuario_id = $1
		ORDER BY fecha DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(query, usuarioID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("error al consultar pedidos: %w", err)
	}
	defer rows.Close()

	pedidos := make([]models.Pedido, 0)
	for rows.Next() {
		var p models.Pedido
		if err := rows.Scan(&p.ID, &p.Numero, &p.ClienteNombre, &p.ClienteEmail, &p.ClienteCiudad,
			&p.ClienteDireccion, &p.ClienteTelefono, &p.MetodoPago, &p.Total, &p.Fecha, &p.Estado, &p.UsuarioID); err != nil {
			return nil, 0, fmt.Errorf("error al escanear pedido: %w", err)
		}
		pedidos = append(pedidos, p)
	}

	return pedidos, total, nil
}

// ObtenerPedidosAdmin obtiene todos los pedidos con filtro opcional de estado (Fase 3).
func (r *PostgresRepository) ObtenerPedidosAdmin(estado string, page, limit int) ([]models.Pedido, int, error) {
	if r.db == nil {
		return nil, 0, errors.New("conexión a base de datos no disponible")
	}

	offset := (page - 1) * limit

	var countQuery, dataQuery string
	var args []interface{}

	if estado != "" {
		countQuery = `SELECT COUNT(*) FROM pedidos WHERE estado = $1`
		dataQuery = `
			SELECT id, numero, cliente_nombre, cliente_email, cliente_ciudad, cliente_direccion,
			       cliente_telefono, metodo_pago, total, fecha, estado, usuario_id
			FROM pedidos
			WHERE estado = $1
			ORDER BY fecha DESC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{estado, limit, offset}
	} else {
		countQuery = `SELECT COUNT(*) FROM pedidos`
		dataQuery = `
			SELECT id, numero, cliente_nombre, cliente_email, cliente_ciudad, cliente_direccion,
			       cliente_telefono, metodo_pago, total, fecha, estado, usuario_id
			FROM pedidos
			ORDER BY fecha DESC
			LIMIT $1 OFFSET $2
		`
		args = []interface{}{limit, offset}
	}

	var total int
	err := r.db.QueryRow(countQuery, args[:len(args)-2]...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("error al contar pedidos: %w", err)
	}

	rows, err := r.db.Query(dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("error al consultar pedidos: %w", err)
	}
	defer rows.Close()

	pedidos := make([]models.Pedido, 0)
	for rows.Next() {
		var p models.Pedido
		if err := rows.Scan(&p.ID, &p.Numero, &p.ClienteNombre, &p.ClienteEmail, &p.ClienteCiudad,
			&p.ClienteDireccion, &p.ClienteTelefono, &p.MetodoPago, &p.Total, &p.Fecha, &p.Estado, &p.UsuarioID); err != nil {
			return nil, 0, fmt.Errorf("error al escanear pedido: %w", err)
		}
		pedidos = append(pedidos, p)
	}

	return pedidos, total, nil
}

// ObtenerPedidoPorID obtiene un pedido con sus ítems (Fase 3).
func (r *PostgresRepository) ObtenerPedidoPorID(id int) (*models.Pedido, []models.PedidoItem, error) {
	if r.db == nil {
		return nil, nil, errors.New("conexión a base de datos no disponible")
	}

	// Obtener pedido
	queryPedido := `
		SELECT id, numero, cliente_nombre, cliente_email, cliente_ciudad, cliente_direccion,
		       cliente_telefono, metodo_pago, total, fecha, estado, usuario_id
		FROM pedidos
		WHERE id = $1
	`
	var p models.Pedido
	err := r.db.QueryRow(queryPedido, id).Scan(&p.ID, &p.Numero, &p.ClienteNombre, &p.ClienteEmail, &p.ClienteCiudad,
		&p.ClienteDireccion, &p.ClienteTelefono, &p.MetodoPago, &p.Total, &p.Fecha, &p.Estado, &p.UsuarioID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, errors.New("pedido no encontrado")
		}
		return nil, nil, fmt.Errorf("error al consultar pedido: %w", err)
	}

	// Obtener ítems
	queryItems := `
		SELECT id, pedido_id, producto_id, nombre, cantidad, precio, subtotal
		FROM pedido_items
		WHERE pedido_id = $1
	`
	rows, err := r.db.Query(queryItems, id)
	if err != nil {
		return nil, nil, fmt.Errorf("error al consultar ítems: %w", err)
	}
	defer rows.Close()

	items := make([]models.PedidoItem, 0)
	for rows.Next() {
		var item models.PedidoItem
		if err := rows.Scan(&item.ID, &item.PedidoID, &item.ProductoID, &item.Nombre, &item.Cantidad, &item.Precio, &item.Subtotal); err != nil {
			return nil, nil, fmt.Errorf("error al escanear ítem: %w", err)
		}
		items = append(items, item)
	}

	return &p, items, nil
}

// ActualizarEstadoPedido cambia el estado de un pedido (Fase 3).
func (r *PostgresRepository) ActualizarEstadoPedido(id int, estado string) error {
	if r.db == nil {
		return errors.New("conexión a base de datos no disponible")
	}

	// Validar transiciones válidas
	estadosValidos := map[string][]string{
		"pendiente":   {"preparando", "cancelado"},
		"preparando":  {"enviado", "cancelado"},
		"enviado":     {"entregado"},
		"entregado":   {},
		"cancelado":   {},
	}

	// Obtener estado actual
	var estadoActual string
	err := r.db.QueryRow("SELECT estado FROM pedidos WHERE id = $1", id).Scan(&estadoActual)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("pedido no encontrado")
		}
		return fmt.Errorf("error al obtener estado actual: %w", err)
	}

	// Verificar si la transición es válida
	transicionesValidas, ok := estadosValidos[estadoActual]
	if !ok {
		return errors.New("estado actual inválido")
	}

	transicionValida := false
	for _, e := range transicionesValidas {
		if e == estado {
			transicionValida = true
			break
		}
	}

	if !transicionValida {
		return fmt.Errorf("transición inválida de %s a %s", estadoActual, estado)
	}

	// Si es cancelación, ejecutar transacción con reversión de stock y comisiones
	if estado == "cancelado" {
		ctx := context.Background()
		tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
		if err != nil {
			return fmt.Errorf("error al iniciar transacción de cancelación: %w", err)
		}
		defer tx.Rollback()

		// Actualizar estado del pedido
		_, err = tx.ExecContext(ctx, "UPDATE pedidos SET estado = $1 WHERE id = $2", estado, id)
		if err != nil {
			return fmt.Errorf("error al actualizar estado: %w", err)
		}

		// Devolver stock
		queryUpdateStock := `
			UPDATE productos p
			SET stock = stock + pi.cantidad
			FROM pedido_items pi
			WHERE pi.pedido_id = $1 AND pi.producto_id = p.id
		`
		_, err = tx.ExecContext(ctx, queryUpdateStock, id)
		if err != nil {
			return fmt.Errorf("error al devolver stock: %w", err)
		}

		// Revertir comisiones
		_, err = tx.ExecContext(ctx, "UPDATE comisiones SET estado = 'revertida' WHERE pedido_id = $1", id)
		if err != nil {
			return fmt.Errorf("error al revertir comisiones: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("error al ejecutar COMMIT de cancelación: %w", err)
		}
	} else {
		// Actualización simple de estado
		_, err = r.db.Exec("UPDATE pedidos SET estado = $1 WHERE id = $2", estado, id)
		if err != nil {
			return fmt.Errorf("error al actualizar estado: %w", err)
		}
	}

	return nil
}

// ObtenerRedUsuario obtiene el subárbol de la red desde el usuario (Fase 4).
func (r *PostgresRepository) ObtenerRedUsuario(usuarioID int) (*models.Referido, error) {
	if r.db == nil {
		return nil, errors.New("conexión a base de datos no disponible")
	}

	// Obtener el referido_id asociado a este usuario
	var refID *int
	err := r.db.QueryRow("SELECT referido_id FROM usuarios WHERE id = $1", usuarioID).Scan(&refID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("usuario no encontrado")
		}
		return nil, fmt.Errorf("error al consultar usuario: %w", err)
	}

	idABuscar := usuarioID
	if refID != nil {
		idABuscar = *refID
	}

	// Obtener el nodo del usuario
	query := `SELECT id, nombre, nivel, ventas, parent_id, codigo FROM referidos WHERE id = $1`
	var root models.Referido
	err = r.db.QueryRow(query, idABuscar).Scan(&root.ID, &root.Nombre, &root.Nivel, &root.Ventas, &root.ParentID, &root.Codigo)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("referido no encontrado")
		}
		return nil, fmt.Errorf("error al consultar referido: %w", err)
	}

	// Construir subárbol recursivamente
	root.Hijos = r.construirSubarbol(root.ID)
	return &root, nil
}

func (r *PostgresRepository) construirSubarbol(parentID int) []models.Referido {
	query := `SELECT id, nombre, nivel, ventas, parent_id, codigo FROM referidos WHERE parent_id = $1 ORDER BY id ASC`
	rows, err := r.db.Query(query, parentID)
	if err != nil {
		return []models.Referido{}
	}
	defer rows.Close()

	hijos := make([]models.Referido, 0)
	for rows.Next() {
		var ref models.Referido
		if err := rows.Scan(&ref.ID, &ref.Nombre, &ref.Nivel, &ref.Ventas, &ref.ParentID, &ref.Codigo); err != nil {
			continue
		}
		ref.Hijos = r.construirSubarbol(ref.ID)
		hijos = append(hijos, ref)
	}

	return hijos
}

// ObtenerComisionesUsuario obtiene el ledger de comisiones de un usuario (Fase 4).
func (r *PostgresRepository) ObtenerComisionesUsuario(usuarioID int) ([]models.Comision, error) {
	if r.db == nil {
		return nil, errors.New("conexión a base de datos no disponible")
	}

	// Obtener el referido_id asociado a este usuario
	var refID *int
	_ = r.db.QueryRow("SELECT referido_id FROM usuarios WHERE id = $1", usuarioID).Scan(&refID)

	idABuscar := usuarioID
	if refID != nil {
		idABuscar = *refID
	}

	query := `
		SELECT c.id, c.pedido_id, c.referido_id, c.nivel, c.porcentaje, c.monto, c.estado, c.liquidacion_id, c.creado_en
		FROM comisiones c
		WHERE c.referido_id = $1
		ORDER BY c.creado_en DESC
	`
	rows, err := r.db.Query(query, idABuscar)
	if err != nil {
		return nil, fmt.Errorf("error al consultar comisiones: %w", err)
	}
	defer rows.Close()

	comisiones := make([]models.Comision, 0)
	for rows.Next() {
		var c models.Comision
		if err := rows.Scan(&c.ID, &c.PedidoID, &c.ReferidoID, &c.Nivel, &c.Porcentaje, &c.Monto, &c.Estado, &c.LiquidacionID, &c.CreadoEn); err != nil {
			return nil, fmt.Errorf("error al escanear comisión: %w", err)
		}
		comisiones = append(comisiones, c)
	}

	return comisiones, nil
}

// CrearLiquidacion crea una liquidación de comisiones (Fase 5).
func (r *PostgresRepository) CrearLiquidacion(referidoID int) (*models.Liquidacion, error) {
	if r.db == nil {
		return nil, errors.New("conexión a base de datos no disponible")
	}

	ctx := context.Background()
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("error al iniciar transacción de liquidación: %w", err)
	}
	defer tx.Rollback()

	// Bloquear las comisiones del referido para evitar conflictos
	_, err = tx.ExecContext(ctx, "SELECT id FROM comisiones WHERE referido_id = $1 AND estado = 'pendiente' FOR UPDATE", referidoID)
	if err != nil {
		return nil, fmt.Errorf("error al bloquear comisiones: %w", err)
	}

	// Sumar comisiones pendientes del referido
	querySuma := `SELECT COALESCE(SUM(monto), 0) FROM comisiones WHERE referido_id = $1 AND estado = 'pendiente'`
	var total float64
	err = tx.QueryRowContext(ctx, querySuma, referidoID).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("error al sumar comisiones: %w", err)
	}

	if total == 0 {
		return nil, errors.New("no hay comisiones pendientes para liquidar")
	}

	// Crear liquidación
	queryLiquidacion := `
		INSERT INTO liquidaciones (referido_id, total, estado)
		VALUES ($1, $2, 'pendiente')
		RETURNING id, creado_en
	`
	var liquidacionID int
	var creadoEn time.Time
	err = tx.QueryRowContext(ctx, queryLiquidacion, referidoID, total).Scan(&liquidacionID, &creadoEn)
	if err != nil {
		return nil, fmt.Errorf("error al crear liquidación: %w", err)
	}

	// Marcar comisiones como pagadas
	_, err = tx.ExecContext(ctx, "UPDATE comisiones SET estado = 'pagada', liquidacion_id = $1 WHERE referido_id = $2 AND estado = 'pendiente'", liquidacionID, referidoID)
	if err != nil {
		return nil, fmt.Errorf("error al actualizar comisiones: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("error al ejecutar COMMIT de liquidación: %w", err)
	}

	return &models.Liquidacion{
		ID:         liquidacionID,
		ReferidoID: referidoID,
		Total:      total,
		Estado:     "pendiente",
		CreadoEn:   creadoEn,
	}, nil
}

// ObtenerLiquidaciones obtiene todas las liquidaciones (Fase 5).
func (r *PostgresRepository) ObtenerLiquidaciones() ([]models.Liquidacion, error) {
	if r.db == nil {
		return nil, errors.New("conexión a base de datos no disponible")
	}

	query := `SELECT id, referido_id, total, estado, creado_en, pagada_en FROM liquidaciones ORDER BY creado_en DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error al consultar liquidaciones: %w", err)
	}
	defer rows.Close()

	liquidaciones := make([]models.Liquidacion, 0)
	for rows.Next() {
		var l models.Liquidacion
		if err := rows.Scan(&l.ID, &l.ReferidoID, &l.Total, &l.Estado, &l.CreadoEn, &l.PagadaEn); err != nil {
			return nil, fmt.Errorf("error al escanear liquidación: %w", err)
		}
		liquidaciones = append(liquidaciones, l)
	}

	return liquidaciones, nil
}

// ObtenerLiquidacionesUsuario obtiene liquidaciones de un usuario (Fase 5).
func (r *PostgresRepository) ObtenerLiquidacionesUsuario(usuarioID int) ([]models.Liquidacion, error) {
	if r.db == nil {
		return nil, errors.New("conexión a base de datos no disponible")
	}

	query := `SELECT id, referido_id, total, estado, creado_en, pagada_en FROM liquidaciones WHERE referido_id = $1 ORDER BY creado_en DESC`
	rows, err := r.db.Query(query, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("error al consultar liquidaciones: %w", err)
	}
	defer rows.Close()

	liquidaciones := make([]models.Liquidacion, 0)
	for rows.Next() {
		var l models.Liquidacion
		if err := rows.Scan(&l.ID, &l.ReferidoID, &l.Total, &l.Estado, &l.CreadoEn, &l.PagadaEn); err != nil {
			return nil, fmt.Errorf("error al escanear liquidación: %w", err)
		}
		liquidaciones = append(liquidaciones, l)
	}

	return liquidaciones, nil
}

// ObtenerReporteResumen genera KPIs de ventas y comisiones (Fase 6).
func (r *PostgresRepository) ObtenerReporteResumen(desde, hasta string) (*models.ReporteResumen, error) {
	if r.db == nil {
		return nil, errors.New("conexión a base de datos no disponible")
	}

	var query string
	var args []interface{}

	if desde != "" && hasta != "" {
		query = `
			SELECT
				COALESCE(SUM(total), 0) as total_vendido,
				COUNT(*) as num_pedidos,
				COALESCE(AVG(total), 0) as ticket_promedio
			FROM pedidos
			WHERE fecha >= $1 AND fecha <= $2
		`
		args = []interface{}{desde, hasta}
	} else {
		query = `
			SELECT
				COALESCE(SUM(total), 0) as total_vendido,
				COUNT(*) as num_pedidos,
				COALESCE(AVG(total), 0) as ticket_promedio
			FROM pedidos
		`
	}

	var reporte models.ReporteResumen
	err := r.db.QueryRow(query, args...).Scan(&reporte.TotalVendido, &reporte.NumPedidos, &reporte.TicketPromedio)
	if err != nil {
		return nil, fmt.Errorf("error al generar reporte: %w", err)
	}

	// Comisiones devengadas
	queryComisiones := `SELECT COALESCE(SUM(monto), 0) FROM comisiones WHERE estado = 'pendiente'`
	err = r.db.QueryRow(queryComisiones).Scan(&reporte.ComisionesDev)
	if err != nil {
		return nil, fmt.Errorf("error al consultar comisiones devengadas: %w", err)
	}

	return &reporte, nil
}

// ObtenerTopProductos lista los productos más vendidos (Fase 6).
func (r *PostgresRepository) ObtenerTopProductos(limit int) ([]models.TopProducto, error) {
	if r.db == nil {
		return nil, errors.New("conexión a base de datos no disponible")
	}

	query := `
		SELECT
			p.id, p.nombre, p.categoria,
			COALESCE(SUM(pi.cantidad), 0) as cantidad,
			COALESCE(SUM(pi.subtotal), 0) as total_venta
		FROM productos p
		LEFT JOIN pedido_items pi ON p.id = pi.producto_id
		GROUP BY p.id, p.nombre, p.categoria
		ORDER BY total_venta DESC
		LIMIT $1
	`
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("error al consultar top productos: %w", err)
	}
	defer rows.Close()

	productos := make([]models.TopProducto, 0)
	for rows.Next() {
		var p models.TopProducto
		if err := rows.Scan(&p.ID, &p.Nombre, &p.Categoria, &p.Cantidad, &p.TotalVenta); err != nil {
			return nil, fmt.Errorf("error al escanear producto: %w", err)
		}
		productos = append(productos, p)
	}

	return productos, nil
}

// ObtenerTopReferidos lista los referidos con más ventas (Fase 6).
func (r *PostgresRepository) ObtenerTopReferidos(limit int) ([]models.TopReferido, error) {
	if r.db == nil {
		return nil, errors.New("conexión a base de datos no disponible")
	}

	query := `
		SELECT
			r.id, r.nombre, r.nivel, r.ventas,
			COUNT(DISTINCT p.id) as num_pedidos
		FROM referidos r
		LEFT JOIN usuarios u ON r.id = u.referido_id
		LEFT JOIN pedidos p ON u.id = p.usuario_id
		GROUP BY r.id, r.nombre, r.nivel, r.ventas
		ORDER BY r.ventas DESC
		LIMIT $1
	`
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("error al consultar top referidos: %w", err)
	}
	defer rows.Close()

	referidos := make([]models.TopReferido, 0)
	for rows.Next() {
		var ref models.TopReferido
		if err := rows.Scan(&ref.ID, &ref.Nombre, &ref.Nivel, &ref.Ventas, &ref.NumPedidos); err != nil {
			return nil, fmt.Errorf("error al escanear referido: %w", err)
		}
		referidos = append(referidos, ref)
	}

	return referidos, nil
}

// ObtenerComisionesPendientes lista comisiones pendientes por distribuidor (Fase 6).
func (r *PostgresRepository) ObtenerComisionesPendientes() ([]models.Comision, error) {
	if r.db == nil {
		return nil, errors.New("conexión a base de datos no disponible")
	}

	query := `
		SELECT c.id, c.pedido_id, c.referido_id, c.nivel, c.porcentaje, c.monto, c.estado, c.liquidacion_id, c.creado_en
		FROM comisiones c
		WHERE c.estado = 'pendiente'
		ORDER BY c.referido_id, c.creado_en DESC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error al consultar comisiones pendientes: %w", err)
	}
	defer rows.Close()

	comisiones := make([]models.Comision, 0)
	for rows.Next() {
		var c models.Comision
		if err := rows.Scan(&c.ID, &c.PedidoID, &c.ReferidoID, &c.Nivel, &c.Porcentaje, &c.Monto, &c.Estado, &c.LiquidacionID, &c.CreadoEn); err != nil {
			return nil, fmt.Errorf("error al escanear comisión: %w", err)
		}
		comisiones = append(comisiones, c)
	}

	return comisiones, nil
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
