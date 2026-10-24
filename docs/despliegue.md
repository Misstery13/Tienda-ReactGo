# Despliegue en un servidor con Docker y Cloudflare Tunnel

Levanta la aplicación completa en Docker y la publica en internet por un túnel
de Cloudflare, sin abrir ningún puerto del router. Después, un runner de GitHub
Actions la vuelve a desplegar sola en cada push a `main`.

```
Internet ──► Cloudflare Tunnel ──► web (nginx)
                                     ├─ /api/saas/* ──► saas (FastAPI)
                                     ├─ /api/*      ──► backend (Go/Fiber)
                                     └─ /*          ──► SPA de React (PWA)
                                               │
                                 postgres 16 + redis
```

Solo `web` queda detrás del túnel. PostgreSQL y Redis no publican puertos:
son inalcanzables desde fuera de la red de Docker.

**Estado actual del servidor:** quick tunnel (opción A del paso 3) y despliegue
automático con runner self-hosted (paso 8). La URL vigente está en el README.

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

---

## 3. Elegir el túnel

### Opción A: quick tunnel, sin dominio (la que usa el servidor)

Cloudflare da una URL aleatoria `https://<palabras>.trycloudflare.com` sin
cuenta ni token. La URL **cambia cada vez que se reinicia el contenedor
`cloudflared`**, y Cloudflare no garantiza su disponibilidad: sirve para pruebas
y demos.

Crear el archivo `docker-compose.override.yml` en la raíz del clon. Docker
Compose lo carga solo junto al `docker-compose.yml`:

```yaml
# Override local de este servidor (no versionado): quick tunnel de trycloudflare.com
# en lugar del túnel con token. Compose lo carga solo junto a docker-compose.yml.
services:
  cloudflared:
    command: tunnel --no-autoupdate --url http://web:80
```

Excluirlo de git solo en este clon, para que `git pull` no choque:

```bash
echo docker-compose.override.yml >> .git/info/exclude
```

### Opción B: túnel con dominio propio (URL fija)

Requiere un dominio administrado en Cloudflare. Se hace en el navegador:

1. Panel de Cloudflare → **Zero Trust** → **Networks** → **Tunnels**.
2. **Create a tunnel** → tipo **Cloudflared** → nombre, por ejemplo `multicatalogo`.
3. Copiar el **token** (empieza con `eyJ...`). Ignorar los comandos de
   instalación: el túnel ya corre dentro de Docker.
4. Pestaña **Public Hostname** → **Add a public hostname**:

   | Campo | Valor |
   |---|---|
   | Subdomain | `multicatalogo` (o el que se prefiera) |
   | Domain | el dominio de la cuenta |
   | Type | `HTTP` |
   | URL | `web:80` |

   > `web:80` es el nombre del servicio dentro de Docker. No es `localhost`
   > ni una IP: cloudflared corre en la misma red de contenedores.

5. **Save tunnel**.

Con esta opción no se crea el `docker-compose.override.yml`.

---

## 4. Configurar las variables

```bash
cp .env.deploy.example .env.deploy
chmod 600 .env.deploy
```

> Es `.env.deploy`, **no** `.env`. El `.env` lo usa el frontend en
> desarrollo y se sobrescribiría.

Generar los dos secretos y escribirlos directo en el archivo:

```bash
sed -i "s|^DB_PASSWORD=.*|DB_PASSWORD=$(openssl rand -hex 16)|" .env.deploy
sed -i "s|^JWT_SECRET=.*|JWT_SECRET=$(openssl rand -hex 32)|" .env.deploy
```

> Se usa `rand -hex` y no `rand -base64` a propósito: el hexadecimal no
> contiene `/` ni `+`, caracteres que complican las cadenas de conexión.

Completar los dos valores restantes con `nano .env.deploy`:

| Variable | Opción A (quick tunnel) | Opción B (dominio) |
|---|---|---|
| `PUBLIC_URL` | Cualquier valor por ahora; el paso 6 pone la URL real | `https://multicatalogo.tudominio.com` |
| `TUNNEL_TOKEN` | `no-usado-quick-tunnel` | El token del paso 3 |

