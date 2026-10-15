// Implementaciones de CacheRepository.
package repository

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// Claves y tiempos de expiración de cada dato cacheado.
const (
	ClaveProductos = "productos:all"
	ClaveRedPrefix = "red:tree:"

	TTLProductos = 5 * time.Minute  // el catálogo cambia poco
	TTLRed       = 10 * time.Minute // armar el árbol es costoso
)

// RedisCache guarda los datos en Redis.
type RedisCache struct {
	cliente *redis.Client
}

// NuevoRedisCache recibe el cliente por inyección.
func NuevoRedisCache(cliente *redis.Client) *RedisCache {
	return &RedisCache{cliente: cliente}
}

func (c *RedisCache) Activa() bool { return c.cliente != nil }

// Obtener devuelve false en un Cache Miss o si el JSON guardado está corrupto.
func (c *RedisCache) Obtener(ctx context.Context, clave string, destino any) bool {
	datos, err := c.cliente.Get(ctx, clave).Bytes()
	if err != nil {
		// redis.Nil (la clave no existe) es el caso normal de un Cache Miss.
		return false
	}

	if err := json.Unmarshal(datos, destino); err != nil {
		log.Printf("caché corrupta en %q, se descarta: %v", clave, err)
		c.cliente.Del(ctx, clave)
		return false
	}

	log.Printf("CACHE HIT  %s", clave)
	return true
}

// Guardar serializa el valor y lo almacena con su expiración.
// Un fallo aquí no interrumpe la respuesta al cliente.
func (c *RedisCache) Guardar(ctx context.Context, clave string, valor any, ttl time.Duration) {
	datos, err := json.Marshal(valor)
	if err != nil {
		log.Printf("no se pudo serializar %q para la caché: %v", clave, err)
		return
	}

	if err := c.cliente.Set(ctx, clave, datos, ttl).Err(); err != nil {
		log.Printf("no se pudo guardar %q en la caché: %v", clave, err)
		return
	}

	log.Printf("CACHE MISS %s -> guardado por %s", clave, ttl)
}

func (c *RedisCache) Eliminar(ctx context.Context, claves ...string) {
	if len(claves) == 0 {
		return
	}
	if err := c.cliente.Del(ctx, claves...).Err(); err != nil {
		log.Printf("no se pudo invalidar la caché: %v", err)
		return
	}
	log.Printf("CACHE DEL  %v", claves)
}

// EliminarPatron borra todas las claves que coincidan (por ejemplo red:tree:*).
// Usa SCAN en lugar de KEYS para no bloquear el servidor de Redis.
func (c *RedisCache) EliminarPatron(ctx context.Context, patron string) {
	var cursor uint64
	borradas := 0

	for {
		claves, siguiente, err := c.cliente.Scan(ctx, cursor, patron, 50).Result()
		if err != nil {
			log.Printf("no se pudo recorrer la caché con %q: %v", patron, err)
			return
		}
		if len(claves) > 0 {
			c.cliente.Del(ctx, claves...)
			borradas += len(claves)
		}
		if siguiente == 0 {
			break
		}
		cursor = siguiente
	}

	if borradas > 0 {
		log.Printf("CACHE DEL  %d clave(s) que coinciden con %s", borradas, patron)
	}
}

// SinCache es la implementación que se usa cuando Redis no está disponible.
// Evita comprobaciones de nil repartidas por todo el código (patrón Null Object).
type SinCache struct{}

func (SinCache) Activa() bool                                        { return false }
func (SinCache) Obtener(context.Context, string, any) bool           { return false }
func (SinCache) Guardar(context.Context, string, any, time.Duration) {}
func (SinCache) Eliminar(context.Context, ...string)                 {}
func (SinCache) EliminarPatron(context.Context, string)              {}
