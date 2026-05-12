package main

import (
	"backend/handler"
	"backend/middleware"
	"backend/repository"
	"backend/router"
	"backend/service"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {

	godotenv.Load()
	connStr := fmt.Sprintf(
		"user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)
	var err error
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(25)                 // max connections to Postgres
	db.SetMaxIdleConns(25)                 // keep 25 ready
	db.SetConnMaxLifetime(5 * time.Minute) // recycle old connections
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		panic("JWT_SECRET envronment variable is required")
	}
	repo := &repository.TaskRepository{DB: db}
	rateLimiter := middleware.NewRateLimiter(100000, time.Minute)
	taskService := &service.TaskService{Repo: repo}
	taskHandler := &handler.TaskHandler{Service: taskService}

	authService := &service.AuthService{Repo: repo, Secret: jwtSecret}
	validate := validator.New()
	authHandler := &handler.AuthHandler{
		Service:  authService,
		Validate: validate,
	}

	router := router.SetupRouter(taskHandler, authHandler, jwtSecret, rateLimiter)
	fmt.Println("Server running on http://localhost:8080")
	err = http.ListenAndServe(":8080", router)
	if err != nil {
		panic(err)
	}
}
