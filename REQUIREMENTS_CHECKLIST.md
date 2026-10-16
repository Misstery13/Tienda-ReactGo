# REQUIREMENTS CHECKLIST — MultiCatálogo

> Fuente de verdad del progreso. Leer al inicio de cada sesión; marcar `- [x]`
> y actualizar "Estado General" al completar cualquier tarea.
>
> **Estructura:** el documento se divide en fases por **Unidad/Tema del sílabo**.
> El producto integrador es UNO: MultiCatálogo — una tienda cosmética con red de
> ventas multinivel (MLM).
>
> **Principios rectores (aplican a toda fase):**
> - Seguridad por defecto: ningún endpoint nuevo nace sin auth ni validación.
> - Inyección de dependencias: interfaces en los bordes; globals solo en el entry point.
> - Funciones puras para el dominio (comisiones, totales, validadores) — testeables sin I/O.
> - Mobile-first: UI a 375px, touch ≥44px.

## Estado General

| Ítem | Valor |
|---|---|
| **Hito actual** | **Unidades 1 y 2 completadas** (SPA con sesión JWT persistida + backend seguro: bcrypt, JWT real, rutas protegidas, health con deps) |
| **Última tarea completada** | Seguridad U2-T4 (bcrypt + JWT + RequireAuth + identidad por token) + mejoras U1 (apiFetch, persistencia, audit mobile) — 2026-10-16 |
| **Siguiente tarea pendiente** | U3-T1: scaffold FastAPI `Unidad3_SaaS/` |
| **Bloqueadores** | Los contenedores Docker de otro proyecto ocupan **:3000** y **:5173** → este proyecto corre en **:3001** (API) y **:5174** (Vite). PostgreSQL del proyecto en **:5433** (el PostgreSQL 18 del instalador EDB ocupa el 5432) |

## Entorno local verificado

```
React SPA (:5174) ──► Go/Fiber (:3001) ──► PostgreSQL 16 (:5433) + Redis (:6379)
```

| Servicio | Comando |
|---|---|
| PostgreSQL | `/opt/homebrew/opt/postgresql@16/bin/pg_ctl -D /opt/homebrew/var/postgresql@16 -o "-p 5433" start` |
| Redis | `brew services start redis` |
| Backend | `cd backend && PORT=3001 go run .` |
| Frontend | `npm run dev` |

---

# UNIDAD 1 — FRAMEWORKS FRONTEND (SPA React + Vite + Tailwind)

## U1-T2 · Entorno SPA con Vite + Tailwind CSS

- [x] Entorno SPA con Vite + Tailwind v4, estructura por responsabilidad
- [x] `package.json` renombrado a `multicatalogo-frontend`
- [x] `VITE_API_URL` en `.env` + cliente único en `services/http.ts`

## U1-T3 · Componentes y gestión de estado global

- [x] `AuthContext` (sesión `{email, rol}`) y `CartContext` (carrito por email en localStorage)
- [x] Persistencia de sesión (`multicatalogo_sesion`: usuario + JWT) — sobrevive a F5
- [x] Split `context/auth.ts` + `context/cart.ts` → `npm run lint` con 0 errores y 0 warnings

## U1-T4 · Consumo de API RESTful y enrutamiento

- [x] `productosService.ts`, `redService.ts`, `checkoutService.ts` + guardas `ProtectedRoute`/`AdminRoute`
- [x] `services/http.ts`: `apiFetch` con `Authorization: Bearer` automático, `ApiError` tipado y helpers de sesión
- [x] `apiFetch` borra la sesión ante un 401 (token expirado)

## U1-T5 · Interfaces full-screen y flujos MLM (Práctica 03)

- [x] F8 Login con rol + redirección por rol
- [x] F1 Storefront `/tienda` | F3 Catálogo con búsqueda/filtros | F2 Detalle `/producto/:id` | F4 Lightbox
- [x] F5 Carrito + checkout + confirmación | F6 Árbol multinivel 10/5/2 % | F7 Dashboard KPIs
- [x] `Checkout.tsx` → `POST /api/checkout` real (sin simulación) + manejo de `ApiError`
- [x] Audit mobile-first: 375px sin overflow en Dashboard/MiRed/Catálogo/Carrito
- [x] Botones de acción ≥44px (`Añadir al Carrito` 36→44, steppers del carrito, `Cerrar Sesión`)
- [x] Corregidos 2 caracteres de control (`\x02`) heredados del material: uno rompía
      `group-hover:opacity-100`, dejando el menú de "Cerrar Sesión" invisible al pasar el mouse

