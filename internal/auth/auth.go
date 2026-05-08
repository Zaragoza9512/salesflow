package auth

import (
	"database/sql"
	"encoding/json"
	"net/http"

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
		// paso 1 - leer JSON
		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "datos inválidos", http.StatusBadRequest)
			return
		}
		// paso 2 - encriptar contraseña
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "error interno", http.StatusInternalServerError)
			return
		}
		//paso 3  - guardar el usuario en la base datos
		_, err = db.Exec(`
    		INSERT INTO users (nombre, correo, contrasena_hash)
    		VALUES ($1, $2, $3)`,
			req.Nombre, req.Correo, string(hash))
		if err != nil {
			http.Error(w, "error al guardar usuario", http.StatusInternalServerError)
			return
		}
		//paso 4  - responder al clinete que el registro fue exitoso
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"mensaje": "usuario registrado exitosamente",
		})
	} // ← cierra la función interna
} // ← cierra Register