> El `JWT_SECRET` lo comparten el backend en Go y la capa SaaS: por eso un
> token emitido al iniciar sesión sirve en los dos servicios.

---

## 5. Levantar

```bash
docker compose --env-file .env.deploy up -d --build
```

La primera vez tarda varios minutos: compila Go, instala las dependencias de
Python y construye el frontend. PostgreSQL ejecuta las migraciones `001` a `009`
de `backend/database/` solo la primera vez que se crea el volumen.

```bash
docker compose --env-file .env.deploy ps
```

Los seis servicios deben decir `Up`. `postgres` y `redis` además `(healthy)`.

---

## 6. Sincronizar la URL (solo opción A)

```bash
./scripts/sincronizar-url.sh
```

El script lee la URL del quick tunnel en los registros de `cloudflared`, la
escribe en `PUBLIC_URL`, recrea `backend` y `saas` y actualiza el enlace del
README. Si el enlace cambió, hace commit y push con `[skip ci]`.

---

## 7. Verificar

Desde el servidor, contra la red interna de Docker:

```bash
docker compose --env-file .env.deploy exec web wget -qO- http://127.0.0.1/api/health
docker compose --env-file .env.deploy exec web wget -qO- http://127.0.0.1/api/saas/health
```

Ambos deben responder `"status":"ok"`.

> Usar `127.0.0.1`, no `localhost`: dentro del contenedor `localhost`
> resuelve a IPv6 y nginx solo escucha en IPv4 (`Connection refused`).

Desde fuera, por el túnel:

```bash
curl https://<url-publica>/api/health
```

Abrir la URL en un navegador: debe aparecer la pantalla de inicio de sesión.

### Contraseñas semilla

Las cuentas de ejemplo (`admin@upse.edu.ec` y `cliente@upse.edu.ec`) traen la
contraseña `123456`. Si el despliegue no es solo de prueba, cambiarlas:

```bash
docker compose --env-file .env.deploy exec postgres \
  psql -U multicatalogo -d multicatalogo \
  -c "UPDATE usuarios SET password = crypt('LA-NUEVA-CLAVE', gen_salt('bf', 10)) WHERE email = 'admin@upse.edu.ec';"
```

---

## 8. Despliegue automático con GitHub Actions

Dos workflows corren en un **runner self-hosted** instalado en el propio
servidor. No usan secretos de GitHub: las credenciales viven solo en
`.env.deploy`, que nunca se sube al repositorio.

| Workflow | Cuándo | Qué hace |
|---|---|---|
| `desplegar.yml` | Cada push a `main`, o a mano desde **Actions** | `git reset --hard origin/main`, `up -d --build`, sincroniza la URL y verifica la API |
| `vigilar-url.yml` | Cada 15 minutos | Solo sincroniza la URL, por si el túnel se reinició sin push |

> `git reset --hard` borra cualquier cambio local a archivos versionados del
> clon. `.env.deploy` y `docker-compose.override.yml` no se tocan porque git
> los ignora.

### Instalar el runner

1. En GitHub: repositorio → **Settings** → **Actions** → **Runners** →
   **New self-hosted runner** → Linux x64. Copiar el token del comando
   `./config.sh` (dura 1 hora).
2. En el servidor, en una carpeta propia para este repositorio:

   ```bash
   mkdir ~/actions-runner-tienda && cd ~/actions-runner-tienda
   curl -fsSL -o runner.tgz https://github.com/actions/runner/releases/download/v2.338.1/actions-runner-linux-x64-2.338.1.tar.gz
   tar xzf runner.tgz && rm runner.tgz
   ./config.sh --url https://github.com/Misstery13/Tienda-ReactGo --token <TOKEN> --unattended
   sudo ./svc.sh install $USER
   sudo ./svc.sh start
   ```

   > Cada runner se registra a un solo repositorio: no reutilizar la carpeta
   > de otro proyecto.

3. El runner debe aparecer como **Idle** en GitHub.

Si el clon no está en `~/Tienda-ReactGo`, crear la variable de repositorio
`RUTA_DESPLIEGUE` (Settings → Secrets and variables → Actions → Variables).

