package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend/handler"
	"backend/middleware"
	"backend/repository"
	"backend/router"
	"backend/service"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"
)

const testSecret = "test-secret"

func generateTestToken(userID int) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": float64(userID),
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString([]byte(testSecret))
	return tokenString
}

func setupTestRouter() http.Handler {
	db, err := sql.Open("postgres", "user=postgres password=postgres dbname=taskdb sslmode=disable")
	if err != nil {
		panic(err)
	}

	validate := validator.New()
	repo := &repository.TaskRepository{DB: db}
	rateLimiter := middleware.NewRateLimiter(100, time.Minute)
	taskService := &service.TaskService{Repo: repo}
	taskHandler := &handler.TaskHandler{Service: taskService}

	authService := &service.AuthService{Repo: repo, Secret: testSecret}
	authHandler := &handler.AuthHandler{Service: authService, Validate: validate}

	return router.SetupRouter(taskHandler, authHandler, testSecret, rateLimiter)
}

func authHeader() string {
	return "Bearer " + generateTestToken(1)
}

func TestRegister(t *testing.T) {
	r := setupTestRouter()

	body := []byte(`{"username":"testuser123","password":"securepass"}`)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	fmt.Printf("TestRegister: status=%d body=%s\n", w.Code, w.Body.String())

	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest {
		t.Fatalf("expected 200 or 400 got %d", w.Code)
	}
}

func TestLogin(t *testing.T) {
	r := setupTestRouter()

	body := []byte(`{"username":"testuser123","password":"securepass"}`)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	fmt.Printf("TestLogin: status=%d body=%s\n", w.Code, w.Body.String())

	if w.Code != http.StatusOK && w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 200 or 401 got %d", w.Code)
	}
}

func TestCreateTask(t *testing.T) {
	r := setupTestRouter()

	body := []byte(`{"name":"learn go"}`)
	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader())
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	fmt.Printf("TestCreateTask: status=%d body=%s\n", w.Code, w.Body.String())

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Code)
	}
}

func TestGetTasks(t *testing.T) {
	r := setupTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	req.Header.Set("Authorization", authHeader())
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	fmt.Printf("TestGetTasks: status=%d body=%s\n", w.Code, w.Body.String())

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Code)
	}
}

func TestDeleteTask(t *testing.T) {
	r := setupTestRouter()

	req := httptest.NewRequest(http.MethodDelete, "/tasks/1", nil)
	req.Header.Set("Authorization", authHeader())
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	fmt.Printf("TestDeleteTask: status=%d body=%s\n", w.Code, w.Body.String())

	if w.Code != http.StatusNoContent && w.Code != http.StatusNotFound {
		t.Fatalf("unexpected status %d", w.Code)
	}
}

func TestUpdateTask(t *testing.T) {
	r := setupTestRouter()

	body := []byte(`{"name":"updated task","done":true}`)
	req := httptest.NewRequest(http.MethodPut, "/tasks/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader())
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	fmt.Printf("TestUpdateTask: status=%d body=%s\n", w.Code, w.Body.String())

	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Fatalf("unexpected status %d", w.Code)
	}
}

func TestCreateTaskUnauthorized(t *testing.T) {
	r := setupTestRouter()

	body := []byte(`{"name":"learn go"}`)
	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	fmt.Printf("TestCreateTaskUnauthorized: status=%d body=%s\n", w.Code, w.Body.String())

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 got %d", w.Code)
	}
}
func TestRateLimit(t *testing.T) {
	r := setupTestRouter()

	for i := 0; i < 101; i++ {
		req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
		req.Header.Set("Authorization", authHeader())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if i == 100 {
			fmt.Printf("TestRateLimit: request #%d status=%d\n", i+1, w.Code)
			if w.Code != http.StatusTooManyRequests {
				t.Fatalf("expected 429 on request %d got %d", i+1, w.Code)
			}
		}
	}
}
