# PROJECT_MAP — Tienda ReactGo (MultiCatálogo)

Mapa arquitectónico para ubicar archivos sin escanear el proyecto.
**Consulta este archivo antes de leer código.**

## Stack

| Capa | Tecnología | Puerto |
|---|---|---|
| Frontend | React 19 + TypeScript + Vite + Tailwind 4 + React Router 7 | 5173 |
| Backend | Go 1.26 + Fiber v2 | 3000 |
| Datos | Estáticos en memoria (PostgreSQL pendiente, Unidad 2) | — |

> Nota: el material del curso nombra las carpetas `Unidad1_Frontend/Proyecto_base` y
> `Unidad2_Backend/multicatalogo-backend`. En este repositorio equivalen a la raíz
> (frontend) y a `backend/` respectivamente.

## Estructura

```
Practica03/              # raíz del repo = frontend
├── src/                 # código React
├── backend/             # API RESTful en Go
├── public/              # estáticos servidos por Vite
└── .env                 # VITE_API_URL=http://localhost:3000
```

## Frontend — `src/`

| Ruta | Responsabilidad |
|---|---|
| `main.tsx` | Punto de entrada; monta `<App />` |
| `App.tsx` | Rutas + guardas `ProtectedRoute` / `AdminRoute` + `CartBoundary` |
| `index.css` | Entrada de Tailwind |
| `context/AuthContext.tsx` | Sesión: `user {email, rol}`, `token`, `login`, `logout` |
| `context/CartContext.tsx` | Carrito persistido por usuario en localStorage |
| `services/api.ts` | `loginRequest()` → `POST /api/login`. Lee `VITE_API_URL` |
| `services/productosService.ts` | `getProductos()` / `getProductoById()` — hoy mock |
| `data/productos.ts` | 8 productos mock con `descripcion`, `categoria`, `galeria` |
| `data/red.ts` | Árbol de referidos + funciones puras de comisiones |

### Componentes — `src/components/`

| Archivo | Vista |
|---|---|
| `Layout.tsx` | Sidebar + Navbar + `<Outlet />` |
| `Sidebar.tsx` | Navegación filtrada por rol |
| `Navbar.tsx` | Título por rol, insignia, carrito, logout |
| `Login.tsx` | Formulario; consume la API real |
| `Storefront.tsx` | Tienda full-screen (hero, categorías, destacados) |
| `Catalogo.tsx` | Grilla + búsqueda + filtro por categoría (`useMemo`) |
| `DetalleProducto.tsx` | Detalle con galería y lightbox (`useParams`) |
| `Carrito.tsx` | Ítems, cantidades, total |
| `Checkout.tsx` | Formulario de envío y pago simulado |
| `Confirmacion.tsx` | Número de pedido y resumen |
| `Dashboard.tsx` | KPIs derivados de `data/red.ts` (solo admin) |
| `MiRed.tsx` | Árbol multinivel recursivo (solo admin) |

### Rutas React

`/login` · `/` (Dashboard, admin) · `/mi-red` (admin) · `/tienda` · `/catalogo`
· `/producto/:id` · `/carrito` · `/checkout` · `/confirmacion`

## Backend — `backend/`

| Archivo | Responsabilidad |
|---|---|
| `main.go` | Arranque de Fiber, CORS y `:3000` |
| `routes/routes.go` | Registro de endpoints |
| `controllers/authController.go` | `Login`: valida entrada y responde token + rol |
| `controllers/prodController.go` | `GetProductos`, `GetProductoPorID` |
| `controllers/healthController.go` | `HealthCheck` |
| `models/models.go` | `LoginRequest`, `LoginResponse`, `Usuario`, `Producto`, `APIError` |
| `repository/repository.go` | **Única fuente de datos.** Aquí entra PostgreSQL en la Fase 2 |

### Endpoints

| Método | Ruta | Respuesta |
|---|---|---|
| GET | `/api/health` | `{status, message}` |
| POST | `/api/login` | `{token, email, rol}` · 400 / 401 con `APIError` |
| GET | `/api/productos` | Array de productos |
| GET | `/api/productos/:id` | Producto · 400 si el id no es número · 404 si no existe |

Formato de error estándar: `{ "status": int, "message": string, "details": any? }`

### Cuentas de prueba

`admin@upse.edu.ec` / `123456` — rol admin
`cliente@upse.edu.ec` / `123456` — rol cliente

## Arranque

```
backend:   cd backend && go run .      # :3000
frontend:  npm run dev                 # :5173
```

## Reglas que no se rompen

1. No cambiar las firmas de endpoints ni la forma del JSON que el frontend ya consume.
2. Los datos viven en `repository/`, nunca dentro de un controlador.
3. Los errores HTTP siempre usan `models.APIError`.
