"""Resolución del tenant (inquilino) de cada petición.

Multitenancy por discriminador: todas las tablas del SaaS llevan tenant_id y
cada consulta se filtra por el tenant resuelto aquí. Es el modelo más simple
de los tres habituales (base por tenant, esquema por tenant, discriminador).
"""

from dataclasses import dataclass

import asyncpg
from fastapi import Depends, Header, HTTPException, status

from .db import get_pool
from .security import UsuarioSesion, usuario_actual


@dataclass
class Tenant:
    id: int
    slug: str
    nombre: str
    plan: str


async def tenant_actual(
    x_tenant: str | None = Header(default=None, alias="X-Tenant"),
    usuario: UsuarioSesion = Depends(usuario_actual),
    pool: asyncpg.Pool = Depends(get_pool),
) -> Tenant:
    """Resuelve el tenant desde la cabecera X-Tenant.

    Sin cabecera se usa el tenant al que pertenece el usuario del token, de
    modo que un cliente normal nunca necesita conocer su tenant.
    """
    async with pool.acquire() as conexion:
        if x_tenant:
            fila = await conexion.fetchrow(
                """
                SELECT t.id, t.slug, t.nombre, p.nombre AS plan
                FROM tenants t
                LEFT JOIN suscripciones s ON s.tenant_id = t.id AND s.estado = 'activa'
                LEFT JOIN planes p ON p.id = s.plan_id
                WHERE t.slug = $1 AND t.activo
                """,
                x_tenant,
            )
        else:
            fila = await conexion.fetchrow(
                """
                SELECT t.id, t.slug, t.nombre, p.nombre AS plan
                FROM tenants t
                JOIN usuarios u ON u.tenant_id = t.id
                LEFT JOIN suscripciones s ON s.tenant_id = t.id AND s.estado = 'activa'
                LEFT JOIN planes p ON p.id = s.plan_id
                WHERE u.id = $1 AND t.activo
                """,
                usuario.id,
            )

    if fila is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail={
                "status": 404,
                "message": "Tenant no encontrado o inactivo",
            },
        )

    return Tenant(
        id=fila["id"],
        slug=fila["slug"],
        nombre=fila["nombre"],
        plan=fila["plan"] or "sin plan",
    )
