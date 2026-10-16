"""Pool de conexiones a PostgreSQL, compartido por toda la aplicación."""

import asyncpg
from fastapi import Request

from .config import get_settings


async def crear_pool() -> asyncpg.Pool:
    """Abre el pool al arrancar el servicio."""
    ajustes = get_settings()
    return await asyncpg.create_pool(
        dsn=ajustes.dsn,
        min_size=1,
        max_size=10,
        command_timeout=10,
    )


async def get_pool(request: Request) -> asyncpg.Pool:
    """Dependencia: entrega el pool guardado en el estado de la aplicación."""
    return request.app.state.pool
