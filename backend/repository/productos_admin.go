// Administración del catálogo: altas, ediciones y borrado lógico.
package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"multicatalogo-backend/models"
)

// Errores de validación del catálogo.
var (
	ErrNombreVacio    = errors.New("el nombre del producto es obligatorio")
	ErrPrecioInvalido = errors.New("el precio debe ser mayor que cero")
	ErrCategoriaVacia = errors.New("la categoría es obligatoria")
	ErrStockNegativo  = errors.New("el stock no puede ser negativo")
)

// columnasProducto evita repetir la lista en cada consulta.
const columnasProducto = `id, nombre, descripcion, precio, categoria, img, galeria, stock, activo`

// CrearProducto da de alta un producto nuevo.
func (r *PostgresRepository) CrearProducto(ctx context.Context, datos models.ProductoRequest) (models.Producto, error) {
	if datos.Nombre == nil || strings.TrimSpace(*datos.Nombre) == "" {
		return models.Producto{}, ErrNombreVacio
	}
	if datos.Precio == nil || *datos.Precio <= 0 {
		return models.Producto{}, ErrPrecioInvalido
	}
	if datos.Categoria == nil || strings.TrimSpace(*datos.Categoria) == "" {
		return models.Producto{}, ErrCategoriaVacia
	}
	if datos.Stock != nil && *datos.Stock < 0 {
		return models.Producto{}, ErrStockNegativo
	}

	galeria := []string{}
	if datos.Galeria != nil {
		galeria = *datos.Galeria
	}
	galeriaJSON, err := json.Marshal(galeria)
	if err != nil {
		return models.Producto{}, fmt.Errorf("galería inválida: %w", err)
	}

	consulta := `
		INSERT INTO productos (nombre, descripcion, precio, categoria, img, galeria, stock, tenant_id)
		VALUES ($1, COALESCE($2, ''), $3, $4, COALESCE($5, ''), $6, COALESCE($7, 0),
		        (SELECT id FROM tenants ORDER BY id LIMIT 1))
		RETURNING ` + columnasProducto

	filas, err := r.db.Query(ctx, consulta,
		strings.TrimSpace(*datos.Nombre), datos.Descripcion, *datos.Precio,
		strings.TrimSpace(*datos.Categoria), datos.Img, galeriaJSON, datos.Stock)
	if err != nil {
		return models.Producto{}, fmt.Errorf("creando producto: %w", err)
	}
	defer filas.Close()

	if !filas.Next() {
		return models.Producto{}, fmt.Errorf("creando producto: sin resultado")
	}
	return escanearProductoAdmin(filas)
}

// ActualizarProducto edita solo los campos recibidos (COALESCE deja el resto igual).
func (r *PostgresRepository) ActualizarProducto(ctx context.Context, id int, datos models.ProductoRequest) (models.Producto, error) {
	if datos.Nombre != nil && strings.TrimSpace(*datos.Nombre) == "" {
		return models.Producto{}, ErrNombreVacio
	}
	if datos.Precio != nil && *datos.Precio <= 0 {
		return models.Producto{}, ErrPrecioInvalido
	}
	if datos.Categoria != nil && strings.TrimSpace(*datos.Categoria) == "" {
		return models.Producto{}, ErrCategoriaVacia
	}
	if datos.Stock != nil && *datos.Stock < 0 {
		return models.Producto{}, ErrStockNegativo
	}

	var galeriaJSON []byte
	if datos.Galeria != nil {
		var err error
		if galeriaJSON, err = json.Marshal(*datos.Galeria); err != nil {
			return models.Producto{}, fmt.Errorf("galería inválida: %w", err)
		}
	}

	consulta := `
		UPDATE productos SET
			nombre      = COALESCE($2, nombre),
			descripcion = COALESCE($3, descripcion),
			precio      = COALESCE($4, precio),
			categoria   = COALESCE($5, categoria),
			img         = COALESCE($6, img),
			galeria     = COALESCE($7, galeria),
			stock       = COALESCE($8, stock),
			activo      = COALESCE($9, activo)
		WHERE id = $1
		RETURNING ` + columnasProducto

	filas, err := r.db.Query(ctx, consulta, id,
		datos.Nombre, datos.Descripcion, datos.Precio, datos.Categoria,
		datos.Img, galeriaJSON, datos.Stock, datos.Activo)
	if err != nil {
		return models.Producto{}, fmt.Errorf("actualizando producto: %w", err)
	}
	defer filas.Close()

	if !filas.Next() {
		if err := filas.Err(); err != nil {
			return models.Producto{}, fmt.Errorf("actualizando producto: %w", err)
		}
		return models.Producto{}, ErrNoEncontrado
	}
	return escanearProductoAdmin(filas)
}

// DesactivarProducto es el borrado lógico: la fila se conserva para el histórico.
func (r *PostgresRepository) DesactivarProducto(ctx context.Context, id int) error {
	etiqueta, err := r.db.Exec(ctx, "UPDATE productos SET activo = FALSE WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("desactivando producto: %w", err)
	}
	if etiqueta.RowsAffected() == 0 {
		return ErrNoEncontrado
	}
	return nil
}

// ListarProductosAdmin incluye los inactivos y permite buscar por texto.
func (r *PostgresRepository) ListarProductosAdmin(ctx context.Context, q string, pagina, limite int) (models.Pagina[models.Producto], error) {
	// ILIKE no distingue mayúsculas; el patrón vacío deja pasar todo.
	patron := "%" + strings.TrimSpace(q) + "%"

	var total int
	if err := r.db.QueryRow(ctx,
		"SELECT count(*) FROM productos WHERE nombre ILIKE $1 OR categoria ILIKE $1",
		patron).Scan(&total); err != nil {
		return models.Pagina[models.Producto]{}, fmt.Errorf("contando productos: %w", err)
	}

	filas, err := r.db.Query(ctx, `
		SELECT `+columnasProducto+`
		FROM productos
		WHERE nombre ILIKE $1 OR categoria ILIKE $1
		ORDER BY id
		LIMIT $2 OFFSET $3`, patron, limite, (pagina-1)*limite)
	if err != nil {
		return models.Pagina[models.Producto]{}, fmt.Errorf("listando productos: %w", err)
	}
	defer filas.Close()

	productos := []models.Producto{}
	for filas.Next() {
		p, err := escanearProductoAdmin(filas)
		if err != nil {
			return models.Pagina[models.Producto]{}, err
		}
		productos = append(productos, p)
	}

	return models.Pagina[models.Producto]{
		Datos: productos, Total: total, Pagina: pagina, Limite: limite,
	}, filas.Err()
}

// escanearProductoAdmin lee también stock y activo, que el catálogo público no expone.
func escanearProductoAdmin(filas pgx.Rows) (models.Producto, error) {
	var p models.Producto
	var galeria []byte

	if err := filas.Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.Precio, &p.Categoria,
		&p.Img, &galeria, &p.Stock, &p.Activo); err != nil {
		return models.Producto{}, fmt.Errorf("leyendo producto: %w", err)
	}

	p.Galeria = []string{}
	if len(galeria) > 0 {
		if err := json.Unmarshal(galeria, &p.Galeria); err != nil {
			return models.Producto{}, fmt.Errorf("galería inválida en el producto %d: %w", p.ID, err)
		}
	}
	return p, nil
}
