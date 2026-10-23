# Desplegar MultiCatálogo en un servidor Ubuntu con Cloudflare Tunnel

Instrucciones para ejecutar en el servidor. Levanta la aplicación completa
en Docker y la publica en internet por un túnel de Cloudflare, sin abrir
ningún puerto del router.

**No se necesita un runner de GitHub.** Un runner sirve para desplegar
automáticamente en cada push; aquí se clona el repositorio y se levanta a
mano. Al final hay una sección opcional por si más adelante se quiere
automatizar.

```
Internet ──► Cloudflare Tunnel ──► web (nginx)
                                     ├─ /api/saas/* ──► saas (FastAPI)
                                     ├─ /api/*      ──► backend (Go/Fiber)
                                     └─ /*          ──► SPA de React
                                               │
                                 postgres 16 + redis
```

Solo `web` queda detrás del túnel. PostgreSQL y Redis no publican puertos:
son inalcanzables desde fuera de la red de Docker.

---

## 1. Comprobar requisitos

```bash
docker --version
docker compose version
git --version
```

Si falta Docker:

```bash
curl -fsSL https://get.docker.com | sudo sh
sudo usermod -aG docker $USER
newgrp docker
```

Comprobar que el usuario puede usar Docker sin `sudo`:

```bash
docker run --rm hello-world
```

---

## 2. Clonar el repositorio

```bash
git clone https://github.com/Misstery13/Tienda-ReactGo.git
cd Tienda-ReactGo
```

Si ya estaba clonado:

```bash
cd Tienda-ReactGo && git pull
```

---

## 3. Crear el túnel en Cloudflare

Este paso se hace en el navegador, no en el servidor.

1. Panel de Cloudflare → **Zero Trust** → **Networks** → **Tunnels**.
2. **Create a tunnel** → tipo **Cloudflared** → nombre, por ejemplo `multicatalogo`.
3. Cloudflare muestra un **token** largo (empieza con `eyJ...`). Copiarlo.
   - Ignorar los comandos de instalación que sugiere: aquí el túnel corre
     dentro de Docker, ya está en el `docker-compose.yml`.
4. Pestaña **Public Hostname** → **Add a public hostname**:

   | Campo | Valor |
   |---|---|
   | Subdomain | `multicatalogo` (o el que se prefiera) |
   | Domain | el dominio de la cuenta |
   | Type | `HTTP` |
   | URL | `web:80` |

   > El valor `web:80` es el nombre del servicio dentro de Docker. No es
   > `localhost` ni una IP: cloudflared corre en la misma red de contenedores.

5. **Save tunnel**.

---

## 4. Configurar las variables

```bash
cp .env.deploy.example .env.deploy
```

Generar los dos secretos y escribirlos directo en el archivo:

```bash
sed -i "s|^DB_PASSWORD=.*|DB_PASSWORD=$(openssl rand -hex 16)|" .env.deploy
sed -i "s|^JWT_SECRET=.*|JWT_SECRET=$(openssl rand -hex 32)|" .env.deploy
```

> Se usa `rand -hex` y no `rand -base64` a propósito: el hexadecimal no
> contiene `/` ni `+`, caracteres que complican las cadenas de conexión.

Ahora editar el archivo y poner los dos valores que faltan:

```bash
nano .env.deploy
```

| Variable | Qué poner |
|---|---|
| `PUBLIC_URL` | `https://multicatalogo.tudominio.com` (el hostname del paso 3, con `https://`) |
| `TUNNEL_TOKEN` | El token copiado en el paso 3 |

Comprobar que no quedó ningún valor de ejemplo:

```bash
grep -E "pon-aqui|tudominio" .env.deploy && echo "FALTAN VALORES POR LLENAR" || echo "configuración completa"
```

---

## 5. Levantar

```bash
docker compose --env-file .env.deploy up -d --build
```

La primera vez tarda varios minutos: compila el binario de Go, instala las
dependencias de Python y construye el frontend de React.

PostgreSQL ejecuta solo las migraciones `001` a `009` de `backend/database/`
la primera vez que se crea el volumen.

Ver el estado:

```bash
docker compose --env-file .env.deploy ps
```

