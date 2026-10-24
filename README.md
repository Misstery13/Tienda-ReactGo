# Tienda ReactGo (MultiCatálogo)

Tienda multicatálogo con frontend en React, núcleo transaccional en Go/Fiber,
capa SaaS en FastAPI y datos en PostgreSQL 16 + Redis.

## Demo en línea

**https://pillow-distribute-structured-attitudes.trycloudflare.com**

> La demo corre en un servidor propio detrás de un quick tunnel de Cloudflare
> (trycloudflare.com). Esa URL es temporal: cambia si se reinicia el túnel,
> y este enlace se actualiza solo cuando eso pasa.

## Documentación

Toda la documentación está en [`docs/`](docs):

| Documento | Contenido |
|---|---|
| [PROJECT_MAP.md](docs/PROJECT_MAP.md) | Arquitectura, endpoints, migraciones y ubicación de los archivos |
| [REQUIREMENTS_CHECKLIST.md](docs/REQUIREMENTS_CHECKLIST.md) | Requisitos del proyecto y avance por unidad |
| [despliegue.md](docs/despliegue.md) | Servidor con Docker, Cloudflare Tunnel y despliegue automático |
| [desarrollo-local.md](docs/desarrollo-local.md) | Cómo levantar el proyecto para programar |

## Despliegue automático

Cada push a `main` se despliega solo en el servidor mediante un runner
self-hosted de GitHub Actions (`.github/workflows/desplegar.yml`).
Detalles en [docs/despliegue.md](docs/despliegue.md#8-despliegue-automático-con-github-actions).
