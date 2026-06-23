# 💰 Finance Tracker API

A RESTful backend API for personal finance tracking built with **Go**, **Gin**, **GORM**, and **PostgreSQL**. Supports user authentication via JWT and full transaction management with an income/expense dashboard.

---

## 🧰 Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.26+ |
| Framework | Gin v1.12 |
| ORM | GORM v1.31 |
| Database | PostgreSQL (Neon / local) |
| Auth | JWT (golang-jwt/jwt v5) |
| Password | bcrypt |
| Config | godotenv |

---

## 📁 Project Structure

```
finance-tracker/
├── config/
│   └── database.go          # PostgreSQL connection
├── controllers/
│   ├── auth_controller.go   # Register & Login
│   └── transaction_controller.go  # Transaction CRUD & Dashboard
├── middleware/
│   └── jwt.go               # JWT auth middleware
├── models/
│   ├── user.go              # User model
│   └── transaction.go       # Transaction model
├── routes/
│   └── routes.go            # Route definitions
├── utils/
│   ├── hash.go              # bcrypt helpers
│   └── jwt.go               # Token generation
├── .env                     # Environment variables (never commit this)
├── go.mod
├── go.sum
└── main.go
```

---

## ⚙️ Prerequisites

- [Go 1.21+](https://go.dev/dl/)
- A PostgreSQL database — local or cloud (e.g. [Neon](https://neon.tech))
- Git

---

## 🐘 PostgreSQL Setup

### Option A — Local PostgreSQL

1. Install PostgreSQL from [postgresql.org](https://www.postgresql.org/download/)
2. Create a database:

```sql
CREATE DATABASE finance_tracker;
```

3. Note your credentials: host, user, password, port (default `5432`).

> GORM auto-migrates the `users` and `transactions` tables on first run — no SQL scripts needed.

### Option B — Neon (Free Cloud PostgreSQL)

1. Sign up at [neon.tech](https://neon.tech)
2. Create a new project and database
3. Go to **Dashboard → Connection Details**
4. Copy the connection string — it looks like:

```
postgresql://user:password@ep-xxxx.neon.tech/dbname?sslmode=require
```

5. Split it into individual env vars for your `.env` file (see below)

---

## 🔧 Local Setup

### 1. Clone the repository

```bash
git clone https://github.com/mahesaPinuntun/REST-API-GIN-Golang-FinanceTrackerBackEndPart.git
cd REST-API-GIN-Golang-FinanceTrackerBackEndPart
```

### 2. Install dependencies

```bash
go mod tidy
```

### 3. Create your `.env` file

Create a file named `.env` in the project root:

```env
# Database
DB_HOST=localhost
DB_USER=postgres
DB_PASS=your_password
DB_NAME=finance_tracker
DB_PORT=5432

# JWT
JWT_SECRET=your_secret_key_here
```

> **Neon users:** parse your connection string into individual fields.
> For `ep-floral-glitter-xxx-pooler.c-9.us-east-1.aws.neon.tech`, set `DB_HOST` to that hostname, `DB_USER`/`DB_PASS`/`DB_NAME` from the URL, and `DB_PORT=5432`.
> Also add `DB_SSLMODE=require` and update `database.go` accordingly if using Neon.

### 4. Run the server

```bash
go run main.go
```

Server starts at `http://localhost:8080`

---

## 🚀 Deployment on Railway

[Railway](https://railway.app) is the recommended free hosting platform for Go/Gin projects.

### 1. Prepare your code

Make sure `main.go` reads the port dynamically (required by Railway):

```go
port := os.Getenv("PORT")
if port == "" {
    port = "8080"
}
r.Run(":" + port)
```

Also make `.env` loading optional (Railway uses real env vars, not a file):

```go
godotenv.Load(".env") // no log.Fatal — .env is optional in production
```

### 2. Deploy

1. Push your code to GitHub
2. Go to [railway.app](https://railway.app) → **New Project** → **Deploy from GitHub**
3. Select your repository
4. Railway auto-detects Go from `go.mod` and builds it

### 3. Set environment variables

In Railway dashboard → your service → **Variables**, add:

```
DB_HOST        = your_neon_host
DB_USER        = your_neon_user
DB_PASS        = your_neon_password
DB_NAME        = your_neon_dbname
DB_PORT        = 5432
JWT_SECRET     = your_secret_key
```

Railway automatically injects `PORT` — no need to set that one yourself.

### 4. Done

Railway provides a public URL like `https://your-app.up.railway.app`.

---

## 📡 API Reference

### Base URL

```
Local:      http://localhost:8080
Production: https://your-app.up.railway.app
```

---

### 🔓 Public Routes

#### `POST /register` — Create a new user

**Request Body:**
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "yourpassword"
}
```

**Response `201`:**
```json
{
  "message": "User created"
}
```

---

#### `POST /login` — Authenticate and get a JWT token

**Request Body:**
```json
{
  "email": "john@example.com",
  "password": "yourpassword"
}
```

**Response `200`:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Error responses:**
- `401 User not found` — email doesn't exist
- `401 Invalid credential` — wrong password

---

### 🔐 Protected Routes

All routes below require a JWT token in the `Authorization` header:

```
Authorization: Bearer <your_token>
```

---

#### `POST /api/transactions` — Create a transaction

**Request Body:**
```json
{
  "user_id": 1,
  "title": "Freelance Payment",
  "amount": 500000,
  "type": "income",
  "category": "Freelance",
  "description": "Website project payment"
}
```

> `type` must be either `"income"` or `"expense"`

**Response `201`:** Returns the created transaction object.

---

#### `GET /api/transactions` — Get all transactions

**Response `200`:**
```json
[
  {
    "id": 1,
    "user_id": 1,
    "title": "Freelance Payment",
    "amount": 500000,
    "type": "income",
    "category": "Freelance",
    "description": "Website project payment",
    "created_at": "2026-06-20T18:37:54Z",
    "updated_at": "2026-06-20T18:37:54Z"
  }
]
```

---

#### `GET /api/dashboard` — Get income/expense summary

**Response `200`:**
```json
{
  "income": 1500000,
  "expense": 300000,
  "balance": 1200000
}
```

---

### PowerShell Examples (Windows)

**Register:**
```powershell
$body = @{ name="John"; email="john@example.com"; password="pass123" } | ConvertTo-Json
Invoke-RestMethod -Uri "http://localhost:8080/register" -Method POST -ContentType "application/json" -Body $body
```

**Login and save token:**
```powershell
$body = @{ email="john@example.com"; password="pass123" } | ConvertTo-Json
$response = Invoke-RestMethod -Uri "http://localhost:8080/login" -Method POST -ContentType "application/json" -Body $body
$token = $response.token
```

**Create transaction:**
```powershell
$body = @{ user_id=1; title="Salary"; amount=5000000; type="income"; category="Work"; description="Monthly salary" } | ConvertTo-Json
Invoke-RestMethod -Uri "http://localhost:8080/api/transactions" -Method POST -ContentType "application/json" -Headers @{ Authorization="Bearer $token" } -Body $body
```

**Get dashboard:**
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/dashboard" -Method GET -Headers @{ Authorization="Bearer $token" }
```

---

## 🔒 Security Notes

- Passwords are hashed with **bcrypt** (cost factor 10) — never stored in plain text
- JWT tokens expire after **24 hours**
- The `password` field is excluded from all JSON responses via `json:"-"`
- Never commit your `.env` file — add it to `.gitignore`

```gitignore
.env
```

---

## 🐛 Known Limitations

- Transactions are not yet scoped per user — all users see all transactions
- JWT secret is not yet read from `JWT_SECRET` env var in middleware (hardcoded as `"secret"`)
- No pagination on `GET /api/transactions`

---

## 📄 License

MIT
