# Desarrollo local

El `docker-compose.yml` de la raíz es el de despliegue. Para trabajar en local
hay otro en `backend/` que solo levanta la base y la caché:

```bash
cd backend && docker compose up -d     # PostgreSQL en 5433, Redis en 6379
cd backend && PORT=3001 go run .       # API
cd saas && ./.venv/bin/uvicorn app.main:app --port 8000
npm run dev                            # frontend en 5173
```

Los nombres de contenedor y los puertos son distintos a propósito: los dos
entornos pueden convivir en la misma máquina.

## Variables

| Archivo | Para qué |
|---|---|
| `.env` (desde `.env.example`) | `VITE_API_URL` del frontend en desarrollo |
| `backend/.env` (desde `backend/.env.example`) | Conexión a la base, Redis y `JWT_SECRET` de la API |
| `saas/.env` (desde `saas/.env.example`) | Conexión a la base y `JWT_SECRET` de la capa SaaS |

El `JWT_SECRET` debe ser el mismo en `backend/.env` y `saas/.env`: un token
emitido al iniciar sesión en Go sirve también en la capa SaaS.

## Pruebas

```bash
cd backend && go test ./domain/...     # reglas de comisiones
```

## PWA en desarrollo

El service worker solo se genera en el build (`npm run build`). Para probar la
instalación y el modo sin conexión en local:

```bash
npm run build && npx vite preview
```
