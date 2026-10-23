// Consulta y gestión del ciclo de vida de los pedidos.
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"multicatalogo-backend/domain"
	"multicatalogo-backend/models"
)

// Errores de negocio de los pedidos.
var (
	ErrEstadoInvalido     = errors.New("el estado indicado no existe")
	ErrTransicionInvalida = errors.New("no se puede pasar a ese estado")
	ErrPedidoAjeno        = errors.New("el pedido pertenece a otra cuenta")
)

// consultaResumen es la base de los listados; el WHERE lo pone cada caso.
const consultaResumen = `
	SELECT p.id, p.numero, p.cliente_nombre, p.cliente_email, p.total, p.estado,
	       to_char(p.creado_en, 'DD/MM/YYYY HH24:MI'),
	       (SELECT coalesce(sum(cantidad), 0) FROM pedido_items WHERE pedido_id = p.id)
	FROM pedidos p`

// ListarPedidosDeUsuario devuelve los pedidos de una cuenta, paginados.
func (r *PostgresRepository) ListarPedidosDeUsuario(ctx context.Context, usuarioID, pagina, limite int) (models.Pagina[models.ResumenPedido], error) {
	var total int
	if err := r.db.QueryRow(ctx,
		"SELECT count(*) FROM pedidos WHERE usuario_id = $1", usuarioID).Scan(&total); err != nil {
		return models.Pagina[models.ResumenPedido]{}, fmt.Errorf("contando pedidos: %w", err)
	}

	filas, err := r.db.Query(ctx, consultaResumen+`
		WHERE p.usuario_id = $1
		ORDER BY p.id DESC
		LIMIT $2 OFFSET $3`, usuarioID, limite, (pagina-1)*limite)
	if err != nil {
		return models.Pagina[models.ResumenPedido]{}, fmt.Errorf("listando pedidos: %w", err)
	}
	defer filas.Close()

	return paginaDeResumenes(filas, total, pagina, limite)
}

// ListarPedidos devuelve todos los pedidos, con filtro opcional por estado.
func (r *PostgresRepository) ListarPedidos(ctx context.Context, estado string, pagina, limite int) (models.Pagina[models.ResumenPedido], error) {
	// Un estado vacío ($1 = NULL) desactiva el filtro.
	var filtro *string
	if estado != "" {
		filtro = &estado
	}

	var total int
	if err := r.db.QueryRow(ctx,
		"SELECT count(*) FROM pedidos WHERE ($1::text IS NULL OR estado = $1)", filtro).Scan(&total); err != nil {
		return models.Pagina[models.ResumenPedido]{}, fmt.Errorf("contando pedidos: %w", err)
	}

	filas, err := r.db.Query(ctx, consultaResumen+`
		WHERE ($1::text IS NULL OR p.estado = $1)
		ORDER BY p.id DESC
		LIMIT $2 OFFSET $3`, filtro, limite, (pagina-1)*limite)
	if err != nil {
		return models.Pagina[models.ResumenPedido]{}, fmt.Errorf("listando pedidos: %w", err)
	}
	defer filas.Close()

	return paginaDeResumenes(filas, total, pagina, limite)
}

// paginaDeResumenes arma la página a partir de las filas ya consultadas.
func paginaDeResumenes(filas pgx.Rows, total, pagina, limite int) (models.Pagina[models.ResumenPedido], error) {
	pedidos := []models.ResumenPedido{}
	for filas.Next() {
		var p models.ResumenPedido
		if err := filas.Scan(&p.ID, &p.Numero, &p.Cliente, &p.Email, &p.Total,
			&p.Estado, &p.Fecha, &p.Articulos); err != nil {
			return models.Pagina[models.ResumenPedido]{}, fmt.Errorf("leyendo pedido: %w", err)
		}
		pedidos = append(pedidos, p)
	}

	return models.Pagina[models.ResumenPedido]{
		Datos: pedidos, Total: total, Pagina: pagina, Limite: limite,
	}, filas.Err()
}

