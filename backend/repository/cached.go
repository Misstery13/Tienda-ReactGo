// Decorador que agrega caché a cualquier DBRepository (patrón Cache-Aside).
// La lógica de caché queda fuera del código SQL y se puede quitar
// simplemente dejando de envolver el repositorio.
package repository

import (
	"context"

	"multicatalogo-backend/models"
)

// CachedRepository implementa DBRepository delegando en otro DBRepository.
type CachedRepository struct {
	DBRepository                 // lecturas y escrituras que no se cachean
	cache        CacheRepository // almacén de caché inyectado
}

// NuevoCachedRepository envuelve el repositorio real con su caché.
func NuevoCachedRepository(inner DBRepository, cache CacheRepository) *CachedRepository {
	return &CachedRepository{DBRepository: inner, cache: cache}
}

// ObtenerProductos aplica Cache-Aside sobre el catálogo.
func (r *CachedRepository) ObtenerProductos(ctx context.Context) ([]models.Producto, error) {
	// 1. Cache Hit: se responde sin tocar PostgreSQL.
	var enCache []models.Producto
	if r.cache.Obtener(ctx, ClaveProductos, &enCache) {
		return enCache, nil
	}

	// 2. Cache Miss: se consulta la base y se guarda el resultado.
	productos, err := r.DBRepository.ObtenerProductos(ctx)
	if err != nil {
		return nil, err
	}

	r.cache.Guardar(ctx, ClaveProductos, productos, TTLProductos)
	return productos, nil
}

// ObtenerRedDe aplica Cache-Aside sobre el árbol, con una clave por usuario.
func (r *CachedRepository) ObtenerRedDe(ctx context.Context, emailUsuario string) (*models.Referido, error) {
	clave := ClaveRedPrefix + emailUsuario

	// 1. Cache Hit: se evita reconstruir el árbol completo.
	var enCache models.Referido
	if r.cache.Obtener(ctx, clave, &enCache) {
		return &enCache, nil
	}

	// 2. Cache Miss: se arma desde PostgreSQL y se guarda con un TTL más largo.
	raiz, err := r.DBRepository.ObtenerRed(ctx)
	if err != nil {
		return nil, err
	}

	r.cache.Guardar(ctx, clave, raiz, TTLRed)
	return raiz, nil
}
