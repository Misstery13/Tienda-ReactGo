"""Configuración del servicio leída del entorno (.env)."""

from functools import lru_cache
from urllib.parse import quote

from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    """Variables de entorno del servicio SaaS."""

    model_config = SettingsConfigDict(env_file=".env", extra="ignore")

    # Base de datos: la misma instancia que usa el núcleo transaccional en Go.
    db_host: str = "localhost"
    db_port: int = 5433
    db_user: str = "multicatalogo"
    db_password: str = "multicatalogo"
    db_name: str = "multicatalogo"

    # El mismo secreto que firma los tokens en Go: así un token emitido por el
    # login de Fiber también es válido aquí, sin duplicar la autenticación.
    jwt_secret: str = ""
    jwt_algoritmo: str = "HS256"

    # Orígenes permitidos por CORS (separados por coma).
    cors_origins: str = "http://localhost:5173,http://localhost:5174"

    app_env: str = "dev"
    port: int = 8000

    @property
    def dsn(self) -> str:
        """Cadena de conexión para asyncpg.

        El usuario y la contraseña van codificados: sin esto, una clave
        con "/" o "@" rompe el análisis de la URL.
        """
        usuario = quote(self.db_user, safe="")
        clave = quote(self.db_password, safe="")
        return (
            f"postgresql://{usuario}:{clave}"
            f"@{self.db_host}:{self.db_port}/{self.db_name}"
        )

    @property
    def origenes_cors(self) -> list[str]:
        return [o.strip() for o in self.cors_origins.split(",") if o.strip()]


@lru_cache
def get_settings() -> Settings:
    """Settings cacheadas: se leen una sola vez por proceso."""
    return Settings()
