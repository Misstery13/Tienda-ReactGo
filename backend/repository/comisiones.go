// Consulta del ledger de comisiones y del subárbol propio de cada usuario.
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"multicatalogo-backend/models"
)

// ObtenerRedDeUsuario devuelve el subárbol que cuelga del nodo del usuario,
// no el árbol global. Cada distribuidor ve solo su propia red.
func (r *PostgresRepository) ObtenerRedDeUsuario(ctx context.Context, usuarioID int) (*models.Referido, error) {
	// CTE recursiva: parte del nodo del usuario y baja por parent_id.
	const consulta = `
		WITH RECURSIVE subarbol AS (
			SELECT re.id, re.nombre, re.nivel, re.ventas, re.parent_id, re.codigo, re.comisiones
			FROM referidos re
			JOIN usuarios u ON u.referido_id = re.id
			WHERE u.id = $1

			UNION ALL

			SELECT hijo.id, hijo.nombre, hijo.nivel, hijo.ventas, hijo.parent_id, hijo.codigo, hijo.comisiones
			FROM referidos hijo
			JOIN subarbol padre ON hijo.parent_id = padre.id
		)
		SELECT id, nombre, nivel, ventas, parent_id, codigo, comisiones
		FROM subarbol
		ORDER BY nivel, id`

	filas, err := r.db.Query(ctx, consulta, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("consultando la red del usuario: %w", err)
	}
	defer filas.Close()

	nodos := map[int]*models.Referido{}
	padres := map[int]*int{}
	orden := []int{}

	for filas.Next() {
		var ref models.Referido
		var parentID *int

		if err := filas.Scan(&ref.ID, &ref.Nombre, &ref.Nivel, &ref.Ventas,
			&parentID, &ref.Codigo, &ref.Comisiones); err != nil {
			return nil, fmt.Errorf("leyendo referido: %w", err)
		}

		nodo := ref
		nodos[ref.ID] = &nodo
		padres[ref.ID] = parentID
		orden = append(orden, ref.ID)
	}
	if err := filas.Err(); err != nil {
		return nil, fmt.Errorf("consultando la red del usuario: %w", err)
	}

	if len(orden) == 0 {
		return nil, ErrNoEncontrado
	}

	// El primero de la lista es la raíz del subárbol: el propio usuario.
	raiz := nodos[orden[0]]
	for _, id := range orden[1:] {
		if parentID := padres[id]; parentID != nil {
			if padre, ok := nodos[*parentID]; ok {
				padre.Hijos = append(padre.Hijos, nodos[id])
			}
		}
	}

	// Los niveles se vuelven relativos: la raíz es 0 desde su propio punto de vista.
	desplazar(raiz, raiz.Nivel)

	return raiz, nil
}

// desplazar vuelve relativos los niveles del subárbol.
func desplazar(nodo *models.Referido, base int) {
	nodo.Nivel -= base
	for _, hijo := range nodo.Hijos {
		desplazar(hijo, base)
	}
}

// ListarComisionesDeUsuario devuelve el ledger del usuario con sus totales.
func (r *PostgresRepository) ListarComisionesDeUsuario(ctx context.Context, usuarioID int) (models.ResumenComisiones, error) {
	var referidoID int
	err := r.db.QueryRow(ctx,
		"SELECT referido_id FROM usuarios WHERE id = $1 AND referido_id IS NOT NULL",
		usuarioID).Scan(&referidoID)
	if errors.Is(err, pgx.ErrNoRows) {
		// Usuario sin nodo en la red: ledger vacío, no es un error.
		return models.ResumenComisiones{Lineas: []models.LineaComision{}}, nil
	}
	if err != nil {
		return models.ResumenComisiones{}, fmt.Errorf("buscando el nodo del usuario: %w", err)
	}

	filas, err := r.db.Query(ctx, `
		SELECT c.id, p.numero, c.nivel, c.porcentaje, c.monto, c.estado,
		       to_char(c.creado_en, 'DD/MM/YYYY')
		FROM comisiones c
		JOIN pedidos p ON p.id = c.pedido_id
		WHERE c.referido_id = $1
		ORDER BY c.id DESC`, referidoID)
	if err != nil {
		return models.ResumenComisiones{}, fmt.Errorf("consultando comisiones: %w", err)
	}
	defer filas.Close()

	resumen := models.ResumenComisiones{Lineas: []models.LineaComision{}}
	for filas.Next() {
		var l models.LineaComision
		if err := filas.Scan(&l.ID, &l.Pedido, &l.Nivel, &l.Porcentaje,
			&l.Monto, &l.Estado, &l.Fecha); err != nil {
			return models.ResumenComisiones{}, fmt.Errorf("leyendo comisión: %w", err)
		}

		// Los totales se acumulan por estado para el panel.
		switch l.Estado {
		case "pendiente", "aprobada":
			resumen.Pendiente += l.Monto
		case "pagada":
			resumen.Pagada += l.Monto
		case "revertida":
			resumen.Revertida += l.Monto
		}

		resumen.Lineas = append(resumen.Lineas, l)
	}

	return resumen, filas.Err()
}
