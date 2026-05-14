package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Zaragoza9512/salesflow/internal/auth"
	"github.com/Zaragoza9512/salesflow/internal/database"
	"github.com/Zaragoza9512/salesflow/internal/leads"
	customMiddleware "github.com/Zaragoza9512/salesflow/internal/middleware"
	cachekg "github.com/Zaragoza9512/salesflow/pkg/cache"
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

	c := cachekg.NewCache()

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

	// servir archivos estáticos del frontend
	fs := http.FileServer(http.Dir("./web"))
	r.Handle("/*", fs)

	// rutas protegidas — requieren token JWT válido
	r.Group(func(r chi.Router) {
		r.Use(customMiddleware.AuthMiddleware)
		// aquí van las rutas de leads
		r.Post("/leads", leads.Create(db))
		r.Get("/leads", leads.List(db))
		r.Get("/leads/{id}", leads.GetByID(db, c))
		r.Put("/leads/{id}", leads.Update(db, c))
		r.Delete("/leads/{id}", leads.Delete(db))
	})

	log.Fatal(http.ListenAndServe(":"+port, r))
}
