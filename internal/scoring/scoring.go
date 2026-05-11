package scoring

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// groqRequest define la estructura del mensaje que mandamos a Groq
// Model es el cerebro que va a responder, Messages es la lista de mensajes
type groqRequest struct {
	Model    string        `json:"model"`
	Messages []groqMessage `json:"messages"`
}

// groqMessage representa un mensaje individual en la conversación
// Role puede ser "user" (nosotros) o "assistant" (la IA)
type groqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// groqResponse define la estructura de la respuesta que nos manda Groq
// Choices es una lista de respuestas — nosotros usamos solo la primera
type groqResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// ScoreLead califica un lead automáticamente usando reglas fijas + IA
// se ejecuta en segundo plano cuando se crea un lead nuevo
func ScoreLead(db *sql.DB, leadID, nombre, canal, tipoCredito, zonaInteres string, montoCredito float64) error {

	// paso 1 - calcular el score base con reglas fijas
	// cada campo completo suma puntos según su importancia
	score := 0

	// tipo de crédito es el dato más importante — define si puede comprar
	if tipoCredito != "" {
		score += 30
	}

	// monto de crédito define qué tipo de propiedad puede comprar
	// rangos basados en precios reales del mercado en Edomex
	if montoCredito >= 900000 {
		score += 20 // monto ideal, muchas opciones disponibles
	} else if montoCredito >= 750000 {
		score += 15 // alcanza propiedad nueva
	} else if montoCredito >= 650000 {
		score += 10 // solo propiedad de uso
	} else if montoCredito < 650000 {
		score += 3 // monto muy bajo, pocas opciones
	}

	// zona de interés definida significa que el lead sabe lo que quiere
	if zonaInteres != "" {
		score += 10
	}

	// tener nombre significa que ya dio información personal — mayor intención
	if nombre != "" {
		score += 10
	}

	// canal de origen — referido vale más porque ya viene con confianza
	switch canal {
	case "WhatsApp":
		score += 15
	case "Referido":
		score += 20
	case "Instagram", "Facebook":
		score += 10
	}

	// paso 2 - definir la categoría basada en el score total
	// HOT = contactar hoy, WARM = esta semana, COLD = cuando haya tiempo
	categoria := "COLD"
	if score >= 70 {
		categoria = "HOT"
	} else if score >= 40 {
		categoria = "WARM"
	}

	// paso 3 - construir el prompt con contexto del mercado inmobiliario
	// este prompt incluye tu conocimiento real del mercado de Edomex y CDMX
	prompt := fmt.Sprintf(`Eres un experto en bienes raíces en México especializado en créditos INFONAVIT y FOVISSSTE.
	Analiza este lead y genera un reasoning corto de 2-3 oraciones:
	Datos del lead:
	- Nombre: %s
	- Canal: %s
	- Tipo de crédito: %s
	- Monto de crédito: $%.0f
	- Zona de interés: %s
	- Score calculado: %d/100
	- Categoría: %s
	Contexto del mercado:
	- Mínimo para propiedad nueva en Edomex: $750,000
	- Mínimo para propiedad de uso en Edomex: $650,000
	- Mínimo para CDMX: $1,500,000
	- Si el monto es bajo, menciona urgencia y que los precios suben constantemente
	Responde solo el reasoning, sin saludos ni explicaciones adicionales.`,
		nombre, canal, tipoCredito, montoCredito, zonaInteres, score, categoria)

	// paso 4 - llamar a Groq para generar el reasoning en lenguaje natural
	// leer la API key del .env
	apiKey := os.Getenv("GROQ_API_KEY")

	// construir el mensaje usando las structs definidas arriba
	reqBody := groqRequest{
		Model: "llama-3.3-70b-versatile",
		Messages: []groqMessage{
			{Role: "user", Content: prompt},
		},
	}

	// convertir la struct a JSON — el formato que Groq entiende
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	// preparar la petición HTTP con la URL de Groq
	req, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	// agregar headers — tipo de contenido y credencial de acceso
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// crear el cliente HTTP y enviar la petición
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close() // cerrar la respuesta al terminar

	// leer y convertir la respuesta de Groq a nuestra struct
	json.NewDecoder(resp.Body).Decode(&groqResp)

	// extraer el texto generado por la IA
	reasoning := "Sin reasoning disponible"
	if len(groqResp.Choices) > 0 {
		reasoning = groqResp.Choices[0].Message.Content
	}

	// paso 5 - guardar el score en la tabla scores para tener historial
	_, err = db.Exec(`
		INSERT INTO scores (lead_id, score, categoria, reasoning)
		VALUES ($1, $2, $3, $4)`,
		leadID, score, categoria, reasoning)
	if err != nil {
		return err
	}

	// actualizar el score directamente en la tabla leads
	// así cuando listes los leads ya vienen con su score actualizado
	_, err = db.Exec(`
		UPDATE leads SET score=$1 WHERE id=$2`,
		score, leadID)
	if err != nil {
		return err
	}

	return nil // todo salió bien
}
