// Cliente de Redis usado como caché de lectura (patrón Cache-Aside).
package config

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// ConectarRedis abre la conexión al servidor de caché.
// Devuelve nil si Redis no responde: la caché es una optimización, no un requisito,
// y la aplicación sigue funcionando consultando siempre a PostgreSQL.
func ConectarRedis() *redis.Client {
	url := env("REDIS_URL", "")

	var opts *redis.Options
	if url != "" {
		parsed, err := redis.ParseURL(url)
		if err != nil {
			log.Printf("Aviso: REDIS_URL inválida (%v); la caché queda desactivada", err)
			return nil
		}
		opts = parsed
	} else {
		db, _ := strconv.Atoi(env("REDIS_DB", "0"))
		opts = &redis.Options{
			Addr:     fmt.Sprintf("%s:%s", env("REDIS_HOST", "localhost"), env("REDIS_PORT", "6379")),
			Password: env("REDIS_PASSWORD", ""),
			DB:       db,
		}
	}

	// Sin reintentos y con timeouts cortos: si Redis no está, conviene saberlo rápido
	// en lugar de llenar el log de intentos fallidos.
	opts.MaxRetries = -1
	opts.DialTimeout = 2 * time.Second

	cliente := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := cliente.Ping(ctx).Err(); err != nil {
		log.Printf("Aviso: Redis no disponible (%v); se consultará siempre PostgreSQL", err)
		_ = cliente.Close()
		return nil
	}

	log.Println("Conexión a Redis establecida")
	return cliente
}
