# Tienda ReactGo (MultiCatálogo)

Tienda multicatálogo con frontend en React, núcleo transaccional en Go/Fiber,
capa SaaS en FastAPI y datos en PostgreSQL 16 + Redis.

## Demo en línea

**https://pillow-distribute-structured-attitudes.trycloudflare.com**

> La demo corre en un servidor propio detrás de un quick tunnel de Cloudflare
> (trycloudflare.com). Esa URL es temporal: cambia si se reinicia el túnel.
> La URL vigente aparece en el resumen del último despliegue, en la pestaña
> **Actions** del repositorio.

## Documentación

- [PROJECT_MAP.md](PROJECT_MAP.md): arquitectura y ubicación de los archivos.
- [DESPLIEGUE_SERVIDOR.md](DESPLIEGUE_SERVIDOR.md): despliegue en un servidor Ubuntu con Cloudflare Tunnel.
- [DESPLIEGUE.md](DESPLIEGUE.md): despliegue en máquina propia.
- [REQUIREMENTS_CHECKLIST.md](REQUIREMENTS_CHECKLIST.md): requisitos del proyecto.

## Despliegue automático

Cada push a `main` se despliega solo en el servidor mediante un runner
self-hosted de GitHub Actions (`.github/workflows/desplegar.yml`).
