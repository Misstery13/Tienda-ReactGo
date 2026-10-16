"""Capa SaaS/multitenant de MultiCatálogo (Unidad 3).

Convive con el núcleo transaccional en Go: comparte PostgreSQL y verifica los
mismos tokens JWT, pero se ocupa de lo que no es transaccional — inquilinos,
planes y suscripciones. Se monta bajo /api/saas.
"""

from contextlib import asynccontextmanager

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from fastapi.exceptions import RequestValidationError

from .config import get_settings
from .db import crear_pool
from .routers import health, planes, suscripciones, tenants


@asynccontextmanager
async def ciclo_de_vida(app: FastAPI):
    """Abre el pool al arrancar y lo cierra al terminar."""
    app.state.pool = await crear_pool()
    try:
        yield
    finally:
        await app.state.pool.close()


ajustes = get_settings()

app = FastAPI(
    title="MultiCatálogo SaaS",
    description="Capa multitenant: inquilinos, planes y suscripciones.",
    version="0.1.0",
    lifespan=ciclo_de_vida,
    docs_url="/api/saas/docs",
    openapi_url="/api/saas/openapi.json",
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=ajustes.origenes_cors,
    allow_methods=["GET", "POST", "PUT", "DELETE"],
    allow_headers=["Origin", "Content-Type", "Accept", "Authorization", "X-Tenant"],
)


@app.exception_handler(RequestValidationError)
async def errores_de_validacion(_, exc: RequestValidationError) -> JSONResponse:
    """Traduce los errores de Pydantic al mismo contrato que usa la API en Go."""
    detalles = exc.errors() if ajustes.app_env == "dev" else None
    return JSONResponse(
        status_code=400,
        content={
            "status": 400,
            "message": "Datos de entrada inválidos",
            **({"details": str(detalles)} if detalles else {}),
        },
    )


# Todos los endpoints viven bajo /api/saas para que el proxy inverso de la
# Unidad 4 pueda repartir: /api/tx/* → Go, /api/saas/* → FastAPI.
PREFIJO = "/api/saas"

app.include_router(health.router, prefix=PREFIJO)
app.include_router(planes.router, prefix=PREFIJO)
app.include_router(tenants.router, prefix=PREFIJO)
app.include_router(suscripciones.router, prefix=PREFIJO)
