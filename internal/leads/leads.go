package leads

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Zaragoza9512/salesflow/internal/middleware"
	"github.com/Zaragoza9512/salesflow/internal/scoring"
	"github.com/Zaragoza9512/salesflow/pkg/cache"
	"github.com/go-chi/chi/v5"
)

// LeadRequest define los campos que el cliente manda al crear o actualizar un lead
type LeadRequest struct {
	Nombre                  string  `json:"nombre"`
	Telefono                string  `json:"telefono"`
	Correo                  string  `json:"correo"`
	Canal                   string  `json:"canal"`
	MontoCredito            float64 `json:"monto_credito"`
	TipoCredito             string  `json:"tipo_credito"`
	ZonaInteres             string  `json:"zona_interes"`
	CaracteristicasVivienda string  `json:"caracteristicas_vivienda"`
}

// Create crea un nuevo lead en la base de datos
func Create(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// paso 1 - obtener el user_id del token JWT via contexto
		// así sabemos a qué asesor pertenece este lead
		userID := r.Context().Value(middleware.UserIDKey).(string)

		// paso 2 - leer el JSON que manda el cliente
		var req LeadRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "datos inválidos", http.StatusBadRequest)
			return
		}

		// paso 3 - insertar el lead en la base de datos
		// RETURNING id devuelve el UUID generado automáticamente
		var id string
		err := db.QueryRow(`
			INSERT INTO leads (user_id, nombre, telefono, correo, canal, monto_credito, tipo_credito, zona_interes, caracteristicas_vivienda)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id`,
			userID, req.Nombre, req.Telefono, req.Correo, req.Canal,
			req.MontoCredito, req.TipoCredito, req.ZonaInteres, req.CaracteristicasVivienda).Scan(&id)
		if err != nil {
			http.Error(w, "error al crear lead", http.StatusInternalServerError)
			return
		}

		// paso 4 - calcular el score del lead en segundo plano
		// "go" significa que se ejecuta sin bloquear la respuesta al cliente
		// el cliente recibe su respuesta inmediatamente mientras Groq genera el reasoning
		go scoring.ScoreLead(db, id, req.Nombre, req.Canal, req.TipoCredito, req.ZonaInteres, req.MontoCredito)

		// paso 4 - responder con el id del lead creado
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"id":      id,
			"mensaje": "lead creado exitosamente",
		})
	}
}

// List devuelve todos los leads del asesor ordenados por score de mayor a menor
// así el asesor ve primero los leads más calientes
func List(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// paso 1 - obtener el user_id del contexto
		userID := r.Context().Value(middleware.UserIDKey).(string)

		// paso 2 - consultar todos los leads del asesor
		// ORDER BY score DESC = los más calientes primero
		rows, err := db.Query(`
			SELECT id, nombre, telefono, correo, canal, estado, score, created_at
			FROM leads
			WHERE user_id = $1
			ORDER BY score DESC`,
			userID)
		if err != nil {
			http.Error(w, "error al obtener leads", http.StatusInternalServerError)
			return
		}
		defer rows.Close() // cierra el cursor al terminar

		// paso 3 - recorrer los resultados y armar la lista
		var leads []map[string]interface{}
		for rows.Next() {
			var id, nombre, telefono, correo, canal, estado, createdAt string
			var score int
			rows.Scan(&id, &nombre, &telefono, &correo, &canal, &estado, &score, &createdAt)
			leads = append(leads, map[string]interface{}{
				"id":         id,
				"nombre":     nombre,
				"telefono":   telefono,
				"correo":     correo,
				"canal":      canal,
				"estado":     estado,
				"score":      score,
				"created_at": createdAt,
			})
		}

		// paso 4 - responder con la lista completa
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(leads)
	}
}

// GetByID devuelve el detalle completo de un lead específico
// solo puede verlo el asesor dueño del lead (AND user_id = $2)
func GetByID(db *sql.DB, c *cache.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(middleware.UserIDKey).(string)
		id := chi.URLParam(r, "id")

		// intentar obtener el lead del cache primero
		cacheKey := "lead:" + id
		if cached, err := c.Get(cacheKey); err == nil {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(cached))
			return
		}

		// buscar el lead — ahora con todos los campos
		var nombre, telefono, correo, canal, estado, createdAt string
		var tipoCredito, zonaInteres string
		var montoCredito float64
		var score int

		err := db.QueryRow(`
            SELECT nombre, telefono, correo, canal, estado, score, created_at,
                   tipo_credito, monto_credito, zona_interes
            FROM leads
            WHERE id = $1 AND user_id = $2`,
			id, userID).Scan(&nombre, &telefono, &correo, &canal, &estado, &score, &createdAt,
			&tipoCredito, &montoCredito, &zonaInteres)
		if err != nil {
			http.Error(w, "lead no encontrado", http.StatusNotFound)
			return
		}

		// construir la respuesta con todos los campos
		response, _ := json.Marshal(map[string]interface{}{
			"id":            id,
			"nombre":        nombre,
			"telefono":      telefono,
			"correo":        correo,
			"canal":         canal,
			"estado":        estado,
			"score":         score,
			"created_at":    createdAt,
			"tipo_credito":  tipoCredito,
			"monto_credito": montoCredito,
			"zona_interes":  zonaInteres,
		})

		// guardar en cache por 24 horas
		c.Set(cacheKey, string(response), 24*time.Hour)

		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	}
}

// Update actualiza los datos de un lead existente
// updated_at=NOW() registra cuándo fue la última modificación
func Update(db *sql.DB, c *cache.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(middleware.UserIDKey).(string)
		id := chi.URLParam(r, "id")

		// leer los nuevos datos del cliente
		var req LeadRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "datos inválidos", http.StatusBadRequest)
			return
		}

		// actualizar el lead en la base de datos
		// solo actualiza si el lead pertenece al asesor (AND user_id=$10)
		_, err := db.Exec(`
			UPDATE leads
			SET nombre=$1, telefono=$2, correo=$3, canal=$4,
				monto_credito=$5, tipo_credito=$6, zona_interes=$7,
				caracteristicas_vivienda=$8, updated_at=NOW()
			WHERE id=$9 AND user_id=$10`,
			req.Nombre, req.Telefono, req.Correo, req.Canal,
			req.MontoCredito, req.TipoCredito, req.ZonaInteres,
			req.CaracteristicasVivienda, id, userID)
		if err != nil {
			http.Error(w, "error al actualizar lead", http.StatusInternalServerError)
			return
		}

		// invalidar el cache del lead actualizado
		c.Delete("lead:" + id)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"mensaje": "lead actualizado exitosamente",
		})
	}
}

// Delete archiva un lead — no lo borra físicamente
// en un CRM nunca se pierden datos, solo se cambia el estado
func Delete(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(middleware.UserIDKey).(string)
		id := chi.URLParam(r, "id")

		// cambiar estado a "archivado" en lugar de borrar el registro
		_, err := db.Exec(`
			UPDATE leads
			SET estado='archivado', updated_at=NOW()
			WHERE id=$1 AND user_id=$2`,
			id, userID)
		if err != nil {
			http.Error(w, "error al archivar lead", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"mensaje": "lead archivado exitosamente",
		})
	}
}
