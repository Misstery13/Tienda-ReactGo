"""Verificación de los JWT emitidos por el núcleo en Go.

El servicio SaaS no emite tokens: confía en los que firma el login de Fiber
usando el mismo secreto. Así no se duplica la gestión de credenciales.
"""

from dataclasses import dataclass

import jwt
from fastapi import Depends, HTTPException, Request, status

from .config import Settings, get_settings


@dataclass
class UsuarioSesion:
    """Identidad extraída del token, no del cuerpo de la petición."""

    id: int
    email: str
    rol: str


def _error(codigo: int, mensaje: str) -> HTTPException:
    """Error con el mismo contrato JSON que usa la API en Go."""
    return HTTPException(
        status_code=codigo,
        detail={"status": codigo, "message": mensaje},
    )


def usuario_actual(
    request: Request,
    ajustes: Settings = Depends(get_settings),
) -> UsuarioSesion:
    """Dependencia que exige un JWT válido en la cabecera Authorization."""
    cabecera = request.headers.get("Authorization", "")
    partes = cabecera.split(" ", 1)

    if len(partes) != 2 or partes[0].lower() != "bearer":
        raise _error(status.HTTP_401_UNAUTHORIZED, "Falta el token de autenticación")

    try:
        claims = jwt.decode(
            partes[1].strip(),
            ajustes.jwt_secret,
            algorithms=[ajustes.jwt_algoritmo],
        )
    except jwt.PyJWTError as exc:
        raise _error(status.HTTP_401_UNAUTHORIZED, "Token inválido o expirado") from exc

    return UsuarioSesion(
        id=int(claims.get("sub", 0)),
        email=claims.get("email", ""),
        rol=claims.get("rol", ""),
    )


def requiere_admin(
    usuario: UsuarioSesion = Depends(usuario_actual),
) -> UsuarioSesion:
    """Dependencia que además exige el rol admin."""
    if usuario.rol != "admin":
        raise _error(
            status.HTTP_403_FORBIDDEN,
            "Esta sección es solo para administradores",
        )
    return usuario
