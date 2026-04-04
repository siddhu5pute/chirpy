# Chirpy 🐦

![Chirpy Logo](assets/logo.png)

A RESTful JSON API backend built in Go — a Twitter-like microblogging platform where users can post short messages called "chirps", follow an authenticated session system, and optionally upgrade to **Chirpy Red** membership.

---

## What Is Chirpy?

Chirpy is a fully featured backend API that handles:
- User registration and login with hashed passwords
- JWT-based authentication with access and refresh tokens
- Creating, reading, and deleting chirps (max 140 characters)
- Authorization — users can only delete their own chirps
- Chirpy Red membership upgrades via Polka payment webhooks
- Profanity filtering on chirp content

---

## Tech Stack

- **Language:** Go
- **Database:** PostgreSQL
- **Query generation:** [sqlc](https://sqlc.dev/)
- **Migrations:** [Goose](https://github.com/pressly/goose)
- **Password hashing:** [Argon2id](https://github.com/alexedwards/argon2id)
- **JWT:** [golang-jwt/jwt](https://github.com/golang-jwt/jwt)

---

## Prerequisites

- Go 1.22+
- PostgreSQL running locally
- `goose` installed: `go install github.com/pressly/goose/v3/cmd/goose@latest`
- `sqlc` installed: `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`

---

## Installation

```bash
# 1. Clone the repository
git clone https://github.com/siddhu5pute/chirpy.git
cd chirpy

# 2. Install dependencies
go mod download

# 3. Create a .env file in the project root
touch .env
```

Add the following to your `.env` file:

```
DB_URL=postgres://your_username:your_password@localhost:5432/chirpy?sslmode=disable
PLATFORM=dev
JWT_SECRET=your_generated_secret_here
POLKA_KEY=your_polka_api_key_here
```

Generate a secure JWT secret with:
```bash
openssl rand -base64 64
```

---

## Database Setup

```bash
# Create the database
createdb chirpy

# Run all migrations
goose -dir sql/schema postgres "postgres://your_username:@localhost:5432/chirpy" up
```

---

## Running the Server

```bash
go build -o out && ./out
```

Server starts on `http://localhost:8080`

---

## API Endpoints

### Health Check
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/healthz` | None | Returns OK if server is running |

---

### Users
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/api/users` | None | Create a new user |
| PUT | `/api/users` | Bearer Token | Update email and password |

**Create User — Request Body:**
```json
{
  "email": "user@example.com",
  "password": "yourpassword"
}
```

**Create User — Response (201):**
```json
{
  "id": "f0f87ec2-a8b5-48cc-b66a-a85ce7c7b862",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "email": "user@example.com",
  "is_chirpy_red": false
}
```

---

### Authentication
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/api/login` | None | Login and receive tokens |
| POST | `/api/refresh` | Refresh Token | Get a new access token |
| POST | `/api/revoke` | Refresh Token | Revoke a refresh token (logout) |

**Login — Request Body:**
```json
{
  "email": "user@example.com",
  "password": "yourpassword"
}
```

**Login — Response (200):**
```json
{
  "id": "f0f87ec2-a8b5-48cc-b66a-a85ce7c7b862",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "email": "user@example.com",
  "is_chirpy_red": false,
  "token": "eyJhbGci...",
  "refresh_token": "56aa826d..."
}
```

**Token Usage:**
```
Authorization: Bearer YOUR_ACCESS_TOKEN
```

- Access tokens expire after **1 hour**
- Refresh tokens expire after **60 days**
- Use `POST /api/refresh` with your refresh token to get a new access token
- Use `POST /api/revoke` to logout and invalidate your refresh token

---

### Chirps
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/api/chirps` | Bearer Token | Create a chirp |
| GET | `/api/chirps` | None | Get all chirps |
| GET | `/api/chirps/{chirpID}` | None | Get a single chirp |
| DELETE | `/api/chirps/{chirpID}` | Bearer Token | Delete your own chirp |

**Create Chirp — Request Body:**
```json
{
  "body": "Hello world! This is my first chirp."
}
```

**Get Chirps — Query Parameters:**
| Parameter | Values | Description |
|-----------|--------|-------------|
| `author_id` | UUID | Filter chirps by a specific user |
| `sort` | `asc` / `desc` | Sort by `created_at` (default: `asc`) |

**Examples:**
```
GET /api/chirps
GET /api/chirps?sort=desc
GET /api/chirps?author_id=f0f87ec2-a8b5-48cc-b66a-a85ce7c7b862
GET /api/chirps?author_id=f0f87ec2-a8b5-48cc-b66a-a85ce7c7b862&sort=desc
```

**Notes:**
- Chirps are limited to **140 characters**
- The words `kerfuffle`, `sharbert`, and `fornax` are automatically replaced with `****`
- Only the author of a chirp can delete it — returns `403` otherwise

---

### Chirpy Red (Webhooks)
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/api/polka/webhooks` | ApiKey | Upgrade user to Chirpy Red |

**Request Header:**
```
Authorization: ApiKey YOUR_POLKA_KEY
```

**Request Body:**
```json
{
  "event": "user.upgraded",
  "data": {
    "user_id": "f0f87ec2-a8b5-48cc-b66a-a85ce7c7b862"
  }
}
```

Only `user.upgraded` events are processed. All other events return `204` immediately.

---

### Admin (Dev Only)
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/admin/metrics` | None | View total server hits |
| POST | `/admin/reset` | None | Reset hit counter and delete all users |

> ⚠️ Admin reset only works when `PLATFORM=dev` is set in your `.env`

---

## Project Structure

```
chirpy/
├── main.go                  # Server setup, routes, apiConfig
├── handler.go               # All HTTP handlers
├── internal/
│   ├── auth/
│   │   ├── auth.go          # JWT, password hashing, token functions
│   │   └── jwt_test.go      # Unit tests for auth package
│   └── database/            # sqlc generated database code
├── sql/
│   ├── schema/              # Goose migration files
│   └── queries/             # sqlc SQL query files
└── sqlc.yaml                # sqlc configuration
```

---

## Authentication Flow

```
1. POST /api/users       → Create account
2. POST /api/login       → Get access token (1hr) + refresh token (60 days)
3. Use access token      → Authorization: Bearer <token> on protected endpoints
4. POST /api/refresh     → Get new access token when it expires
5. POST /api/revoke      → Logout — invalidates refresh token
```

---

## Running Tests

```bash
go test ./...
```

---

## Environment Variables

| Variable | Description |
|----------|-------------|
| `DB_URL` | PostgreSQL connection string |
| `PLATFORM` | Set to `dev` to enable admin reset endpoint |
| `JWT_SECRET` | Secret key for signing JWTs — keep this private |
| `POLKA_KEY` | API key provided by Polka payment webhook service |