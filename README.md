# Poll & Go

A small vibe coded poll app created for infrastructure and integration testing.

- Anonymous single-choice polling app 
- A visitor receives a random voter ID in browser local storage
- A visitor can create a poll, share its short code, and cast one vote per poll
- The home page keeps a server-backed list of polls that the current browser created or voted in.

![Poll & Go application preview](docs/app_screenshot.png)

## Stack

- `frontend`: React, TypeScript, and Vite
- `backend`: Go HTTP API using PostgreSQL through pgx
- PostgreSQL 17 for local development

## Run locally

Running the complete containerized stack requires Docker with Docker Compose. The hot-reload development workflow additionally requires Go 1.24+, Node.js 22+, and npm.

The quickest way to run the whole stack is:

```sh
docker compose up --build
```

Then open `http://localhost:5173`.

For development with hot reload, start PostgreSQL and the API. `DATABASE_URL` is required on every platform. In shells such as Bash or Zsh, it can be set for the single server command:

```sh
docker compose up -d database
cd backend
go mod download
DATABASE_URL='postgres://pollandgo:pollandgo@localhost:5432/pollandgo?sslmode=disable' go run ./cmd/server
```

Then start the frontend in another terminal:

```sh
cd frontend
npm install
npm run dev
```

Open `http://localhost:5173`. Vite proxies `/api` to the backend during development. For separate production hosts, build the frontend with `VITE_API_URL` set to the public backend URL and set the backend's comma-separated `ALLOWED_ORIGINS` value to the frontend origin.

The backend applies its idempotent schema migration on startup. Its health endpoint is `GET /healthz`.

## API documentation

The HTTP endpoints and request formats are documented in [docs/API.md](docs/API.md).
