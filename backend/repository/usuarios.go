// Registro de cuentas y administración de usuarios.
package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/jackc/pgx/v5"

	"multicatalogo-backend/models"
)

// GenerarCodigoReferido arma el código de invitación de un nodo:
// las 8 primeras letras del nombre, sin tildes ni espacios, más su id.
// Es una función pura, así que puede probarse sin base de datos.
func GenerarCodigoReferido(nombre string, id int) string {
	sinTildes := map[rune]rune{
		'á': 'a', 'é': 'e', 'í': 'i', 'ó': 'o', 'ú': 'u', 'ñ': 'n', 'ü': 'u',
		'Á': 'A', 'É': 'E', 'Í': 'I', 'Ó': 'O', 'Ú': 'U', 'Ñ': 'N', 'Ü': 'U',
	}

	var letras []rune
	for _, r := range nombre {
		if reemplazo, ok := sinTildes[r]; ok {
			r = reemplazo
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			letras = append(letras, unicode.ToUpper(r))
		}
		if len(letras) == 8 {
			break
		}
	}

	if len(letras) == 0 {
		letras = []rune("USUARIO")
	}

	return fmt.Sprintf("%s-%03d", string(letras), id)
}

// Errores de negocio del registro.
var (
	ErrEmailRegistrado = errors.New("ya existe una cuenta con ese correo")
	ErrCodigoInvalido  = errors.New("el código de referido no existe")
	ErrUsuarioInactivo = errors.New("la cuenta está desactivada")
)

// CrearUsuario da de alta una cuenta y su nodo en la red, todo en una transacción.
// El hash ya viene calculado: el repositorio nunca ve la contraseña en claro.
func (r *PostgresRepository) CrearUsuario(ctx context.Context, datos models.RegistroRequest, hash string) (models.Usuario, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return models.Usuario{}, fmt.Errorf("iniciando transacción: %w", err)
	}
	defer tx.Rollback(ctx)

	// El correo es único: se comprueba antes para dar un error claro.
	var existe bool
	if err := tx.QueryRow(ctx,
		"SELECT EXISTS (SELECT 1 FROM usuarios WHERE LOWER(email) = LOWER($1))",
		datos.Email).Scan(&existe); err != nil {
		return models.Usuario{}, fmt.Errorf("verificando el correo: %w", err)
	}
	if existe {
		return models.Usuario{}, ErrEmailRegistrado
	}

	// El código de invitación determina de quién cuelga el nuevo distribuidor.
	// Sin código, el nodo cuelga de la raíz de la red.
	var padreID, padreNivel int
	codigo := strings.TrimSpace(datos.CodigoReferido)

	if codigo != "" {
		err := tx.QueryRow(ctx,
			"SELECT id, nivel FROM referidos WHERE UPPER(codigo) = UPPER($1)",
			codigo).Scan(&padreID, &padreNivel)
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Usuario{}, ErrCodigoInvalido
		}
		if err != nil {
			return models.Usuario{}, fmt.Errorf("buscando el código de referido: %w", err)
		}
	} else {
		if err := tx.QueryRow(ctx,
			"SELECT id, nivel FROM referidos WHERE parent_id IS NULL ORDER BY id LIMIT 1").
			Scan(&padreID, &padreNivel); err != nil {
			return models.Usuario{}, fmt.Errorf("buscando la raíz de la red: %w", err)
		}
	}

	// El id se reserva primero para poder construir el código de invitación.
	var referidoID int
	if err := tx.QueryRow(ctx, "SELECT nextval('referidos_id_seq')").Scan(&referidoID); err != nil {
		return models.Usuario{}, fmt.Errorf("reservando el nodo de la red: %w", err)
	}

	// Nodo del nuevo distribuidor, un nivel por debajo de quien lo invitó.
	if _, err := tx.Exec(ctx, `
		INSERT INTO referidos (id, nombre, nivel, ventas, parent_id, codigo, tenant_id)
		VALUES ($1, $2, $3, 0, $4, $5,
		        (SELECT tenant_id FROM referidos WHERE id = $4))`,
		referidoID, datos.Nombre, padreNivel+1, padreID,
		GenerarCodigoReferido(datos.Nombre, referidoID)); err != nil {
		return models.Usuario{}, fmt.Errorf("creando el nodo de la red: %w", err)
	}

	// La cuenta queda enlazada a su nodo por clave foránea.
	var u models.Usuario
	if err := tx.QueryRow(ctx, `
		INSERT INTO usuarios (email, password, rol, nombre, referido_id, tenant_id)
		VALUES (LOWER($1), $2, 'cliente', $3, $4,
		        (SELECT tenant_id FROM referidos WHERE id = $4))
		RETURNING id, email, rol, nombre, activo, referido_id`,
		datos.Email, hash, datos.Nombre, referidoID).
		Scan(&u.ID, &u.Email, &u.Rol, &u.Nombre, &u.Activo, &u.ReferidoID); err != nil {
		return models.Usuario{}, fmt.Errorf("creando la cuenta: %w", err)
	}

	// El nodo apunta de vuelta al usuario.
	if _, err := tx.Exec(ctx,
		"UPDATE referidos SET usuario_id = $1 WHERE id = $2", u.ID, referidoID); err != nil {
		return models.Usuario{}, fmt.Errorf("enlazando el nodo con la cuenta: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return models.Usuario{}, fmt.Errorf("confirmando el registro: %w", err)
	}

	return u, nil
}

