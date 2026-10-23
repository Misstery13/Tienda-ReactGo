// Reportes de negocio: agregaciones SQL sobre pedidos y comisiones.
package repository

import (
	"context"
	"fmt"

	"multicatalogo-backend/models"
)

// rangoFechas convierte las fechas opcionales en parámetros nulos cuando vienen vacías.
func rangoFechas(desde, hasta string) (*string, *string) {
	var d, h *string
	if desde != "" {
		d = &desde
	}
	if hasta != "" {
		h = &hasta
	}
	return d, h
}

// ResumenDeVentas calcula los KPIs del periodo. Los pedidos cancelados
// no cuentan como venta, pero se informan aparte.
func (r *PostgresRepository) ResumenDeVentas(ctx context.Context, desde, hasta string) (models.ResumenVentas, error) {
	d, h := rangoFechas(desde, hasta)

	var res models.ResumenVentas
	err := r.db.QueryRow(ctx, `
		SELECT
			coalesce(sum(total) FILTER (WHERE estado <> 'cancelado'), 0),
			count(*) FILTER (WHERE estado <> 'cancelado'),
			coalesce(avg(total) FILTER (WHERE estado <> 'cancelado'), 0),
			count(*) FILTER (WHERE estado = 'cancelado')
		FROM pedidos
		WHERE ($1::date IS NULL OR creado_en >= $1::date)
		  AND ($2::date IS NULL OR creado_en < $2::date + INTERVAL '1 day')`,
		d, h).Scan(&res.TotalVendido, &res.Pedidos, &res.TicketPromedio, &res.Cancelados)
	if err != nil {
		return models.ResumenVentas{}, fmt.Errorf("calculando el resumen: %w", err)
	}

	// Las comisiones devengadas salen del ledger, no de los contadores.
	if err := r.db.QueryRow(ctx, `
		SELECT coalesce(sum(c.monto), 0)
		FROM comisiones c
		JOIN pedidos p ON p.id = c.pedido_id
		WHERE c.estado <> 'revertida'
		  AND ($1::date IS NULL OR p.creado_en >= $1::date)
		  AND ($2::date IS NULL OR p.creado_en < $2::date + INTERVAL '1 day')`,
		d, h).Scan(&res.ComisionesDevengadas); err != nil {
		return models.ResumenVentas{}, fmt.Errorf("calculando comisiones devengadas: %w", err)
	}

	return res, nil
}

// TopProductos ordena los productos por dinero vendido.
func (r *PostgresRepository) TopProductos(ctx context.Context, limite int) ([]models.FilaReporte, error) {
	return r.ranking(ctx, `
		SELECT pr.id, pr.nombre, sum(pi.cantidad)::int, sum(pi.subtotal)
		FROM pedido_items pi
		JOIN productos pr ON pr.id = pi.producto_id
		JOIN pedidos p ON p.id = pi.pedido_id
		WHERE p.estado <> 'cancelado'
		GROUP BY pr.id, pr.nombre
		ORDER BY sum(pi.subtotal) DESC
		LIMIT $1`, limite)
}

// TopReferidos ordena los distribuidores por comisiones generadas.
func (r *PostgresRepository) TopReferidos(ctx context.Context, limite int) ([]models.FilaReporte, error) {
	return r.ranking(ctx, `
		SELECT re.id, re.nombre, count(*)::int, sum(c.monto)
		FROM comisiones c
		JOIN referidos re ON re.id = c.referido_id
		WHERE c.estado <> 'revertida'
		GROUP BY re.id, re.nombre
		ORDER BY sum(c.monto) DESC
		LIMIT $1`, limite)
}

// ranking ejecuta cualquiera de las consultas de top y arma las filas.
func (r *PostgresRepository) ranking(ctx context.Context, consulta string, limite int) ([]models.FilaReporte, error) {
	filas, err := r.db.Query(ctx, consulta, limite)
	if err != nil {
		return nil, fmt.Errorf("consultando el ranking: %w", err)
	}
	defer filas.Close()

	resultado := []models.FilaReporte{}
	for filas.Next() {
		var f models.FilaReporte
		if err := filas.Scan(&f.ID, &f.Nombre, &f.Cantidad, &f.Total); err != nil {
			return nil, fmt.Errorf("leyendo fila del ranking: %w", err)
		}
		resultado = append(resultado, f)
	}

	return resultado, filas.Err()
}
