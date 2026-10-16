"""Catálogo de planes del SaaS."""

import asyncpg
from fastapi import APIRouter, Depends

from ..db import get_pool
from ..schemas import PlanOut

router = APIRouter(prefix="/planes", tags=["planes"])


@router.get("", response_model=list[PlanOut])
async def listar_planes(pool: asyncpg.Pool = Depends(get_pool)) -> list[PlanOut]:
    """Los planes son públicos: se muestran antes de registrarse."""
    async with pool.acquire() as conexion:
        filas = await conexion.fetch(
            """
            SELECT id, nombre, precio_mensual, max_usuarios, max_productos
            FROM planes
            ORDER BY precio_mensual
            """
        )
    return [PlanOut(**dict(f)) for f in filas]