// ObtenerPedido devuelve la cabecera y el detalle de un pedido.
// Si usuarioID es distinto de cero, solo lo entrega a su dueño.
func (r *PostgresRepository) ObtenerPedido(ctx context.Context, id, usuarioID int) (models.Pedido, error) {
	var p models.Pedido
	var duenoID *int

	err := r.db.QueryRow(ctx, `
		SELECT id, numero, cliente_nombre, cliente_email, cliente_ciudad, total, estado,
		       to_char(creado_en, 'DD/MM/YYYY, HH24:MI:SS'), usuario_id
		FROM pedidos WHERE id = $1`, id).
		Scan(&p.ID, &p.Numero, &p.Cliente.Nombre, &p.Cliente.Email, &p.Cliente.Ciudad,
			&p.Total, &p.Estado, &p.Fecha, &duenoID)

	if errors.Is(err, pgx.ErrNoRows) {
		return models.Pedido{}, ErrNoEncontrado
	}
	if err != nil {
		return models.Pedido{}, fmt.Errorf("consultando pedido: %w", err)
	}

	// El cliente solo puede ver lo suyo; el admin pasa usuarioID = 0.
	if usuarioID != 0 && (duenoID == nil || *duenoID != usuarioID) {
		return models.Pedido{}, ErrPedidoAjeno
	}

	filas, err := r.db.Query(ctx, `
		SELECT pi.producto_id, pr.nombre, pi.cantidad, pi.precio_unitario, pi.subtotal
		FROM pedido_items pi
		JOIN productos pr ON pr.id = pi.producto_id
		WHERE pi.pedido_id = $1
		ORDER BY pi.id`, id)
	if err != nil {
		return models.Pedido{}, fmt.Errorf("consultando el detalle: %w", err)
	}
	defer filas.Close()

	p.Items = []models.LineaPedido{}
	for filas.Next() {
		var l models.LineaPedido
		if err := filas.Scan(&l.ProductoID, &l.Nombre, &l.Cantidad, &l.Precio, &l.Subtotal); err != nil {
			return models.Pedido{}, fmt.Errorf("leyendo el detalle: %w", err)
		}
		p.Items = append(p.Items, l)
	}

	return p, filas.Err()
}

// CambiarEstadoPedido aplica una transición válida del flujo.
// Cancelar es la operación espejo del checkout: devuelve el stock reservado.
func (r *PostgresRepository) CambiarEstadoPedido(ctx context.Context, id int, nuevo string) (models.Pedido, error) {
	if !domain.EstadoValido(nuevo) {
		return models.Pedido{}, ErrEstadoInvalido
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return models.Pedido{}, fmt.Errorf("iniciando transacción: %w", err)
	}
	defer tx.Rollback(ctx)

	// FOR UPDATE bloquea el pedido: dos administradores no pueden cambiarlo a la vez.
	var actual string
	err = tx.QueryRow(ctx, "SELECT estado FROM pedidos WHERE id = $1 FOR UPDATE", id).Scan(&actual)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Pedido{}, ErrNoEncontrado
	}
	if err != nil {
		return models.Pedido{}, fmt.Errorf("consultando el estado: %w", err)
	}

	if !domain.TransicionPermitida(actual, nuevo) {
		return models.Pedido{}, fmt.Errorf("%w: de %q a %q", ErrTransicionInvalida, actual, nuevo)
	}

	if _, err := tx.Exec(ctx, "UPDATE pedidos SET estado = $2 WHERE id = $1", id, nuevo); err != nil {
		return models.Pedido{}, fmt.Errorf("actualizando el estado: %w", err)
	}

	// Al cancelar, cada artículo vuelve al stock dentro de la misma transacción.
	if nuevo == domain.EstadoCancelado {
		if _, err := tx.Exec(ctx, `
			UPDATE productos pr
			SET stock = pr.stock + pi.cantidad
			FROM pedido_items pi
			WHERE pi.pedido_id = $1 AND pr.id = pi.producto_id`, id); err != nil {
			return models.Pedido{}, fmt.Errorf("devolviendo el stock: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return models.Pedido{}, fmt.Errorf("confirmando el cambio: %w", err)
	}

	return r.ObtenerPedido(ctx, id, 0)
}