Los seis servicios deben decir `Up`. `postgres` y `redis` además `(healthy)`.

---

## 6. Verificar

Desde el propio servidor, contra la red interna de Docker:

```bash
docker compose --env-file .env.deploy exec web wget -qO- http://localhost/api/health
docker compose --env-file .env.deploy exec web wget -qO- http://localhost/api/saas/health
```

Ambos deben responder `"status":"ok"`.

Desde fuera, por el túnel:

```bash
curl https://multicatalogo.tudominio.com/api/health
```

Y abrir el dominio en un navegador. Debe aparecer la pantalla de inicio de
sesión.

---

## 7. Cambiar las contraseñas semilla

Las cuentas de ejemplo (`admin@upse.edu.ec` y `cliente@upse.edu.ec`) traen la
contraseña `123456`. En un despliegue público hay que cambiarlas antes de
compartir el enlace:

```bash
docker compose --env-file .env.deploy exec postgres \
  psql -U multicatalogo -d multicatalogo \
  -c "UPDATE usuarios SET password = crypt('LA-NUEVA-CLAVE', gen_salt('bf', 10)) WHERE email = 'admin@upse.edu.ec';"
```

Repetir para `cliente@upse.edu.ec` si se va a usar esa cuenta.

---

## Operación diaria

| Qué | Comando |
|---|---|
| Ver registros de un servicio | `docker compose --env-file .env.deploy logs -f backend` |
| Reiniciar un servicio | `docker compose --env-file .env.deploy restart backend` |
| Apagar todo | `docker compose --env-file .env.deploy down` |
| Apagar y borrar los datos | `docker compose --env-file .env.deploy down -v` |
| Actualizar tras un cambio en git | `git pull && docker compose --env-file .env.deploy up -d --build` |

### Respaldo de la base

```bash
docker compose --env-file .env.deploy exec postgres \
  pg_dump -U multicatalogo multicatalogo > respaldo-$(date +%F).sql
```

### Restaurar

```bash
cat respaldo-2026-10-23.sql | docker compose --env-file .env.deploy exec -T postgres \
  psql -U multicatalogo -d multicatalogo
```

### Arranque automático al reiniciar el servidor

Los servicios llevan `restart: unless-stopped`, así que Docker los vuelve a
levantar solo. Basta con que el servicio de Docker arranque con el sistema:

```bash
sudo systemctl enable docker
```

---

## Si algo falla

| Síntoma | Causa y arreglo |
|---|---|
| `Provided Tunnel token is not valid` | El `TUNNEL_TOKEN` quedó con el texto de ejemplo o está mal copiado |
| `falta JWT_SECRET` al levantar | Falta `--env-file .env.deploy` en el comando |
| El túnel conecta pero da 502 | El hostname apunta mal; en Cloudflare debe decir `web:80`, no `localhost` |
| `backend` reinicia en bucle | Mirar `logs backend`; suele ser la conexión a PostgreSQL |
| `host not found in upstream` | Un servicio no levantó; mirar `ps` y los registros de ese servicio |
| Las tablas no existen | El volumen ya existía de antes: `down -v` y volver a levantar |
| Arquitectura distinta (ARM/x86) | Las imágenes se construyen en el servidor, así que no hay problema |

Ver todos los registros juntos:

```bash
docker compose --env-file .env.deploy logs --tail 50
```

---

## Opcional: desplegar automático con GitHub Actions

Solo si se quiere que cada `git push` actualice el servidor. **No es necesario**
para que la aplicación funcione.

Hacen falta dos cosas:

1. Un **runner self-hosted** registrado en el servidor:
   repositorio → **Settings** → **Actions** → **Runners** → **New self-hosted runner**,
   y seguir los comandos que da GitHub.

2. Un flujo de trabajo en `.github/workflows/desplegar.yml`:

```yaml
name: Desplegar

on:
  push:
    branches: [main]

jobs:
  desplegar:
    runs-on: self-hosted
    steps:
      - uses: actions/checkout@v4
      - name: Levantar la pila
        run: docker compose --env-file .env.deploy up -d --build
```

> El archivo `.env.deploy` vive solo en el servidor y está en `.gitignore`:
> los secretos nunca viajan al repositorio.
