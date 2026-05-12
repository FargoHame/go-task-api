package service

import (
	"backend/model"
	"backend/repository"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type TaskService struct {
	Repo      *repository.TaskRepository
	AsyncChan chan model.Task
}
type AuthService struct {
	Repo   *repository.TaskRepository
	Secret string
}

func (s *AuthService) Register(username, password string) (model.User, error) {
	if username == "" || password == "" {
		return model.User{}, errors.New("username and password cannot be empty")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return model.User{}, errors.New("error hashing password")
	}
	return s.Repo.CreateUser(username, string(hashed))
}
func (s *AuthService) Login(username, password string) (string, error) {
	user, err := s.Repo.GetUserByUsername(username)
	if err != nil {
		return "", errors.New("invalid username or password")
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", errors.New("invalid username or password")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte(s.Secret))
	return tokenString, err
}
func (s *TaskService) Create(userID int, name string) (model.Task, error) {
	if name == "" {
		return model.Task{}, errors.New("task name cannot be empty")
	}
	task, err := s.Repo.Create(userID, name)
	if err != nil {
		return model.Task{}, err
	}
	go func(t model.Task) {
		if s.AsyncChan != nil {
			s.AsyncChan <- t
		}
	}(task)
	return task, nil
}

func (s *TaskService) GetAll(userID int) ([]model.Task, error) {
	return s.Repo.GetAll(userID)
}

func (s *TaskService) Delete(userID, id int) error {
	return s.Repo.Delete(userID, id)
}

func (s *TaskService) Update(userID, id int, name string, done bool) (model.Task, error) {
	if name == "" {
		return model.Task{}, errors.New("task name cannot be empty")
	}
	return s.Repo.Update(userID, id, name, done)
}
