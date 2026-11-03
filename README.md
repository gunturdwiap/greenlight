# Greenlight

A REST API for managing movies.

## Requirements

- Go 1.26.1+
- Docker and Docker Compose
- [golang-migrate](https://github.com/golang-migrate/migrate)

---

## Setup Guide

### 1. Create the `.env` file

Copy the example environment file:

```bash
cp .env.example .env

```

---

### 2. Start the Database and Mail Server

Run this command in your terminal to start PostgreSQL and Mailpit using Docker:

```bash
docker compose up -d

```

*(This runs PostgreSQL in the background and starts Mailpit so you can view sent emails at `http://localhost:8025`).*

---

### 3. Run Database Migrations

Apply the database structure using your Makefile:

```bash
make db/migrations/up

```

*(Press `y` and hit Enter when it asks for confirmation).*

---

### 4. Run the API

Start the application server:

```bash
make run/api

```

The server is now running at `http://localhost:4000`.

---

## Quick Test

Open a new terminal window and run these commands to test it:

1. **Check health:**
```bash
curl localhost:4000/v1/healthcheck

```


2. **Register a user:**
```bash
curl -X POST localhost:4000/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice", "email": "alice@example.com", "password": "pa5sw0rd"}'

```



---

## API Endpoints

| Method | URL Pattern | Action | Required Permission |
| --- | --- | --- | --- |
| **GET** | `/v1/healthcheck` | Show application health and version info | - |
| **GET** | `/v1/movies` | Show all movies | `movies:read` |
| **POST** | `/v1/movies` | Create a new movie | `movies:write` |
| **GET** | `/v1/movies/:id` | Show a specific movie | `movies:read` |
| **PATCH** | `/v1/movies/:id` | Update a specific movie | `movies:write` |
| **DELETE** | `/v1/movies/:id` | Delete a specific movie | `movies:write` |
| **POST** | `/v1/users` | Register a new user | - |
| **PUT** | `/v1/users/activated` | Activate a user | - |
| **PUT** | `/v1/users/password` | Update a user password | - |
| **POST** | `/v1/tokens/authentication` | Generate authentication token | - |
| **POST** | `/v1/tokens/password-reset` | Generate password reset token | - |
| **GET** | `/debug/vars` | Show app metrics | - |

