package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// tipo propio para las claves del contexto — evita conflictos con otros paquetes
type contextKey string

// UserIDKey es la clave con la que guardamos el user_id en el contexto
// los endpoints la usan para saber qué usuario está haciendo la petición
const UserIDKey contextKey = "user_id"

// AuthMiddleware es el guardia de seguridad — verifica el token JWT
// antes de dejar pasar la petición al endpoint
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// paso 1 - leer el header Authorization de la petición
		// si no viene el header, rechaza la petición con error 401
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "token requerido", http.StatusUnauthorized)
			return
		}

		// quitar la palabra "Bearer " y quedarse solo con el token
		// "Bearer eyJhbGci..." → "eyJhbGci..."
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// paso 2 - verificar que el token es válido
		// usa el JWT_SECRET para comprobar que el token no fue modificado
		// y que no ha expirado
		secret := os.Getenv("JWT_SECRET")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "token inválido", http.StatusUnauthorized)
			return
		}

		// paso 3 - extraer el user_id guardado dentro del token
		claims := token.Claims.(jwt.MapClaims)
		userID := claims["user_id"].(string)

		// agregar el user_id al contexto de la petición
		// así los endpoints pueden saber quién está haciendo la petición
		ctx := context.WithValue(r.Context(), UserIDKey, userID)

		// dejar pasar la petición al endpoint con el contexto actualizado
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
