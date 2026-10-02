# REQUIREMENTS_CHECKLIST — Tienda ReactGo (MultiCatálogo)

Fuente de verdad del progreso. **Léelo antes de explorar código.**
Al completar una tarea verificada, marca su casilla y actualiza el Estado General.

> Fuentes: guion de la Práctica 03 (Unidad 1, Tema 5) y documento de Fase 1
> (Unidad 2, Arquitectura y API robusta). El sílabo no está en el repositorio,
> así que los requisitos de evaluación se transcriben desde la rúbrica del guion.

## Estado General

- **Hito actual:** Unidad 2 — Fase 1 (Arquitectura y API robusta)
- **Última tarea completada:** Fase 1 pasos 1–4 (APIError, validaciones, `GetProductoPorID`, `/api/health`) verificados con 11 pruebas
- **Siguiente tarea pendiente:** Fase 2 — persistencia en PostgreSQL dentro de `backend/repository/`
- **Última actualización:** 2026-10-02

---

## Hito 1 — Práctica 03 (Unidad 1, Tema 5) · COMPLETADO

### Entregables

- [x] F1 — Storefront a pantalla completa con hero y secciones (`/tienda`)
- [x] F2 — Ruta dinámica de detalle (`/producto/:id`)
- [x] F3 — Búsqueda y filtros del catálogo con `useMemo`
- [x] F4 — Galería con lightbox accesible en el detalle
- [x] F5 — Carrito persistente por usuario + checkout + confirmación
- [x] F6 — Árbol multinivel de referidos con comisiones por nivel
- [x] F7 — Dashboard con KPIs derivados del estado
- [x] F8 — Login con rol y navegación filtrada por perfil

### Frontend

- [x] `AuthContext` guarda usuario (`email`, `rol`) y token
- [x] `CartContext` persiste en localStorage con clave por usuario
- [x] `CartProvider` se remonta con `key={user.email}` al cambiar de cuenta
- [x] Guardas `ProtectedRoute` (autenticado) y `AdminRoute` (rol admin)
- [x] Capa de servicios (`services/`) separada de los componentes
- [x] Datos mock aislados en `data/productos.ts` y `data/red.ts`
- [x] Cierre del lightbox con tecla Escape y `aria-label` en controles
- [x] Diseño responsive con sidebar colapsable

### Backend

- [x] `POST /api/login` devuelve el rol (`admin` / `cliente`)
- [x] CORS habilitado para `http://localhost:5173`

### Verificación manual

- [x] Admin entra al Dashboard; cliente entra a la Tienda
- [x] Cliente pidiendo `/mi-red` o `/` es redirigido a `/tienda`
- [x] Sidebar muestra 4 opciones a admin y 2 a cliente
- [x] Búsqueda y filtro funcionan; la URL refleja `?categoria=`
- [x] Lightbox abre, navega con ‹ › y cierra con Escape
- [x] El carrito sobrevive a la recarga de la página
- [x] Checkout genera número de pedido y vacía el carrito
- [x] Carritos aislados: cada cuenta recupera el suyo
- [x] KPIs correctos: $4,090 ventas · 7 referidos · $319.50 comisiones · nivel Plata
- [x] `npm run lint` sin errores · `npm run build` correcto

---

## Hito 2 — Unidad 2, Fase 1 (Arquitectura y API robusta) · COMPLETADO

- [x] Paso 1 — `APIError` en `models/models.go` (`Status`, `Message`, `Details`)
- [x] Paso 2 — `Login` valida email y password vacíos con HTTP 400
- [x] Paso 3 — `GetProductoPorID` maneja el error de `strconv.Atoi` con HTTP 400
- [x] Paso 4 — `GET /api/health` registrado en `routes.go`
- [x] Extra — `repository/repository.go` como única fuente de datos
- [x] Contrato del frontend intacto (`/api/login`, `/api/productos`)
- [x] `go vet` y `go build` sin errores
- [x] 11 casos probados con curl (200 / 400 / 401 / 404)

---

## Hito 3 — Unidad 2, Fase 2 (Persistencia) · PENDIENTE

- [ ] Conectar PostgreSQL desde `backend/repository/`
- [ ] Migrar usuarios y productos a tablas reales
- [ ] Variables de entorno para la conexión (sin credenciales en el código)
- [ ] Reemplazar el token ficticio por JWT firmado
- [ ] Middleware que valide el JWT en las rutas protegidas
- [ ] Hash de contraseñas (bcrypt) en lugar de texto plano

## Hito 4 — Integración pendiente del frontend

- [ ] `productosService.ts` consume `GET /api/productos` en vez del mock
- [ ] El backend devuelve `descripcion`, `categoria` y `galeria`
- [ ] `services/api.ts` lee el campo `message` del nuevo `APIError`
- [ ] Endpoint `/api/red` para la red multinivel (hoy se calcula en el frontend)
- [ ] Persistir la sesión para que sobreviva a una recarga

## Rúbrica de evaluación

| Criterio | Peso |
|---|---|
| Funcionalidad de los flujos (compra + red MLM) | 30 % |
| Calidad del código (componentes, hooks, tipado) | 25 % |
| UX a pantalla completa (responsive + accesibilidad) | 25 % |
| Integración con API y estado global | 20 % |
