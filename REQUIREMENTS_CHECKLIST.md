# REQUIREMENTS_CHECKLIST — Tienda ReactGo (MultiCatálogo)

Fuente de verdad del progreso. **Léelo antes de explorar código.**
Al completar una tarea verificada, marca su casilla y actualiza el Estado General.

> Fuentes: guion de la Práctica 03 (Unidad 1, Tema 5) y documento de Fase 1
> (Unidad 2, Arquitectura y API robusta). El sílabo no está en el repositorio,
> así que los requisitos de evaluación se transcriben desde la rúbrica del guion.

## Estado General

- **Hito actual:** Unidad 2 completa (Fases 1 a 4)
- **Última tarea completada:** Fase 4 verificada: checkout transaccional, goroutine de comisiones e invalidación de caché
- **Siguiente tarea pendiente:** Seguridad (Hito 5): el backend no valida quién llama a sus endpoints
- **Última actualización:** 2026-10-15

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

## Hito 3 — Unidad 2, Fase 2 (Persistencia PostgreSQL) · COMPLETADO

- [x] Paso 1 — `docker-compose.yml` con PostgreSQL 16 y variables de entorno
- [x] Paso 2 — `config/db.go` con pool `pgxpool`, carga de `.env` y cierre ordenado
- [x] Paso 3 — `database/init.sql` con `usuarios`, `productos` y `referidos` (parent_id autorreferenciado) + datos semilla
- [x] Paso 4 — `repository.go` sin variables globales; todo sale de `SELECT`
- [x] `ObtenerRed` reconstruye el árbol multinivel desde la lista plana
- [x] Endpoint `GET /api/red` que expone la jerarquía
- [x] Verificado: `UPDATE productos SET precio = 99.99 WHERE id = 1` se refleja en la API
- [x] Apagado ordenado del pool con Ctrl+C (SIGINT/SIGTERM)

## Hito 4 — Unidad 2, Fase 3 (Caché con Redis) · COMPLETADO

- [x] Paso 1 — Servicio `redis:alpine` agregado al `docker-compose.yml`
- [x] Paso 2 — `config/redis.go` con cliente `go-redis/v9` y variables de entorno
- [x] Paso 3 — Cache-Aside en el catálogo con clave `productos:all` y TTL de 5 min
- [x] Paso 4 — Cache-Aside en la red con clave `red:tree:{email}` y TTL de 10 min
- [x] La API degrada con elegancia: sin Redis responde igual desde PostgreSQL
- [x] Medido: `/api/red` pasa de 11.2 ms (Miss) a 0.6 ms (Hit) — unas 19 veces más rápido
- [x] Verificadas las claves y sus TTL en Redis (`productos:all` 300 s, `red:tree:{email}` 600 s)
- [ ] Invalidar la caché cuando existan endpoints de escritura

## Hito 5 — Unidad 2, Fase 4 (Transacciones y concurrencia) · COMPLETADO

- [x] Paso 1 — Inyección de dependencias: `DBRepository` y `CacheRepository` como interfaces
- [x] `PostgresRepository`, `RedisCache` y `SinCache` (Null Object) inyectados desde `main.go`
- [x] `CachedRepository` como decorador: el Cache-Aside ya no vive dentro del SQL
- [x] Controladores convertidos en structs con sus dependencias
- [x] Paso 2 — `POST /api/checkout` recibe el carrito de `Checkout.tsx`
- [x] Paso 3 — Transacción ACID: `BEGIN`, pedido, items, descuento de stock, `COMMIT`
- [x] `defer tx.Rollback(ctx)` revierte todo ante cualquier fallo
- [x] Los precios se leen de PostgreSQL, nunca del cliente
- [x] Paso 4 — Goroutine que acredita comisiones (10 %, 5 %, 2 %) sin bloquear la respuesta
- [x] Paso 5 — Invalidación de `red:tree:*` tras actualizar las comisiones
- [x] `SIMULAR_FALLA_CHECKOUT=true` para demostrar el ROLLBACK en clase
- [x] Verificado: pedido MC-000005 de $150.50 desde el navegador
- [x] Verificado: KPIs del Dashboard suben solos de $4,090 a $4,358.50
- [x] Verificado: id inexistente, stock insuficiente y falla simulada no dejan pedidos fantasma

## Hito 7 — Seguridad (PENDIENTE, ningún punto implementado)

> El backend no valida quién llama a sus endpoints: las guardas de React se saltan con `curl`.

- [ ] **Crítico:** middleware que exija autenticación en `/api/red` y `/api/checkout`
- [ ] **Crítico:** reemplazar el token ficticio por JWT firmado con expiración
- [ ] **Crítico:** IDOR en `/api/red` — la identidad debe salir del token, no del query param
- [ ] Hash de contraseñas con bcrypt en lugar de texto plano
- [ ] Límite de intentos en `/api/login` contra fuerza bruta

## Hito 8 — Funcionalidad que falta al sistema

- [ ] CRUD de productos (el panel de administración solo lee)
- [ ] Historial de pedidos para el cliente (`GET /api/pedidos`)
- [ ] Persistir la sesión para que sobreviva a una recarga
- [ ] Alta de referidos y enlace de invitación
- [ ] Registro de usuarios y recuperación de contraseña
- [ ] Estados del pedido (pendiente, pagado, enviado, entregado)
- [ ] Mostrar y respetar el stock en el catálogo
- [ ] Búsqueda y paginación en el servidor
- [ ] `/api/health` debe verificar PostgreSQL y Redis, hoy responde "ok" siempre
- [ ] Pruebas automatizadas (la inyección de dependencias ya las habilita)

## Hito 6 — Integración del frontend con la API · COMPLETADO

- [x] El backend devuelve `descripcion`, `categoria` y `galeria`
- [x] Endpoint `/api/red` para la red multinivel
- [x] `productosService.ts` consume `GET /api/productos` y `GET /api/productos/:id`
- [x] `redService.ts` consume `GET /api/red?email=`
- [x] `MiRed.tsx` y `Dashboard.tsx` leen el árbol de la API
- [x] `Catalogo.tsx` y `Storefront.tsx` derivan las categorías de los datos reales
- [x] `services/api.ts` lee el campo `message` del `APIError`
- [x] Eliminados `productosMock` y `redInicial`: no queda ningún dato quemado
- [x] Verificado: `UPDATE` en PostgreSQL se refleja en la vista de React
- [x] `Checkout.tsx` envía la compra a `POST /api/checkout`

## Rúbrica de evaluación

| Criterio | Peso |
|---|---|
| Funcionalidad de los flujos (compra + red MLM) | 30 % |
| Calidad del código (componentes, hooks, tipado) | 25 % |
| UX a pantalla completa (responsive + accesibilidad) | 25 % |
| Integración con API y estado global | 20 % |
