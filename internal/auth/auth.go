package auth

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Nombre   string `json:"nombre"`
	Correo   string `json:"correo"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Correo   string `json:"correo"`
	Password string `json:"password"`
}

func Register(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// paso 1 - leer JSON que manda el cliente
		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "datos inválidos", http.StatusBadRequest)
			return
		}
		// paso 2 - encriptar contraseña con bcrypt antes de guardarla
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "error interno", http.StatusInternalServerError)
			return
		}
		// paso 3 - guardar el usuario en la base de datos
		_, err = db.Exec(`
			INSERT INTO users (nombre, correo, contrasena_hash)
			VALUES ($1, $2, $3)`,
			req.Nombre, req.Correo, string(hash))
		if err != nil {
			http.Error(w, "error al guardar usuario", http.StatusInternalServerError)
			return
		}
		// paso 4 - responder al cliente que el registro fue exitoso
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"mensaje": "usuario registrado exitosamente",
		})
	} // ← cierra la función interna
} // ← cierra Register

func Login(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// paso 1 - leer JSON que manda el cliente
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "datos inválidos", http.StatusBadRequest)
			return
		}
		// paso 2 - buscar el usuario en la base de datos por correo
		var id, contrasenaHash string
		err := db.QueryRow(`
			SELECT id, contrasena_hash FROM users WHERE correo = $1`,
			req.Correo).Scan(&id, &contrasenaHash)
		if err != nil {
			http.Error(w, "credenciales inválidas", http.StatusUnauthorized)
			return
		}
		// paso 3 - comparar la contraseña con el hash guardado
		if err := bcrypt.CompareHashAndPassword([]byte(contrasenaHash), []byte(req.Password)); err != nil {
			http.Error(w, "credenciales inválidas", http.StatusUnauthorized)
			return
		}
		// paso 4 - generar el token JWT con el id del usuario y expiración de 24hrs
		secret := os.Getenv("JWT_SECRET")
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": id,
			"exp":     time.Now().Add(24 * time.Hour).Unix(),
		})
		tokenString, err := token.SignedString([]byte(secret))
		if err != nil {
			http.Error(w, "error generando token", http.StatusInternalServerError)
			return
		}
		// paso 5 - devolver el token al cliente
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"token": tokenString,
		})
	} // ← cierra la función interna
} // ← cierra Login
