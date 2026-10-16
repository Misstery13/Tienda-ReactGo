"""Estado del servicio SaaS."""

import asyncpg
from fastapi import APIRouter, Depends, Response, status

from ..db import get_pool

router = APIRouter(tags=["health"])


@router.get("/health")
async def health(response: Response, pool: asyncpg.Pool = Depends(get_pool)) -> dict:
    """Informa el estado real: sin PostgreSQL el servicio no puede operar."""
    estado_bd = "ok"
    try:
        async with pool.acquire() as conexion:
            await conexion.fetchval("SELECT 1")
    except Exception:
        estado_bd = "caída"

    if estado_bd != "ok":
        response.status_code = status.HTTP_503_SERVICE_UNAVAILABLE
        return {
            "status": "unavailable",
            "message": "Servicio SaaS sin acceso a PostgreSQL",
            "postgresql": estado_bd,
        }

    return {
        "status": "ok",
        "message": "Servicio SaaS (FastAPI) operativo",
        "postgresql": estado_bd,
    }
