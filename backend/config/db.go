// Paquete config: configuración de infraestructura (variables de entorno y conexiones).
package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// CargarEnv lee el archivo .env si existe. Cuando la app corre dentro de Docker
// las variables llegan del entorno, así que la ausencia del archivo no es un error.
func CargarEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("Aviso: no se encontró .env, se usarán las variables del entorno")
	}
}

// env devuelve la variable de entorno o un valor por defecto si está vacía.
func env(clave, porDefecto string) string {
	if valor := os.Getenv(clave); valor != "" {
		return valor
	}
	return porDefecto
}

// ConectarDB abre el pool de conexiones y verifica que la base responda.
// Devuelve el pool para inyectarlo donde haga falta, sin variables globales.
func ConectarDB() (*pgxpool.Pool, error) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=disable",
			env("DB_USER", "multicatalogo"),
			env("DB_PASSWORD", "multicatalogo"),
			env("DB_HOST", "localhost"),
			env("DB_PORT", "5432"),
			env("DB_NAME", "multicatalogo"),
		)
	}

	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("configuración de conexión inválida: %w", err)
	}

	// Límites del pool: suficiente para desarrollo y evita abrir conexiones de más.
	cfg.MaxConns = 10
	cfg.MinConns = 1
	cfg.MaxConnLifetime = time.Hour

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("no se pudo crear el pool de conexiones: %w", err)
	}

	// Ping: confirma que la base realmente responde antes de levantar el servidor.
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("no se pudo conectar a PostgreSQL: %w", err)
	}

	log.Println("Conexión a PostgreSQL establecida")
	return pool, nil
}
