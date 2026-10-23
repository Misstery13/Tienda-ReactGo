// Registro público de clientes y administración de cuentas.
package controllers

import (
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"

	"multicatalogo-backend/auth"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// LargoMinimoPassword es el mínimo exigido al registrarse.
const LargoMinimoPassword = 6

// UsuarioController recibe sus dependencias por inyección.
type UsuarioController struct {
	repo    repository.DBRepository
	secreto string
}

// NuevoUsuarioController construye el controlador.
func NuevoUsuarioController(repo repository.DBRepository, secreto string) *UsuarioController {
	return &UsuarioController{repo: repo, secreto: secreto}
}

// Registrar crea una cuenta de cliente y su nodo en la red multinivel.
// Devuelve ya la sesión iniciada para que el usuario no tenga que loguearse aparte.
func (uc *UsuarioController) Registrar(c *fiber.Ctx) error {
	var req models.RegistroRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NuevoAPIError(
			fiber.StatusBadRequest, "Cuerpo de petición inválido", err.Error(),
		))
	}

	req.Nombre = strings.TrimSpace(req.Nombre)
	req.Email = strings.TrimSpace(req.Email)
	// El código también puede llegar como ?ref= en la URL del enlace de invitación.
	if req.CodigoReferido == "" {
		req.CodigoReferido = c.Query("ref")
	}

	faltantes := []string{}
	if req.Nombre == "" {
		faltantes = append(faltantes, "nombre")
	}
	if req.Email == "" || !strings.Contains(req.Email, "@") {
		faltantes = append(faltantes, "email")
	}
	if len(req.Password) < LargoMinimoPassword {
		faltantes = append(faltantes, "password")
	}

	if len(faltantes) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(models.NuevoAPIError(
			fiber.StatusBadRequest,
			"Revisa los datos: el nombre y un correo válido son obligatorios y la contraseña necesita al menos 6 caracteres",
			fiber.Map{"campos_invalidos": faltantes},
		))
	}

	// El hash se calcula aquí: la contraseña en claro nunca llega al repositorio.
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("error generando el hash: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NuevoAPIError(
			fiber.StatusInternalServerError, "No se pudo crear la cuenta", nil,
		))
	}

	usuario, err := uc.repo.CrearUsuario(c.Context(), req, string(hash))
	switch {
	case errors.Is(err, repository.ErrEmailRegistrado):
		return c.Status(fiber.StatusConflict).JSON(models.NuevoAPIError(
			fiber.StatusConflict, err.Error(), nil,
		))
	case errors.Is(err, repository.ErrCodigoInvalido):
		return c.Status(fiber.StatusBadRequest).JSON(models.NuevoAPIError(
			fiber.StatusBadRequest, err.Error(), fiber.Map{"codigo": req.CodigoReferido},
		))
	case err != nil:
		log.Printf("error registrando usuario: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NuevoAPIError(
			fiber.StatusInternalServerError, "No se pudo crear la cuenta", nil,
		))
	}

	token, err := auth.GenerarToken(usuario, uc.secreto)
	if err != nil {
		log.Printf("error firmando el token: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NuevoAPIError(
			fiber.StatusInternalServerError, "Cuenta creada, pero no se pudo iniciar la sesión", nil,
		))
	}

	return c.Status(fiber.StatusCreated).JSON(models.LoginResponse{
		Token: token,
		Email: usuario.Email,
		Rol:   usuario.Rol,
	})
}

// Listar devuelve las cuentas paginadas. Solo para administradores.
func (uc *UsuarioController) Listar(c *fiber.Ctx) error {
	pagina := max(c.QueryInt("page", 1), 1)
	limite := c.QueryInt("limit", 20)
	if limite < 1 || limite > 100 {
		limite = 20
	}

	resultado, err := uc.repo.ListarUsuarios(c.Context(), pagina, limite)
	if err != nil {
		log.Printf("error listando usuarios: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NuevoAPIError(
			fiber.StatusInternalServerError, "Error al consultar los usuarios", nil,
		))
	}

	return c.JSON(resultado)
}

// Actualizar cambia el rol o el estado de una cuenta. Solo para administradores.
func (uc *UsuarioController) Actualizar(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NuevoAPIError(
			fiber.StatusBadRequest, "El id del usuario debe ser un número entero",
			fiber.Map{"valor_recibido": c.Params("id")},
		))
	}

	var req models.ActualizarUsuarioRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NuevoAPIError(
			fiber.StatusBadRequest, "Cuerpo de petición inválido", err.Error(),
		))
	}

	if req.Rol != nil && *req.Rol != "admin" && *req.Rol != "cliente" {
		return c.Status(fiber.StatusBadRequest).JSON(models.NuevoAPIError(
			fiber.StatusBadRequest, "El rol debe ser 'admin' o 'cliente'",
			fiber.Map{"valor_recibido": *req.Rol},
		))
	}

	usuario, err := uc.repo.ActualizarUsuario(c.Context(), id, req)
	if errors.Is(err, repository.ErrNoEncontrado) {
		return c.Status(fiber.StatusNotFound).JSON(models.NuevoAPIError(
			fiber.StatusNotFound, "Usuario no encontrado", fiber.Map{"id": id},
		))
	}
	if err != nil {
		log.Printf("error actualizando usuario %d: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NuevoAPIError(
			fiber.StatusInternalServerError, "No se pudo actualizar el usuario", nil,
		))
	}

	return c.JSON(usuario)
}
