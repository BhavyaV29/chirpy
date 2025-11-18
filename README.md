
# Chirpy

![Go Version](https://img.shields.io/badge/Go-1.23.5-blue)
![Build](https://img.shields.io/badge/build-passing-brightgreen)

## 🐦 What is Chirpy?

Chirpy is a lightweight, extensible microblogging platform written in Go. It allows users to post short messages ("chirps"), manage accounts, and interact securely with a PostgreSQL backend. Chirpy is designed for learning, experimentation, and as a foundation for building more complex social applications.

---

## ✨ Key Features

- User registration and authentication (JWT-based)
- Secure password hashing (Argon2id)
- Post, fetch, and manage chirps (short messages)
- Refresh token support for session management
- Admin endpoints for platform management
- Environment-based configuration (dotenv)
- Modular, testable Go codebase
- SQL migrations and code generation (sqlc)
- RESTful API with clear separation of concerns

---

## 🏗️ Design Overview

Chirpy is structured for clarity and extensibility:

- **main.go**: Entry point, HTTP server, routing, and handler registration.
- **internal/database/**: SQLc-generated database access layer, models, and query logic.
- **internal/auth/**: Authentication, JWT, password hashing, and token utilities.
- **sql/schema/**: SQL migration scripts for PostgreSQL schema.
- **sql/queries/**: SQLc query definitions for CRUD operations.
- **assets/**: Static assets (e.g., logo).

### Architecture

- **API-first**: All core features are exposed via a RESTful API.
- **Middleware**: Metrics, authentication, and admin logic are handled via middleware and dedicated handlers.
- **Environment-driven**: Configuration is loaded from environment variables (dotenv supported).
- **Security**: Passwords are hashed with Argon2id; JWTs are used for stateless authentication.

---

## 📚 Documentation & Technical Reference

### Project Structure

- [`main.go`](main.go): HTTP server, API endpoints, and handler logic.
- [`internal/database/`](internal/database/):
  - `db.go`, `models.go`: Database models and connection logic.
  - `*.sql.go`: SQLc-generated CRUD/query code for users, chirps, tokens, etc.
- [`internal/auth/`](internal/auth/):
  - `auth.go`: Password hashing, JWT creation/validation, API key and bearer token parsing, refresh token generation.
- [`sql/schema/`](sql/schema/):
  - `001_users.sql`, `002_chirps.sql`, etc.: PostgreSQL schema migrations for users, chirps, tokens, and premium features.
- [`sql/queries/`](sql/queries/):
  - Query definitions for SQLc code generation.
- [`assets/`](assets/): Static files (e.g., `logo.png`).
- [`.github/copilot-instructions.md`](.github/copilot-instructions.md): Project documentation and contribution instructions.

### API Endpoints

#### User Management
- `POST /api/users` — Register a new user
- `PUT /api/users` — Update user email/password (JWT required)
- `POST /api/login` — Authenticate and receive JWT/refresh token
- `POST /api/refresh` — Exchange refresh token for new JWT
- `POST /api/revoke` — Revoke refresh token

#### Chirps
- `POST /api/chirps` — Create a new chirp (JWT required)
- `GET /api/chirps/` — List all chirps (optionally filter/sort)
- `GET /api/chirps/{chirpID}` — Get a single chirp by ID
- `DELETE /api/chirps/{chirpID}` — Delete a chirp (JWT required, owner only)

#### Admin & Premium
- `GET /admin/metrics` — View server metrics
- `POST /admin/reset` — Reset metrics and (in dev) delete users
- `POST /api/polka/webhooks` — Webhook for "Chirpy Red" premium upgrades

### Authentication & Security

- Passwords are hashed using Argon2id before storage.
- JWTs are used for stateless authentication; refresh tokens for session renewal.
- API keys are used for webhook authentication.

---

## 🚀 Getting Started

### Prerequisites
- Go 1.23.5 or later
- PostgreSQL database

### Installation

1. **Clone the repository:**
   ```bash
   git clone https://github.com/BhavyaV29/chirpy.git
   cd chirpy
   ```
2. **Set up environment variables:**
   - Copy `.env.example` to `.env` and fill in your secrets (DB connection, JWT secret, etc).
3. **Install dependencies:**
   ```bash
   go mod tidy
   ```
4. **Set up the database:**
   - Run the SQL scripts in `sql/schema/` to create tables.
   - Use `sql/queries/` for query reference.
5. **Run the server:**
   ```bash
   go run main.go
   ```

### Usage Example

- Register a user:
  ```http
  POST /api/users
  { "email": "user@example.com", "password": "yourpassword" }
  ```
- Post a chirp:
  ```http
  POST /api/chirps
  { "body": "Hello, Chirpy!" }
  ```

---

## 🧩 Extending & Customizing

- Add new endpoints by creating handlers in `main.go` and queries in `internal/database/`.
- Update the database schema via new migration scripts in `sql/schema/`.
- Use the modular structure to add features (e.g., likes, comments, notifications).

---

## 🤝 Contributing

Contributions are welcome! Please fork the repo, create a branch, and submit a pull request. For major changes, open an issue first to discuss what you’d like to change.

- See `.github/` for guidelines and instructions

---

## 👤 Maintainer

- **Bhavya V**  
  [GitHub Profile](https://github.com/BhavyaV29)

---

## 📝 License

This project is open source and available under the MIT License.

- See [internal/database/](internal/database/) for database models and queries
- See [internal/auth/](internal/auth/) for authentication logic
- [sqlc.yaml](sqlc.yaml) for SQL codegen config
- For issues, open a GitHub issue or contact the maintainer

## 🤝 Contributing

Contributions are welcome! Please fork the repo, create a branch, and submit a pull request. For major changes, open an issue first to discuss what you’d like to change.

- See `.github/` for guidelines and instructions

## 👤 Maintainer

- **Bhavya V**  
  [GitHub Profile](https://github.com/BhavyaV29)

## 📝 License

This project is open source and available under the MIT License.