---

# UNIDAD 2 — BACKEND ALTA TRANSACCIONALIDAD (Go/Fiber)

## U2-T1 · Rendimiento, memoria y concurrencia

- [x] Goroutine `procesarComisiones` tras el COMMIT, con contexto propio (no el del request)

## U2-T2 · Microservicios y APIs robustas (Fiber)

- [x] Capas routes → controllers → repository; endpoints login/productos/red/checkout/health
- [x] `APIError{status,message,details}` en todos los errores + validación de entrada (400/404)
- [x] Puerto configurable (`PORT`, luego `APP_PORT`, defecto 3000)
- [x] `/api/health` real: ping a PostgreSQL (503 si cae) y Redis (`"degraded"` si cae)
- [x] CORS configurable por entorno (`CORS_ORIGINS`) con `Authorization` permitido

## U2-T3 · Persistencia PostgreSQL + caché Redis

- [x] `docker-compose` PostgreSQL 16 + Redis alpine + volumen + healthchecks
- [x] `001_init.sql` (usuarios/productos/referidos + seed), repositorio con SQL real, árbol por `parent_id`
- [x] Cache-Aside: `productos:all` TTL 5 min, `red:tree:{email}` TTL 10 min, degradación sin caché
- [x] Índices: `referidos(parent_id)`, `pedidos(cliente_email)`, `pedido_items(pedido_id)`, `productos(categoria)`
- [x] Errores propagados: el repositorio devuelve `error`; `ErrNoEncontrado`→404/401, resto→500

## U2-T4 · Inyección de dependencias y microservicio seguro

- [x] Interfaces `DBRepository`/`CacheRepository`; `PostgresRepository`/`RedisCache`/`SinCache`
- [x] `CachedRepository` como decorador: el Cache-Aside no vive dentro del SQL
- [x] Controladores como structs; construidos e inyectados en `main.go`
- [x] `POST /api/checkout` ACID: `FOR UPDATE`, descuento de stock, `pedidos`/`pedido_items`, ROLLBACK → 400/500
- [x] Total recalculado en el servidor: los precios salen de PostgreSQL, nunca del cliente
- [x] Invalidación de caché tras la compra (`red:tree:*`) + migración `002_checkout.sql`
- [x] bcrypt de contraseñas (pgcrypto `crypt()/gen_salt('bf')` en `003_seguridad.sql`; login con `CompareHashAndPassword`)
- [x] JWT real (`JWT_SECRET`, claims `{sub,email,rol,iat,exp}` 2 h, HS256) + `RequireAuth`/`RequireAdmin`
- [x] Protegidos `/api/red` (admin) y `/api/checkout` (autenticado): 401 sin token, 401 adulterado, 403 sin rol
- [x] Checkout toma la identidad del TOKEN — verificado: un `email` falso en el cuerpo no se persiste
      ni atribuye comisiones (pedido MC-000006 quedó a nombre del cliente real)
- [x] Errores sanitizados: `Details` solo con `APP_ENV=dev`
- [x] `ErrCantidadInvalida` propio (400), distinto de `ErrStockInsuficiente`
- [x] `SIMULAR_FALLA_CHECKOUT=true` para demostrar el ROLLBACK sin editar código

---

# PENDIENTE

## Seguridad que falta

- [ ] Límite de intentos en `/api/login` contra fuerza bruta
- [ ] Refresh token (hoy la sesión expira a las 2 h y obliga a iniciar sesión otra vez)
- [ ] HTTPS en despliegue

## Funcionalidad del sistema

- [ ] CRUD de productos (el panel de administración solo lee)
- [ ] Historial de pedidos para el cliente (`GET /api/pedidos`)
- [ ] Alta de referidos y enlace de invitación
- [ ] Registro de usuarios y recuperación de contraseña
- [ ] Estados del pedido (pendiente, pagado, enviado, entregado)
- [ ] Mostrar y respetar el stock en el catálogo
- [ ] Búsqueda y paginación en el servidor
- [ ] Pruebas automatizadas (la inyección de dependencias ya las habilita)

## Unidades siguientes

- [ ] U3 — Capa SaaS/multitenant en FastAPI
- [ ] U4 — Despliegue contenerizado con proxy inverso y SSL
