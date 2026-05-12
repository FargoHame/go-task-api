package middleware

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}
type client struct {
	count    int
	lastSeen time.Time
}
type RateLimiter struct {
	clients map[string]*client
	mu      sync.Mutex
	limit   int
	window  time.Duration
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}
func AuthMiddleware(secret string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			http.Error(w, "invalid authorization header", http.StatusUnauthorized)
			return
		}
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "invalid token claims", http.StatusUnauthorized)
			return
		}
		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			http.Error(w, "invalud token user_id", http.StatusUnauthorized)
			return
		}
		r = r.WithContext(context.WithValue(r.Context(), "user_id", int(userIDFloat)))
		next(w, r)
	}
}

func LoggerMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{w, http.StatusOK}
		next(wrapped, r)
		log.Printf(
			"%s %s %d %v ",
			r.Method,
			r.URL.Path,
			wrapped.status,
			time.Since(start),
		)
	}
}
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		clients: make(map[string]*client),
		limit:   limit,
		window:  window,
	}
}
func (rl *RateLimiter) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}
		rl.mu.Lock()
		defer rl.mu.Unlock()
		c, exists := rl.clients[ip]
		if !exists {
			rl.clients[ip] = &client{
				count:    1,
				lastSeen: time.Now(),
			}
			next(w, r)
			return
		}
		if time.Since(c.lastSeen) > rl.window {
			c.count = 1
			c.lastSeen = time.Now()

			next(w, r)
			return
		}
		if c.count >= rl.limit {

			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		c.count++
		c.lastSeen = time.Now()
		next(w, r)
	}
}
