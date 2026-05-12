package handler

import (
	"backend/model"
	"backend/service"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
)

type TaskHandler struct {
	Service *service.TaskService
}
type AuthHandler struct {
	Service  *service.AuthService
	Validate *validator.Validate
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	json.NewDecoder(r.Body).Decode(&req)

	err := h.Validate.Struct(req)
	if err != nil {
		var errMessages []string
		for _, e := range err.(validator.ValidationErrors) {
			errMessages = append(errMessages, e.Error())
		}
		http.Error(w, strings.Join(errMessages, ", "), http.StatusBadRequest)
		return
	}

	user, err := h.Service.Register(req.Username, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	json.NewDecoder(r.Body).Decode(&req)
	token, err := h.Service.Login(req.Username, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	json.NewEncoder(w).Encode(model.AuthResponse{Token: token})
}

func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)
	tasks, err := h.Service.GetAll(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(tasks)
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)
	var task model.Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	task, err = h.Service.Create(userID, task.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)
	idStr := strings.TrimPrefix(r.URL.Path, "/tasks/")
	if idStr == "" {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}
	err = h.Service.Delete(userID, id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	// extract id safely
	userID := r.Context().Value("user_id").(int)
	idStr := strings.TrimPrefix(r.URL.Path, "/tasks/")
	if idStr == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	// decode body
	var task model.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	updated, err := h.Service.Update(userID, id, task.Name, task.Done)
	if err != nil {
		// IMPORTANT: this should NOT always be 500
		http.Error(w, "task not found or update failed", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}
