# Despliegue en máquina propia con Cloudflare Tunnel

Levanta los cinco servicios en tu máquina y los expone por un dominio
de Cloudflare, sin abrir ni un puerto del router.

```
Internet ──► Cloudflare Tunnel ──► web (nginx)
                                     ├─ /api/saas/* ──► saas (FastAPI)
                                     ├─ /api/*      ──► backend (Go/Fiber)
                                     └─ /*          ──► SPA de React
                                               │
                                 postgres 16 + redis
```

Solo `web` queda detrás del túnel. PostgreSQL y Redis no publican puertos:
son invisibles desde fuera de la red de Docker.

## 1. Crear el túnel en Cloudflare

1. Entra al panel de Cloudflare → **Zero Trust** → **Networks** → **Tunnels**.
2. **Create a tunnel** → tipo **Cloudflared** → ponle un nombre.
3. Cloudflare muestra un **token** largo. Cópialo.
4. En **Public hostname**, agrega:
   - **Subdomain**: el que quieras (ej. `multicatalogo`)
   - **Domain**: tu dominio
   - **Service**: `HTTP` → `web:80`

## 2. Configurar las variables

```bash
cp .env.deploy.example .env
```

Edita `.env` y rellena:

| Variable | Qué poner |
|---|---|
| `DB_PASSWORD` | Clave larga. Generar: `openssl rand -base64 24` |
| `JWT_SECRET` | Secreto largo. Generar: `openssl rand -base64 48` |
| `PUBLIC_URL` | `https://multicatalogo.tudominio.com` |
| `TUNNEL_TOKEN` | El token del paso 1 |

> El `JWT_SECRET` lo comparten el backend en Go y la capa SaaS: por eso un
> token emitido al iniciar sesión sirve en los dos servicios.

## 3. Levantar

```bash
docker compose up -d --build
```

La primera vez tarda: compila Go, instala las dependencias de Python y
construye el frontend. PostgreSQL ejecuta solo las migraciones `001` a `009`
de `backend/database/`.

Comprobar que todo arrancó:

```bash
docker compose ps
```

## 4. Verificar

```bash
curl https://multicatalogo.tudominio.com/api/health
curl https://multicatalogo.tudominio.com/api/saas/health
```

Ambos deben responder `"status":"ok"`. Luego abre el dominio en el navegador
e inicia sesión con `admin@upse.edu.ec` / `123456`.

## 5. Cambiar las contraseñas semilla

Las cuentas de ejemplo tienen la contraseña `123456`. En un despliegue
público hay que cambiarlas:

```bash
docker compose exec postgres psql -U multicatalogo -d multicatalogo \
  -c "UPDATE usuarios SET password = crypt('TU-NUEVA-CLAVE', gen_salt('bf', 10)) WHERE email = 'admin@upse.edu.ec';"
```

## Operación diaria

| Qué | Comando |
|---|---|
| Ver registros | `docker compose logs -f backend` |
| Reiniciar un servicio | `docker compose restart backend` |
| Apagar todo | `docker compose down` |
| Apagar y borrar datos | `docker compose down -v` |
| Actualizar tras un cambio | `git pull && docker compose up -d --build` |

### Respaldo de la base

```bash
docker compose exec postgres pg_dump -U multicatalogo multicatalogo > respaldo.sql
```

### Restaurar

```bash
cat respaldo.sql | docker compose exec -T postgres psql -U multicatalogo -d multicatalogo
```

## Desarrollo local

Este `docker-compose.yml` es el de despliegue. Para trabajar en local hay
otro en `backend/` que solo levanta la base y la caché:

```bash
cd backend && docker compose up -d     # PostgreSQL en 5433, Redis en 6379
cd backend && PORT=3001 go run .       # API
cd saas && ./.venv/bin/uvicorn app.main:app --port 8000
npm run dev                            # frontend en 5173
```

Los nombres de contenedor y los puertos son distintos a propósito: los dos
entornos pueden convivir en la misma máquina.

## Si algo falla

| Síntoma | Causa probable |
|---|---|
| `falta JWT_SECRET` al levantar | No copiaste `.env.deploy.example` a `.env` |
| El túnel no conecta | Token mal copiado, o el hostname no apunta a `web:80` |
| 502 en `/api/` | El backend no arrancó; mira `docker compose logs backend` |
| La web carga pero el login falla | Revisa que `PUBLIC_URL` tenga `https://` y sea el dominio real |
| Las tablas no existen | El volumen ya existía; `docker compose down -v` y vuelve a levantar |
