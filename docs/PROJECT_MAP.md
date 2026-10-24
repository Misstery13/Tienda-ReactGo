# PROJECT_MAP — Tienda ReactGo (MultiCatálogo)

Mapa arquitectónico para ubicar archivos sin escanear el proyecto.
**Consulta este archivo antes de leer código.**

## Stack

| Capa | Tecnología | Puerto |
|---|---|---|
| Frontend | React 19 + TypeScript + Vite 8 + Tailwind 4 + React Router 7, PWA | 5173 |
| Núcleo transaccional | Go + Fiber v2 (U2) | 3000 |
| Capa SaaS/multitenant | Python 3 + FastAPI + asyncpg (U3) | 8000 |
| Datos | PostgreSQL 16 + Redis (caché) | 5433 / 6379 |
| Despliegue | Docker Compose + nginx + Cloudflare Tunnel | 80 (interno) |

> Si otros proyectos ocupan esos puertos, use `PORT` (Go) y `PORT` (SaaS).
> En esta máquina los contenedores de otro proyecto toman 3000 y 5173, por lo
> que se usa 3001 (API) y 5174 (Vite).

> Nota: el material del curso nombra las carpetas `Unidad1_Frontend/Proyecto_base` y
> `Unidad2_Backend/multicatalogo-backend`. En este repositorio equivalen a la raíz
> (frontend) y a `backend/` respectivamente.

## Estructura

```
Tienda-ReactGo/          # raíz del repo = frontend
├── src/                 # código React
├── public/              # íconos de la PWA (generados desde pwa-icon.svg)
├── backend/             # núcleo transaccional en Go (U2)
│   └── database/        # TODAS las migraciones SQL (001 a 009)
├── saas/                # capa SaaS/multitenant en FastAPI (U3)
├── docs/                # documentación
├── scripts/             # scripts del servidor (sincronizar-url.sh)
├── .github/workflows/   # despliegue automático (runner self-hosted)
├── docker-compose.yml   # despliegue completo (6 servicios)
├── Dockerfile           # frontend: build de Vite + nginx
└── nginx.conf           # proxy inverso: /api/saas, /api y la SPA
```

## Frontend — `src/`

| Ruta | Responsabilidad |
|---|---|
| `main.tsx` | Punto de entrada; monta `<App />` |
| `App.tsx` | Rutas + guardas `ProtectedRoute` / `AdminRoute` + `CartBoundary` |
| `index.css` | Entrada de Tailwind |
| `context/AuthContext.tsx`, `auth.ts` | Sesión: `user {email, rol}`, `token`, `login`, `logout` |
| `context/CartContext.tsx`, `cart.ts` | Carrito persistido por usuario en localStorage |
| `services/http.ts` | `apiFetch()` con el JWT; lee y guarda la sesión. Lee `VITE_API_URL` |
| `services/api.ts` | `loginRequest()` → `POST /api/login` |
| `services/productosService.ts` | Catálogo público y CRUD de administración |
| `services/checkoutService.ts` | `crearPedido()` → `POST /api/checkout` |
| `services/pedidosService.ts` | Pedidos propios, todos (admin) y cambio de estado |
| `services/redService.ts` | Árbol de referidos y comisiones propias |
| `services/reportesService.ts` | KPIs, rankings, deuda y liquidaciones |
| `services/usuariosService.ts` | Registro y administración de cuentas |
| `types/producto.ts` | Tipo `Producto` |
| `domain/red.ts` | Tipo `Referido` + funciones puras de comisiones |
| `hooks/useInstalarPWA.ts` | Botón "Instalar app" (`beforeinstallprompt`) |

### Componentes — `src/components/`

| Archivo | Vista |
|---|---|
| `Layout.tsx` | Sidebar + Navbar + `<Outlet />` |
| `Sidebar.tsx` | Navegación filtrada por rol, autora y botón de instalar la PWA |
| `Navbar.tsx` | Título por rol, insignia, carrito, logout |
| `Login.tsx` / `Registro.tsx` | Inicio de sesión y registro con código de referido |
| `Storefront.tsx` | Tienda full-screen (hero, categorías, destacados) |
| `Catalogo.tsx` | Grilla + búsqueda + filtro por categoría (`useMemo`) |
| `DetalleProducto.tsx` | Detalle con galería y lightbox (`useParams`) |
| `Carrito.tsx` | Ítems, cantidades, total |
| `Checkout.tsx` | Formulario de envío y pago simulado |
| `Confirmacion.tsx` | Número de pedido y resumen |
| `MisPedidos.tsx` | Historial de pedidos del usuario |
| `MiRed.tsx` | Árbol multinivel recursivo, código de invitación y comisiones |
| `Dashboard.tsx` | KPIs y rankings reales (solo admin) |
| `AdminProductos.tsx` | Alta, edición y desactivación de productos (solo admin) |
| `AdminPedidos.tsx` | Todos los pedidos y su estado (solo admin) |
| `AdminLiquidaciones.tsx` | Deuda por distribuidor y pagos (solo admin) |

### Rutas React

Públicas: `/login` · `/registro`

Ambos roles: `/tienda` · `/catalogo` · `/producto/:id` · `/carrito` · `/checkout`
· `/confirmacion` · `/mis-pedidos` · `/mi-red`

Solo admin: `/` (Dashboard) · `/admin/productos` · `/admin/pedidos` · `/admin/liquidaciones`

### PWA