### Credenciales de git en el servidor

`sincronizar-url.sh` hace push del README, y subir cambios a `.github/workflows/`
desde el servidor exige un Personal Access Token con los scopes **`repo`** y
**`workflow`**. Sin `workflow`, GitHub rechaza el push con:

```
refusing to allow a Personal Access Token to create or update workflow ... without `workflow` scope
```

> Ningún commit normal debe llevar el texto `[skip ci]` en el mensaje: GitHub
> salta el despliegue si aparece en cualquier parte. Solo lo usa el commit
> automático del README.

---

## Operación diaria

| Qué | Comando |
|---|---|
| Ver registros de un servicio | `docker compose --env-file .env.deploy logs -f backend` |
| Ver la URL del quick tunnel | `docker logs multicatalogo-tunnel 2>&1 \| grep trycloudflare` |
| Reiniciar un servicio | `docker compose --env-file .env.deploy restart backend` |
| Apagar todo | `docker compose --env-file .env.deploy down` |
| Apagar y borrar los datos | `docker compose --env-file .env.deploy down -v` |
| Actualizar a mano | `git pull && docker compose --env-file .env.deploy up -d --build` |

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
levantar solo. Basta con que Docker arranque con el sistema:

```bash
sudo systemctl enable docker
```

Con la opción A, al reiniciar cambia la URL; `vigilar-url.yml` la sincroniza en
menos de 15 minutos.

---

## Si algo falla

| Síntoma | Causa y arreglo |
|---|---|
| `Provided Tunnel token is not valid` | El `TUNNEL_TOKEN` quedó con el texto de ejemplo o está mal copiado (opción B) |
| `falta JWT_SECRET` al levantar | Falta `--env-file .env.deploy` en el comando |
| El túnel conecta pero da 502 | El hostname apunta mal; en Cloudflare debe decir `web:80`, no `localhost` |
| `backend` reinicia en bucle | Mirar `logs backend`; suele ser la conexión a PostgreSQL |
| `host not found in upstream` | Un servicio no levantó; mirar `ps` y los registros de ese servicio |
| Las tablas no existen | El volumen ya existía de antes: `down -v` y volver a levantar |
| `wget: can't connect to remote host: Connection refused` | Se usó `localhost`; usar `127.0.0.1` |
| El job falla en "Comprobar que responde" | Revisar los registros que imprime el job; la API no respondió en 50 s |
| El push no dispara el despliegue | El mensaje del commit contiene `[skip ci]` |
| `permission denied ... docker.sock` | El usuario no está en el grupo `docker` (paso 1) |
| La PWA instalada deja de cargar datos | Cambió la URL del quick tunnel; instalarla de nuevo desde la URL nueva |

Ver todos los registros juntos:

```bash
docker compose --env-file .env.deploy logs --tail 50
```

---

## Anexo: túnel con archivo de credenciales

Alternativa a la opción B para quien prefiera el método clásico de
`cloudflared` con credenciales en archivo en lugar de token.

1. `cloudflared tunnel login`
2. `cloudflared tunnel create multicatalogo`
3. Copiar el `<UUID>.json` que genera a `cloudflared/credentials.json`.
4. `cloudflared tunnel route dns multicatalogo multicatalogo.tudominio.com`
5. Crear `cloudflared/config.yml`:

   ```yaml
   tunnel: multicatalogo
   credentials-file: /etc/cloudflared/credentials.json

   ingress:
     # Todo el tráfico del dominio entra por nginx, que reparte internamente.
     - hostname: multicatalogo.tudominio.com
       service: http://web:80

     # Regla final obligatoria: lo que no coincida devuelve 404.
     - service: http_status:404
   ```

6. En `docker-compose.yml`, cambiar el servicio `cloudflared` por:

   ```yaml
   cloudflared:
     image: cloudflare/cloudflared:latest
     command: tunnel --no-autoupdate --config /etc/cloudflared/config.yml run
     volumes:
       - ./cloudflared:/etc/cloudflared:ro
   ```

> `credentials.json` es un secreto: no subirlo al repositorio.
