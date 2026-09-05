# Blog API

A small blogging REST API written in Go, using [Gin](https://github.com/gin-gonic/gin), SQLite, and JWT-based authentication.

## Features

- User registration & login with bcrypt-hashed passwords
- JWT authentication (7-day expiry) protecting all blog routes
- Blog CRUD, scoped to the authenticated user (you can only update/delete your own blogs)
- Pagination on blog listing endpoints
- Config via environment variables / `.env` (see below)
- Graceful shutdown on `SIGINT`/`SIGTERM`

## Requirements

- Go 1.26+
- SQLite (bundled via `mattn/go-sqlite3`, no separate install needed)

## Setup

```bash
git clone git@github.com:Kyaw-Min-Khant/golang-blog-api.git
cd golang-blog-api
cp .env.example .env
```

Edit `.env` and set a real `JWT_SECRET` — generate one with:

```bash
openssl rand -base64 32
```

Run it:

```bash
go run .
```

Or with live-reload during development ([air](https://github.com/air-verse/air)):

```bash
air
```

Run the tests:

```bash
go test ./...
```

## Configuration

All config is read from environment variables (`.env` is loaded automatically if present; it's a no-op if it's missing, which is the case in most production deployments where the platform injects env vars directly).

| Variable      | Default                          | Notes                                                                 |
|---------------|-----------------------------------|------------------------------------------------------------------------|
| `APP_ENV`     | `development`                    | Set to `production` to enable release mode. Also enforces that `JWT_SECRET` must be set (the server refuses to start without it). |
| `PORT`        | `8080`                           | HTTP port to listen on.                                                |
| `DB_PATH`     | `blog.db`                        | Path to the SQLite database file. Use a different value per environment — see [Deployment](#deployment). |
| `JWT_SECRET`  | *(insecure dev default)*         | Secret used to sign JWTs. **Must** be set explicitly outside local dev. |

## API Reference

Base path: `v1/api`

All routes except `health-check`, `auth/register`, and `auth/login` require an `Authorization: Bearer <token>` header.

| Method | Path              | Auth | Description                          |
|--------|-------------------|------|---------------------------------------|
| GET    | `/health-check`   | No   | Liveness check                        |
| POST   | `/auth/register`  | No   | Create a user account                 |
| POST   | `/auth/login`     | No   | Log in, returns a JWT                 |
| GET    | `/profile`        | Yes  | Current authenticated user's identity |
| POST   | `/blog`           | Yes  | Create a blog post                    |
| GET    | `/blog`           | Yes  | List all blogs (paginated)            |
| GET    | `/blog/user`      | Yes  | List the current user's own blogs (paginated) |
| PATCH  | `/blog/:blog_id`  | Yes  | Update a blog you own                 |
| DELETE | `/blog/:blog_id`  | Yes  | Delete a blog you own                 |

Pagination query params (on `GET /blog` and `GET /blog/user`): `limit` (default `10`, max `100`) and `page` (offset, default `0`).

### Response shape

Every response follows the same envelope:

```json
{ "status": "success", "data": { ... } }
```

```json
{ "status": "error", "message": "..." }
```

### Example: register, login, create a blog

```bash
curl -X POST http://localhost:8080/v1/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"secretpw"}'

TOKEN=$(curl -s -X POST http://localhost:8080/v1/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"secretpw"}' | jq -r .data.token)

curl -X POST http://localhost:8080/v1/api/blog \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Hello","content":"World"}'
```

## Project Structure

| File              | Responsibility                                      |
|-------------------|------------------------------------------------------|
| `main.go`         | Wiring: config, routes, HTTP server, graceful shutdown |
| `config.go`       | Env var / `.env` loading                              |
| `db.go`           | SQLite connection + schema migration                  |
| `models.go`       | Request/response structs                              |
| `store.go`        | Database queries                                      |
| `handlers.go`     | HTTP handlers                                          |
| `auth.go`         | Password hashing, JWT generation/parsing               |
| `middleware.go`   | JWT auth middleware                                    |
| `response.go`     | Uniform success/error JSON helpers                     |

## Deployment

Deployed on [Render](https://render.com) as a native Go web service:

- **Build Command**: `go build -o app .`
- **Start Command**: `./app`
- **Root Directory**: leave blank (this repo *is* the service root)
- **Environment**: set `APP_ENV=production` and `JWT_SECRET` under the service's Environment tab (`PORT` is provided automatically by Render)
- **Persistent storage**: SQLite needs a real file on disk that survives redeploys. Attach a [Render Disk](https://render.com/docs/disks) and point `DB_PATH` at a path inside it — without this, the database resets on every deploy.

## Known limitations

- SQLite is single-writer; if this ever needs to run as more than one instance/process, move to a real DB server (Postgres/MySQL) instead of scaling SQLite.
