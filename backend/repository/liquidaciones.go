// Liquidaciones: pago de las comisiones pendientes del ledger.
package repository

import (
	"context"
	"errors"
	"fmt"

	"multicatalogo-backend/models"
)

// ErrSinDeuda indica que el distribuidor no tiene comisiones por cobrar.
var ErrSinDeuda = errors.New("el distribuidor no tiene comisiones pendientes")

// DeudaPorDistribuidor suma lo que debe pagarse a cada referido.
func (r *PostgresRepository) DeudaPorDistribuidor(ctx context.Context) ([]models.DeudaDistribuidor, error) {
	filas, err := r.db.Query(ctx, `
		SELECT re.id, re.nombre, re.codigo, sum(c.monto), count(*)
		FROM comisiones c
		JOIN referidos re ON re.id = c.referido_id
		WHERE c.estado IN ('pendiente', 'aprobada')
		GROUP BY re.id, re.nombre, re.codigo
		ORDER BY sum(c.monto) DESC`)
	if err != nil {
		return nil, fmt.Errorf("consultando la deuda: %w", err)
	}
	defer filas.Close()

	deudas := []models.DeudaDistribuidor{}
	for filas.Next() {
		var d models.DeudaDistribuidor
		if err := filas.Scan(&d.ReferidoID, &d.Nombre, &d.Codigo, &d.Pendiente, &d.Lineas); err != nil {
			return nil, fmt.Errorf("leyendo deuda: %w", err)
		}
		deudas = append(deudas, d)
	}

	return deudas, filas.Err()
}

// Liquidar paga de una vez todas las comisiones pendientes de un referido.
// Es idempotente: si no hay nada pendiente, no crea una liquidación vacía.
func (r *PostgresRepository) Liquidar(ctx context.Context, referidoID int) (models.Liquidacion, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return models.Liquidacion{}, fmt.Errorf("iniciando transacción: %w", err)
	}
	defer tx.Rollback(ctx)

	// Las filas se bloquean dentro de la CTE y se suman fuera: PostgreSQL no
	// permite FOR UPDATE junto a funciones de agregado. El bloqueo evita que
	// dos liquidaciones simultáneas paguen la misma comisión dos veces.
	var total float64
	var lineas int
	if err := tx.QueryRow(ctx, `
		WITH bloqueadas AS (
			SELECT monto
			FROM comisiones
			WHERE referido_id = $1 AND estado IN ('pendiente', 'aprobada')
			FOR UPDATE
		)
		SELECT coalesce(sum(monto), 0), count(*) FROM bloqueadas`,
		referidoID).Scan(&total, &lineas); err != nil {
		return models.Liquidacion{}, fmt.Errorf("sumando comisiones: %w", err)
	}

	if lineas == 0 || total <= 0 {
		return models.Liquidacion{}, ErrSinDeuda
	}

	var liq models.Liquidacion
	if err := tx.QueryRow(ctx, `
		INSERT INTO liquidaciones (referido_id, total, estado, pagada_en)
		VALUES ($1, $2, 'pagada', now())
		RETURNING id, referido_id, total, estado, to_char(creado_en, 'DD/MM/YYYY HH24:MI')`,
		referidoID, total).
		Scan(&liq.ID, &liq.ReferidoID, &liq.Total, &liq.Estado, &liq.Fecha); err != nil {
		return models.Liquidacion{}, fmt.Errorf("creando la liquidación: %w", err)
	}

	// Las comisiones quedan pagadas y atadas a esta liquidación.
	etiqueta, err := tx.Exec(ctx, `
		UPDATE comisiones SET estado = 'pagada', liquidacion_id = $2
		WHERE referido_id = $1 AND estado IN ('pendiente', 'aprobada')`,
		referidoID, liq.ID)
	if err != nil {
		return models.Liquidacion{}, fmt.Errorf("marcando comisiones pagadas: %w", err)
	}
	liq.Comisiones = int(etiqueta.RowsAffected())

	if err := tx.QueryRow(ctx,
		"SELECT nombre FROM referidos WHERE id = $1", referidoID).Scan(&liq.Nombre); err != nil {
		return models.Liquidacion{}, fmt.Errorf("leyendo el nombre: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return models.Liquidacion{}, fmt.Errorf("confirmando la liquidación: %w", err)
	}

	return liq, nil
}

// ListarLiquidaciones devuelve el historial de pagos.
// Con referidoID distinto de cero, solo las de ese distribuidor.
func (r *PostgresRepository) ListarLiquidaciones(ctx context.Context, referidoID int) ([]models.Liquidacion, error) {
	var filtro *int
	if referidoID != 0 {
		filtro = &referidoID
	}

	filas, err := r.db.Query(ctx, `
		SELECT l.id, l.referido_id, re.nombre, l.total, l.estado,
		       (SELECT count(*) FROM comisiones c WHERE c.liquidacion_id = l.id),
		       to_char(l.creado_en, 'DD/MM/YYYY HH24:MI')
		FROM liquidaciones l
		JOIN referidos re ON re.id = l.referido_id
		WHERE ($1::int IS NULL OR l.referido_id = $1)
		ORDER BY l.id DESC`, filtro)
	if err != nil {
		return nil, fmt.Errorf("consultando liquidaciones: %w", err)
	}
	defer filas.Close()

	liquidaciones := []models.Liquidacion{}
	for filas.Next() {
		var l models.Liquidacion
		if err := filas.Scan(&l.ID, &l.ReferidoID, &l.Nombre, &l.Total,
			&l.Estado, &l.Comisiones, &l.Fecha); err != nil {
			return nil, fmt.Errorf("leyendo liquidación: %w", err)
		}
		liquidaciones = append(liquidaciones, l)
	}

	return liquidaciones, filas.Err()
}

// ReferidoDeUsuario devuelve el nodo de red de una cuenta (0 si no tiene).
func (r *PostgresRepository) ReferidoDeUsuario(ctx context.Context, usuarioID int) (int, error) {
	var referidoID *int
	if err := r.db.QueryRow(ctx,
		"SELECT referido_id FROM usuarios WHERE id = $1", usuarioID).Scan(&referidoID); err != nil {
		return 0, fmt.Errorf("buscando el nodo del usuario: %w", err)
	}
	if referidoID == nil {
		return 0, nil
	}
	return *referidoID, nil
}
