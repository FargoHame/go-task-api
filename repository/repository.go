package repository

import (
	"backend/model"
	"database/sql"
	"errors"
)

type TaskRepository struct {
	DB *sql.DB
}

func (r *TaskRepository) Create(userID int, name string) (model.Task, error) {
	var task model.Task
	err := r.DB.QueryRow(
		"INSERT INTO tasks (user_id, name) VALUES ($1, $2) RETURNING id, user_id, name, done",
		userID, name,
	).Scan(&task.ID, &task.UserID, &task.Name, &task.Done)
	return task, err
}

func (r *TaskRepository) GetAll(userID int) ([]model.Task, error) {
	rows, err := r.DB.Query("SELECT id, user_id, name, done FROM tasks WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []model.Task

	for rows.Next() {
		var t model.Task
		if err := rows.Scan(&t.ID, &t.UserID, &t.Name, &t.Done); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (r *TaskRepository) Delete(userID, id int) error {
	result, err := r.DB.Exec("DELETE FROM tasks WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (r *TaskRepository) Update(userID, id int, name string, done bool) (model.Task, error) {
	var task model.Task
	err := r.DB.QueryRow(
		"UPDATE tasks SET name = $1, done = $2 WHERE id = $3 AND user_id = $4 RETURNING id, user_id, name, done",
		name, done, id, userID,
	).Scan(&task.ID, &task.UserID, &task.Name, &task.Done)
	if err != nil {
		return model.Task{}, errors.New("task not found or unauthorized")
	}
	return task, nil
}
func (r *TaskRepository) CreateUser(username, hashedPassword string) (model.User, error) {

	var user model.User
	err := r.DB.QueryRow(
		"INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id, username",
		username, hashedPassword,
	).Scan(&user.ID, &user.Username)
	return user, err
}

func (r *TaskRepository) GetUserByUsername(username string) (model.User, error) {
	var user model.User
	err := r.DB.QueryRow(
		"SELECT id, username, password FROM users WHERE username = $1",
		username,
	).Scan(&user.ID, &user.Username, &user.Password)
	return user, err
}
