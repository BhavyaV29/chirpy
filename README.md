# Chirpy

[![Go](https://img.shields.io/badge/Go-1.23.5-00ADD8?style=flat&logo=go&logoColor=white)](https://golang.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-316192?style=flat&logo=postgresql&logoColor=white)](https://postgresql.org)
[![SQLC](https://img.shields.io/badge/SQLC-type--safe%20queries-4479A1?style=flat)](https://sqlc.dev)
[![Goose](https://img.shields.io/badge/migrations-Goose-F58220?style=flat)](https://github.com/pressly/goose)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**Production-grade micro-blogging REST API in Go — JWT auth, Argon2id, SQLC data layer, and Polka webhooks for premium upgrades.**

---

## What it does

Chirpy is a micro-blogging backend where users register, authenticate, and post short messages ("chirps"). It implements stateless JWT access tokens with refresh token rotation, Argon2id-hashed passwords, and a full CRUD layer over PostgreSQL generated entirely by SQLC. A Polka webhook endpoint allows an external payment provider to toggle premium ("Chirpy Red") membership on a user account — no manual database edits required.

---

## Key Features

- **JWT auth with refresh token rotation** — 1h access tokens + long-lived refresh tokens stored and revocable in PostgreSQL
- **Argon2id password hashing** — credentials never stored in plaintext; hash/verify on every login
- **SQLC type-safe query layer** — SQL queries compiled to idiomatic Go at build time; no ORM, no `interface{}`
- **Goose schema migrations** — versioned `sql/schema/` files; reproducible DB setup with one command
- **Polka webhook** — `POST /api/polka/webhooks` accepts external events to upgrade users to Chirpy Red; validated by API key
- **Admin tooling** — `GET /admin/metrics` for live request counts; `POST /admin/reset` for dev teardown

---

## How it works

All routing is handled by Go's standard `net/http`. Auth logic lives in `internal/auth` (JWT sign/verify, Argon2id, API key parsing); the database access layer lives in `internal/database` (SQLC-generated from `sql/queries/`). Each handler is a method on `apiConfig`, which carries the DB connection pool and secrets loaded from environment variables. Migrations in `sql/schema/` set up the schema; SQLC regenerates the Go layer whenever queries change.

---

## Quick Start

```bash
git clone https://github.com/BhavyaV29/chirpy.git
cd chirpy
# create a .env with DB_URL, PLATFORM, JWT_SECRET, POLKA_KEY

go mod tidy

# apply migrations (requires goose + a running Postgres instance)
goose -dir sql/schema postgres "$DB_URL" up

go run main.go              # server starts on :8080
```

---

## API Reference

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/users` | — | Register |
| PUT | `/api/users` | JWT | Update email / password |
| POST | `/api/login` | — | Get JWT + refresh token |
| POST | `/api/refresh` | Refresh token | Issue new JWT |
| POST | `/api/revoke` | Refresh token | Revoke session |
| POST | `/api/chirps` | JWT | Create chirp |
| GET | `/api/chirps` | — | List chirps (filterable, sortable) |
| GET | `/api/chirps/{id}` | — | Get single chirp |
| DELETE | `/api/chirps/{id}` | JWT (owner) | Delete chirp |
| POST | `/api/polka/webhooks` | API key | Premium upgrade event |
| GET | `/admin/metrics` | — | Request counter |
| POST | `/admin/reset` | — | Dev teardown |

---

## Project Structure

```
chirpy/
├── main.go                  # server setup, routing, handler registration
├── internal/
│   ├── auth/                # JWT sign/verify, Argon2id, API key helpers
│   └── database/            # SQLC-generated models and queries
├── sql/
│   ├── schema/              # Goose migration files (001_users.sql, …)
│   └── queries/             # SQL definitions read by SQLC
└── sqlc.yaml                # SQLC codegen config
```

---

## Maintainer

[BhavyaV29](https://github.com/BhavyaV29) · [bhavyaportfolio.site](https://bhavyaportfolio.site)

Licensed under the MIT License.
