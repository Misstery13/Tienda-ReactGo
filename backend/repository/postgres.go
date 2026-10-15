// Implementación de DBRepository sobre PostgreSQL.
package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"multicatalogo-backend/models"
)

// ErrNoEncontrado indica que la consulta no devolvió filas.
// Permite que el controlador distinga un 404 de un error real de base de datos.
var ErrNoEncontrado = errors.New("recurso no encontrado")

// PostgresRepository recibe el pool de conexiones por inyección.
type PostgresRepository struct {
	db *pgxpool.Pool
}

// NuevoPostgresRepository construye el repositorio con su dependencia.
func NuevoPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// BuscarUsuario devuelve el usuario cuyas credenciales coinciden.
// El correo se compara sin distinguir mayúsculas; la contraseña sí es sensible.
func (r *PostgresRepository) BuscarUsuario(ctx context.Context, email, password string) (models.Usuario, error) {
	const consulta = `
		SELECT id, email, rol
		FROM usuarios
		WHERE LOWER(email) = LOWER($1) AND password = $2`

	var u models.Usuario
	err := r.db.QueryRow(ctx, consulta, email, password).Scan(&u.ID, &u.Email, &u.Rol)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Usuario{}, ErrNoEncontrado
	}
	if err != nil {
		return models.Usuario{}, fmt.Errorf("consultando usuario: %w", err)
	}

	// El token sigue siendo ficticio; se reemplazará por JWT firmado más adelante.
	u.Token = fmt.Sprintf("fake-jwt-token-%d", u.ID)
	return u, nil
}

// ObtenerProductos devuelve el catálogo completo ordenado por id.
func (r *PostgresRepository) ObtenerProductos(ctx context.Context) ([]models.Producto, error) {
	const consulta = `
		SELECT id, nombre, descripcion, precio, categoria, img, galeria
		FROM productos
		ORDER BY id`

	filas, err := r.db.Query(ctx, consulta)
	if err != nil {
		return nil, fmt.Errorf("consultando productos: %w", err)
	}
	defer filas.Close()

	// Slice inicializado: así un catálogo vacío se serializa como [] y no como null.
	productos := []models.Producto{}
	for filas.Next() {
		producto, err := escanearProducto(filas)
		if err != nil {
			return nil, err
		}
		productos = append(productos, producto)
	}
	if err := filas.Err(); err != nil {
		return nil, fmt.Errorf("consultando productos: %w", err)
	}

	return productos, nil
}

// ObtenerProductoPorID busca un producto por su identificador.
func (r *PostgresRepository) ObtenerProductoPorID(ctx context.Context, id int) (models.Producto, error) {
	const consulta = `
		SELECT id, nombre, descripcion, precio, categoria, img, galeria
		FROM productos
		WHERE id = $1`

	filas, err := r.db.Query(ctx, consulta, id)
	if err != nil {
		return models.Producto{}, fmt.Errorf("consultando producto: %w", err)
	}
	defer filas.Close()

	if !filas.Next() {
		if err := filas.Err(); err != nil {
			return models.Producto{}, fmt.Errorf("consultando producto: %w", err)
		}
		return models.Producto{}, ErrNoEncontrado
	}

	return escanearProducto(filas)
}

// escanearProducto convierte una fila en un models.Producto.
// La galería llega como JSONB, por eso se deserializa aparte.
func escanearProducto(filas pgx.Rows) (models.Producto, error) {
	var p models.Producto
	var galeria []byte

	if err := filas.Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.Precio, &p.Categoria, &p.Img, &galeria); err != nil {
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

// ObtenerRed arma el árbol multinivel completo.
// Trae la tabla como lista plana y reconstruye la jerarquía con parent_id.
func (r *PostgresRepository) ObtenerRed(ctx context.Context) (*models.Referido, error) {
	const consulta = `
		SELECT id, nombre, nivel, ventas, parent_id
		FROM referidos
		ORDER BY nivel, id`

	filas, err := r.db.Query(ctx, consulta)
	if err != nil {
		return nil, fmt.Errorf("consultando referidos: %w", err)
	}
	defer filas.Close()

	nodos := map[int]*models.Referido{}
	padres := map[int]*int{}
	orden := []int{}

	for filas.Next() {
		var r models.Referido
		var parentID *int

		if err := filas.Scan(&r.ID, &r.Nombre, &r.Nivel, &r.Ventas, &parentID); err != nil {
			return nil, fmt.Errorf("leyendo referido: %w", err)
		}

		nodo := r
		nodos[r.ID] = &nodo
		padres[r.ID] = parentID
		orden = append(orden, r.ID)
	}
	if err := filas.Err(); err != nil {
		return nil, fmt.Errorf("consultando referidos: %w", err)
	}

	// Enlaza cada nodo con su padre; el que no tiene padre es la raíz del árbol.
	var raiz *models.Referido
	for _, id := range orden {
		parentID := padres[id]
		if parentID == nil {
			raiz = nodos[id]
			continue
		}
		if padre, ok := nodos[*parentID]; ok {
			padre.Hijos = append(padre.Hijos, nodos[id])
		}
	}

	if raiz == nil {
		return nil, ErrNoEncontrado
	}

	return raiz, nil
}
