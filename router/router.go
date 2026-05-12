package router

import (
	"backend/handler"
	"backend/middleware"
	"net/http"
)

func SetupRouter(h *handler.TaskHandler, ah *handler.AuthHandler, secret string, rl *middleware.RateLimiter) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/register", middleware.LoggerMiddleware(ah.Register))
	mux.HandleFunc("/login", middleware.LoggerMiddleware(ah.Login))

	mux.HandleFunc("/tasks", rl.Middleware(middleware.LoggerMiddleware(middleware.AuthMiddleware(secret, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.GetTasks(w, r)
		case http.MethodPost:
			h.CreateTask(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))))
	mux.HandleFunc("/tasks/", rl.Middleware(middleware.LoggerMiddleware(middleware.AuthMiddleware(secret, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			h.UpdateTask(w, r)
		case http.MethodDelete:
			h.DeleteTask(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))))
	return mux
}
