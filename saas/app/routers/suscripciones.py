"""Suscripción del inquilino actual y su consumo frente a los límites del plan."""

import asyncpg
from fastapi import APIRouter, Depends, HTTPException, status

from ..db import get_pool
from ..schemas import SuscripcionOut
from ..tenancy import Tenant, tenant_actual

router = APIRouter(tags=["suscripciones"])


@router.get("/mi-suscripcion", response_model=SuscripcionOut)
async def mi_suscripcion(
    tenant: Tenant = Depends(tenant_actual),
    pool: asyncpg.Pool = Depends(get_pool),
) -> SuscripcionOut:
    """Devuelve el plan vigente y cuánto lleva consumido el tenant."""
    async with pool.acquire() as conexion:
        fila = await conexion.fetchrow(
            """
            SELECT p.nombre AS plan, s.estado, s.inicia, s.renueva,
                   p.max_usuarios, p.max_productos,
                   (SELECT count(*) FROM usuarios  WHERE tenant_id = $1) AS usuarios,
                   (SELECT count(*) FROM productos WHERE tenant_id = $1) AS productos
            FROM suscripciones s
            JOIN planes p ON p.id = s.plan_id
            WHERE s.tenant_id = $1 AND s.estado = 'activa'
            """,
            tenant.id,
        )

    if fila is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail={"status": 404, "message": "El tenant no tiene una suscripción activa"},
        )

    return SuscripcionOut(
        tenant=tenant.nombre,
        plan=fila["plan"],
        estado=fila["estado"],
        inicia=fila["inicia"],
        renueva=fila["renueva"],
        uso={"usuarios": fila["usuarios"], "productos": fila["productos"]},
        limites={
            "usuarios": fila["max_usuarios"],
            "productos": fila["max_productos"],
        },
    )