// ListarUsuarios devuelve una página del listado de cuentas.
func (r *PostgresRepository) ListarUsuarios(ctx context.Context, pagina, limite int) (models.Pagina[models.Usuario], error) {
	var total int
	if err := r.db.QueryRow(ctx, "SELECT count(*) FROM usuarios").Scan(&total); err != nil {
		return models.Pagina[models.Usuario]{}, fmt.Errorf("contando usuarios: %w", err)
	}

	filas, err := r.db.Query(ctx, `
		SELECT id, email, rol, nombre, activo, referido_id, to_char(creado_en, 'DD/MM/YYYY')
		FROM usuarios
		ORDER BY id
		LIMIT $1 OFFSET $2`, limite, (pagina-1)*limite)
	if err != nil {
		return models.Pagina[models.Usuario]{}, fmt.Errorf("listando usuarios: %w", err)
	}
	defer filas.Close()

	usuarios := []models.Usuario{}
	for filas.Next() {
		var u models.Usuario
		if err := filas.Scan(&u.ID, &u.Email, &u.Rol, &u.Nombre, &u.Activo, &u.ReferidoID, &u.CreadoEn); err != nil {
			return models.Pagina[models.Usuario]{}, fmt.Errorf("leyendo usuario: %w", err)
		}
		usuarios = append(usuarios, u)
	}

	return models.Pagina[models.Usuario]{
		Datos:  usuarios,
		Total:  total,
		Pagina: pagina,
		Limite: limite,
	}, filas.Err()
}

// ActualizarUsuario cambia el rol o el estado de una cuenta.
func (r *PostgresRepository) ActualizarUsuario(ctx context.Context, id int, datos models.ActualizarUsuarioRequest) (models.Usuario, error) {
	var u models.Usuario
	err := r.db.QueryRow(ctx, `
		UPDATE usuarios
		SET rol    = COALESCE($2, rol),
		    activo = COALESCE($3, activo)
		WHERE id = $1
		RETURNING id, email, rol, nombre, activo, referido_id`,
		id, datos.Rol, datos.Activo).
		Scan(&u.ID, &u.Email, &u.Rol, &u.Nombre, &u.Activo, &u.ReferidoID)

	if errors.Is(err, pgx.ErrNoRows) {
		return models.Usuario{}, ErrNoEncontrado
	}
	if err != nil {
		return models.Usuario{}, fmt.Errorf("actualizando usuario: %w", err)
	}

	return u, nil
}