`vite-plugin-pwa` (configurado en `vite.config.ts`) genera `manifest.webmanifest`
y un service worker de Workbox. La interfaz y las fotos de productos se guardan
para abrir sin conexión; **la API no se cachea**. Los íconos se regeneran con
`npx pwa-assets-generator` (`pwa-assets.config.ts`).

## Backend — `backend/`

| Ruta | Responsabilidad |
|---|---|
| `main.go` | Arranque de Fiber, CORS, inyección de dependencias y puerto (`PORT`) |
| `routes/routes.go` | Registro de endpoints |
| `controllers/` | Solo HTTP: validan la entrada y responden (`auth`, `prod`, `checkout`, `pedido`, `red`, `usuario`, `admin`, `health`) |
| `middleware/auth.go` | `RequireAuth` (JWT) y `RequireAdmin` |
| `auth/jwt.go` | Firma y verificación de tokens |
| `config/` | Conexión a PostgreSQL (`db.go`) y Redis (`redis.go`) |
| `domain/` | Reglas puras: comisiones (con pruebas) y estados del pedido |
| `models/models.go` | Estructuras de entrada y salida, `APIError` |
| `repository/` | **Única fuente de datos.** `postgres.go` + un archivo por agregado; `cached.go` envuelve las lecturas con Redis |

### Endpoints

| Método | Ruta | Acceso |
|---|---|---|
| GET | `/api/health` | Público. Estado con ping a PostgreSQL y Redis |
| POST | `/api/login` | Público. `{token, email, rol}` |
| POST | `/api/register` | Público. Acepta `codigo_referido` o `?ref=` |
| GET | `/api/productos`, `/api/productos/:id` | Público |
| POST | `/api/checkout` | Sesión |
| GET | `/api/pedidos/mios`, `/api/pedidos/:id` | Sesión |
| GET | `/api/red/mia`, `/api/comisiones/mias`, `/api/liquidaciones/mias` | Sesión |
| GET | `/api/red` | Admin |
| GET / POST / PUT / DELETE | `/api/admin/productos`, `/api/productos[/:id]` | Admin. DELETE desactiva (borrado lógico) |
| GET / PATCH | `/api/pedidos`, `/api/pedidos/:id/estado` | Admin |
| GET / POST | `/api/liquidaciones`, `/api/comisiones/deuda` | Admin |
| GET | `/api/reportes/resumen`, `/top-productos`, `/top-referidos` | Admin |
| GET / PATCH | `/api/usuarios`, `/api/usuarios/:id` | Admin |

Formato de error estándar: `{ "status": int, "message": string, "details": any? }`

Estados del pedido: `pendiente` → `preparando` → `enviado` → `entregado`, o `cancelado`.

### Migraciones — `backend/database/`

PostgreSQL las ejecuta en orden alfabético solo al crear el volumen.

| Archivo | Contenido |
|---|---|
| `001_init.sql` | Esquema base y datos semilla |
| `002_checkout.sql` | Pedidos, stock y comisiones |
| `003_seguridad.sql` | Contraseñas con bcrypt e índices |
| `004_saas.sql` | Planes, tenants y suscripciones (multitenancy) |
| `005_registro.sql` | Registro de clientes y vínculo usuario ↔ referido |
| `006_productos.sql` | Administración del catálogo con borrado lógico |
| `007_pedidos.sql` | Ciclo de vida del pedido |
| `008_comisiones.sql` | Ledger de comisiones |
| `009_liquidaciones.sql` | Liquidaciones a distribuidores |

### Cuentas de prueba

`admin@upse.edu.ec` / `123456` — rol admin
`cliente@upse.edu.ec` / `123456` — rol cliente

## Capa SaaS — `saas/` (Unidad 3)

Python + FastAPI + asyncpg. Se monta bajo `/api/saas` y comparte la base de
datos y el secreto JWT con el núcleo en Go: un token emitido por `/api/login`
de Fiber es válido aquí, sin duplicar la autenticación.

| Archivo | Responsabilidad |
|---|---|
| `app/main.py` | App FastAPI, CORS, ciclo de vida del pool, prefijo `/api/saas` |
| `app/config.py` | Settings desde `.env` (pydantic-settings) |
| `app/db.py` | Pool asyncpg; dependencia `get_pool` |
| `app/security.py` | Verifica el JWT de Go; dependencias `usuario_actual` y `requiere_admin` |
| `app/tenancy.py` | Resuelve el tenant por `X-Tenant` o por el usuario del token |
| `app/schemas.py` | Modelos Pydantic de entrada y salida |
| `app/routers/health.py` | `GET /api/saas/health` con ping real a PostgreSQL |
| `app/routers/planes.py` | `GET /api/saas/planes` (público) |
| `app/routers/tenants.py` | `GET`/`POST /api/saas/tenants` (solo admin) |
| `app/routers/suscripciones.py` | `GET /api/saas/mi-suscripcion` con uso vs. límites |

Modelo de multitenancy: **discriminador** (`tenant_id` en cada tabla).
Documentación interactiva: <http://localhost:8000/api/saas/docs>

## Arranque y despliegue

- Desarrollo local: [desarrollo-local.md](desarrollo-local.md)
- Servidor con Docker, túnel y despliegue automático: [despliegue.md](despliegue.md)

## Reglas que no se rompen

1. No cambiar las firmas de endpoints ni la forma del JSON que el frontend ya consume.
2. Los datos viven en `repository/`, nunca dentro de un controlador.
3. Los errores HTTP siempre usan `models.APIError`.
