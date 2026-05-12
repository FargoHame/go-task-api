# go-task-api

A production-ready REST API built in Go — JWT authentication, rate limiting, per-user data isolation, and load tested to 200 concurrent users.

## Tech Stack

- **Go** — net/http standard library
- **PostgreSQL** — primary database
- **JWT** — authentication via golang-jwt
- **bcrypt** — password hashing
- **go-playground/validator** — input validation

## Architecture

Layered architecture with strict separation of concerns:

```
router → middleware → handler → service → repository → postgres
```

- **Router** — route registration and method dispatch
- **Middleware** — JWT auth, rate limiting, request logging
- **Handler** — request parsing and response encoding
- **Service** — business logic and validation
- **Repository** — all database interactions

## Features

- JWT authentication with 24hr expiry
- Per-user data isolation — users can only access their own tasks
- Rate limiting — 100 requests per minute per IP (configurable)
- Request logging — method, path, status code, duration
- Input validation on all endpoints
- Custom error types with proper HTTP status codes
- IDOR protection on all write operations

## Performance

Load tested with k6 to 200 concurrent users:

| Metric | Before Index | After Index |
|--------|-------------|-------------|
| avg response time | 1.01s | 142ms |
| p(90) | 3.12s | 385ms |
| p(95) | 3.52s | 445ms |
| requests/sec | 59 | 136 |
| failure rate | 0% | 0% |


## Getting Started

### Prerequisites

- Go 1.21+
- PostgreSQL

### Setup

```bash
git clone https://github.com/fargohame/go-task-api.git
cd go-task-api
go mod tidy
```

### Database

Create the database and tables:

```sql
CREATE DATABASE taskdb;

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL
);

CREATE TABLE tasks (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    done BOOLEAN DEFAULT FALSE
);

CREATE INDEX idx_tasks_user_id ON tasks (user_id);
```

### Environment Variables

Create a `.env` file in the project root:

```
JWT_SECRET=your_secret_key_here
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=taskdb
```

### Run

```bash
go run main.go
```

Server starts at `http://localhost:8080`

## API Endpoints

### Auth

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | /register | Register a new user | No |
| POST | /login | Login and receive JWT token | No |

### Tasks

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | /tasks | Get all tasks for authenticated user | Yes |
| POST | /tasks | Create a new task | Yes |
| PUT | /tasks/{id} | Update a task | Yes |
| DELETE | /tasks/{id} | Delete a task | Yes |

### Authentication

Include the JWT token in the Authorization header:

```
Authorization: Bearer <token>
```

### Example Requests

**Register:**
```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"username": "hamza", "password": "securepass"}'
```

**Login:**
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username": "hamza", "password": "securepass"}'
```

**Create Task:**
```bash
curl -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"name": "learn go"}'
```

**Get Tasks:**
```bash
curl http://localhost:8080/tasks \
  -H "Authorization: Bearer <token>"
```

## Testing

```bash
go test -v
```

8 tests covering auth, CRUD operations, unauthorized access, and rate limiting.

## Load Testing

Install [k6](https://k6.io) and run:

```bash
k6 run loadtest.js
```
