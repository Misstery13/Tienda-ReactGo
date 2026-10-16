"""Administración de inquilinos (tenants)."""

import asyncpg
from fastapi import APIRouter, Depends, HTTPException, status

from ..db import get_pool
from ..schemas import TenantIn, TenantOut
from ..security import UsuarioSesion, requiere_admin

router = APIRouter(prefix="/tenants", tags=["tenants"])


@router.get("", response_model=list[TenantOut])
async def listar_tenants(
    _: UsuarioSesion = Depends(requiere_admin),
    pool: asyncpg.Pool = Depends(get_pool),
) -> list[TenantOut]:
    """Listado de inquilinos. Solo para administradores."""
    async with pool.acquire() as conexion:
        filas = await conexion.fetch(
            """
            SELECT t.id, t.slug, t.nombre, t.activo, p.nombre AS plan
            FROM tenants t
            LEFT JOIN suscripciones s ON s.tenant_id = t.id AND s.estado = 'activa'
            LEFT JOIN planes p ON p.id = s.plan_id
            ORDER BY t.id
            """
        )
    return [TenantOut(**dict(f)) for f in filas]


@router.post("", response_model=TenantOut, status_code=status.HTTP_201_CREATED)
async def crear_tenant(
    datos: TenantIn,
    _: UsuarioSesion = Depends(requiere_admin),
    pool: asyncpg.Pool = Depends(get_pool),
) -> TenantOut:
    """Alta de un inquilino con su suscripción inicial.

    Ambas inserciones ocurren en una transacción: un tenant sin suscripción
    sería un registro a medias.
    """
    async with pool.acquire() as conexion:
        existe = await conexion.fetchval(
            "SELECT 1 FROM tenants WHERE slug = $1", datos.slug
        )
        if existe:
            raise HTTPException(
                status_code=status.HTTP_409_CONFLICT,
                detail={"status": 409, "message": f"El slug '{datos.slug}' ya está en uso"},
            )

        plan = await conexion.fetchrow(
            "SELECT id, nombre FROM planes WHERE id = $1", datos.plan_id
        )
        if plan is None:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail={"status": 400, "message": "El plan indicado no existe"},
            )

        async with conexion.transaction():
            tenant = await conexion.fetchrow(
                """
                INSERT INTO tenants (slug, nombre)
                VALUES ($1, $2)
                RETURNING id, slug, nombre, activo
                """,
                datos.slug,
                datos.nombre,
            )
            await conexion.execute(
                """
                INSERT INTO suscripciones (tenant_id, plan_id, estado, inicia, renueva)
                VALUES ($1, $2, 'activa', CURRENT_DATE, CURRENT_DATE + INTERVAL '1 month')
                """,
                tenant["id"],
                datos.plan_id,
            )

    return TenantOut(**dict(tenant), plan=plan["nombre"])
