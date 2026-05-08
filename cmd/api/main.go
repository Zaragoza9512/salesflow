package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Zaragoza9512/salesflow/internal/auth"
	"github.com/Zaragoza9512/salesflow/internal/database"
	customMiddleware "github.com/Zaragoza9512/salesflow/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	db := database.Connect()
	defer db.Close()

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// rutas públicas — no requieren token
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","service":"salesflow"}`)
	})
	r.Post("/auth/register", auth.Register(db))
	r.Post("/auth/login", auth.Login(db))

	// rutas protegidas — requieren token JWT válido
	r.Group(func(r chi.Router) {
		r.Use(customMiddleware.AuthMiddleware)
		// aquí van las rutas de leads
	})

	log.Fatal(http.ListenAndServe(":"+port, r))
}
