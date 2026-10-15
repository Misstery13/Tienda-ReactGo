// Procesamiento transaccional del checkout y acreditación de comisiones.
package repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"

	"multicatalogo-backend/models"
)

// Errores de negocio del checkout. El controlador los traduce a códigos HTTP.
var (
	ErrCarritoVacio        = errors.New("el carrito está vacío")
	ErrProductoInexistente = errors.New("el producto no existe")
	ErrStockInsuficiente   = errors.New("stock insuficiente")
)

// TASAS define la comisión según la distancia del ancestro al comprador:
// el primero arriba gana 10 %, el segundo 5 % y el tercero 2 %.
var TASAS = []float64{0.10, 0.05, 0.02}

// CrearPedido guarda el pedido completo dentro de una sola transacción.
// O se escriben el pedido, todos sus items y el descuento de stock, o no se escribe nada.
func (r *PostgresRepository) CrearPedido(ctx context.Context, req models.CheckoutRequest) (models.Pedido, error) {
	if len(req.Items) == 0 {
		return models.Pedido{}, ErrCarritoVacio
	}

	// BEGIN
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return models.Pedido{}, fmt.Errorf("iniciando transacción: %w", err)
	}
	// Si la función retorna antes del Commit, este Rollback deshace todo.
	// Tras un Commit exitoso, el Rollback no tiene efecto.
	defer tx.Rollback(ctx)

	// El id se reserva primero para poder construir el número de pedido.
	var pedidoID int
	if err := tx.QueryRow(ctx, "SELECT nextval('pedidos_id_seq')").Scan(&pedidoID); err != nil {
		return models.Pedido{}, fmt.Errorf("reservando número de pedido: %w", err)
	}
	numero := fmt.Sprintf("MC-%06d", pedidoID)

	const insertarPedido = `
		INSERT INTO pedidos (id, numero, usuario_id, cliente_nombre, cliente_email,
		                     cliente_ciudad, direccion, telefono, metodo_pago, total)
		VALUES ($1, $2, (SELECT id FROM usuarios WHERE LOWER(email) = LOWER($3)),
		        $4, $5, $6, $7, $8, $9, 0)`

	if _, err := tx.Exec(ctx, insertarPedido, pedidoID, numero, req.Email,
		req.Nombre, req.Email, req.Ciudad, req.Direccion, req.Telefono, req.MetodoPago); err != nil {
		return models.Pedido{}, fmt.Errorf("insertando pedido: %w", err)
	}

	// Punto de inyección de falla para la clase: con SIMULAR_FALLA_CHECKOUT=true
	// el pedido principal ya está insertado, pero la transacción se revierte.
	if os.Getenv("SIMULAR_FALLA_CHECKOUT") == "true" {
		return models.Pedido{}, errors.New("falla simulada: caída del sistema antes de procesar los items")
	}

	lineas := make([]models.LineaPedido, 0, len(req.Items))
	var total float64

	for _, item := range req.Items {
		if item.Cantidad <= 0 {
			return models.Pedido{}, fmt.Errorf("%w: cantidad inválida para el producto %d", ErrProductoInexistente, item.ProductoID)
		}

		// El precio se lee de la base, nunca del cliente. FOR UPDATE bloquea la
		// fila para que dos compras simultáneas no vacíen el mismo stock.
		var nombre string
		var precio float64
		err := tx.QueryRow(ctx,
			"SELECT nombre, precio FROM productos WHERE id = $1 FOR UPDATE",
			item.ProductoID).Scan(&nombre, &precio)
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Pedido{}, fmt.Errorf("%w: id %d", ErrProductoInexistente, item.ProductoID)
		}
		if err != nil {
			return models.Pedido{}, fmt.Errorf("consultando producto %d: %w", item.ProductoID, err)
		}

		// El descuento solo ocurre si hay existencias suficientes.
		etiqueta, err := tx.Exec(ctx,
			"UPDATE productos SET stock = stock - $1 WHERE id = $2 AND stock >= $1",
			item.Cantidad, item.ProductoID)
		if err != nil {
			return models.Pedido{}, fmt.Errorf("descontando stock: %w", err)
		}
		if etiqueta.RowsAffected() == 0 {
			return models.Pedido{}, fmt.Errorf("%w para %q", ErrStockInsuficiente, nombre)
		}

		subtotal := precio * float64(item.Cantidad)
		total += subtotal

		if _, err := tx.Exec(ctx, `
			INSERT INTO pedido_items (pedido_id, producto_id, cantidad, precio_unitario, subtotal)
			VALUES ($1, $2, $3, $4, $5)`,
			pedidoID, item.ProductoID, item.Cantidad, precio, subtotal); err != nil {
			return models.Pedido{}, fmt.Errorf("insertando item: %w", err)
		}

		lineas = append(lineas, models.LineaPedido{
			ProductoID: item.ProductoID,
			Nombre:     nombre,
			Cantidad:   item.Cantidad,
			Precio:     precio,
			Subtotal:   subtotal,
		})
	}

	// El total se calcula en el servidor con los precios reales.
	if _, err := tx.Exec(ctx, "UPDATE pedidos SET total = $1 WHERE id = $2", total, pedidoID); err != nil {
		return models.Pedido{}, fmt.Errorf("actualizando total: %w", err)
	}

	// COMMIT
	if err := tx.Commit(ctx); err != nil {
		return models.Pedido{}, fmt.Errorf("confirmando transacción: %w", err)
	}

	return models.Pedido{
		ID:     pedidoID,
		Numero: numero,
		Cliente: models.ClientePedido{
			Nombre: req.Nombre,
			Email:  req.Email,
			Ciudad: req.Ciudad,
		},
		Items: lineas,
		Total: total,
		Fecha: time.Now().Format("02/01/2006, 15:04:05"),
	}, nil
}

// AcreditarComisiones reparte el total de una compra hacia arriba en el árbol.
// Se ejecuta en una goroutine: el cliente ya recibió su confirmación.
func (r *PostgresRepository) AcreditarComisiones(ctx context.Context, email string, total float64) error {
	// Nodo del comprador dentro de la red.
	var nodoID int
	var padreID *int
	err := r.db.QueryRow(ctx, `
		SELECT re.id, re.parent_id
		FROM referidos re
		JOIN usuarios u ON u.id = re.usuario_id
		WHERE LOWER(u.email) = LOWER($1)`, email).Scan(&nodoID, &padreID)
	if errors.Is(err, pgx.ErrNoRows) {
		// Quien compra sin pertenecer a la red no genera comisiones.
		return nil
	}
	if err != nil {
		return fmt.Errorf("buscando al comprador en la red: %w", err)
	}

	// La venta se suma al propio comprador: así sube "Ventas de la Red".
	if _, err := r.db.Exec(ctx,
		"UPDATE referidos SET ventas = ventas + $1 WHERE id = $2", total, nodoID); err != nil {
		return fmt.Errorf("registrando la venta: %w", err)
	}

	// Se recorre hacia arriba hasta tres ancestros, uno por cada tasa.
	actual := padreID
	for nivel := 0; nivel < len(TASAS) && actual != nil; nivel++ {
		comision := total * TASAS[nivel]

		var siguiente *int
		err := r.db.QueryRow(ctx, `
			UPDATE referidos SET comisiones = comisiones + $1
			WHERE id = $2
			RETURNING parent_id`, comision, *actual).Scan(&siguiente)
		if err != nil {
			return fmt.Errorf("acreditando comisión al nodo %d: %w", *actual, err)
		}

		actual = siguiente
	}

	return nil
}
