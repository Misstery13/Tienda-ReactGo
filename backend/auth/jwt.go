// Paquete auth: emisión y verificación de tokens JWT.
package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"multicatalogo-backend/models"
)

// DuracionToken es la vigencia de una sesión.
const DuracionToken = 2 * time.Hour

// ErrTokenInvalido agrupa cualquier problema de validación del token.
var ErrTokenInvalido = errors.New("token inválido o expirado")

// Claims son los datos que viajan firmados dentro del token.
// Subject guarda el id del usuario; email y rol evitan consultar la base en cada petición.
type Claims struct {
	Email string `json:"email"`
	Rol   string `json:"rol"`
	jwt.RegisteredClaims
}

// GenerarToken firma un JWT HS256 para el usuario indicado.
func GenerarToken(usuario models.Usuario, secreto string) (string, error) {
	ahora := time.Now()

	claims := Claims{
		Email: usuario.Email,
		Rol:   usuario.Rol,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", usuario.ID),
			IssuedAt:  jwt.NewNumericDate(ahora),
			ExpiresAt: jwt.NewNumericDate(ahora.Add(DuracionToken)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secreto))
}

// VerificarToken comprueba la firma y la expiración, y devuelve los claims.
func VerificarToken(tokenString, secreto string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		// Se rechaza cualquier algoritmo distinto de HS256 (ataque de confusión de algoritmo).
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("algoritmo de firma inesperado: %v", t.Header["alg"])
		}
		return []byte(secreto), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

	if err != nil || !token.Valid {
		return nil, ErrTokenInvalido
	}

	return claims, nil
}
