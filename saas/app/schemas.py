"""Esquemas de entrada y salida (Pydantic)."""

from datetime import date

from pydantic import BaseModel, Field


class PlanOut(BaseModel):
    id: int
    nombre: str
    precio_mensual: float
    max_usuarios: int
    max_productos: int


class TenantOut(BaseModel):
    id: int
    slug: str
    nombre: str
    activo: bool
    plan: str | None = None


class TenantIn(BaseModel):
    """Datos para dar de alta un inquilino nuevo."""

    slug: str = Field(min_length=3, max_length=40, pattern=r"^[a-z0-9-]+$")
    nombre: str = Field(min_length=3, max_length=150)
    plan_id: int


class SuscripcionOut(BaseModel):
    tenant: str
    plan: str
    estado: str
    inicia: date
    renueva: date
    uso: dict[str, int]
    limites: dict[str, int]
