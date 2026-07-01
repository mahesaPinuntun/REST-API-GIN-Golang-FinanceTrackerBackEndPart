# Finance Tracker API

A RESTful backend API for personal finance tracking built with Go, Gin, GORM, and PostgreSQL. Supports JWT authentication, multi-currency conversion, email confirmation, and full transaction management.

---

## Tech Stack

| Layer | Technology |
|---|---|
| Language | `Go 1.26+` |
| Framework | `Gin v1.12` |
| ORM | `GORM v1.31` |
| Database | `PostgreSQL` via `Neon` |
| Auth | `JWT` — `golang-jwt/jwt v5` |
| Password | `bcrypt` |
| Email | `Resend API` |
| Currency | `fxapi.app` |
| Config | `godotenv` |
| Hosting | `Vercel` |

---

## Project Structure

```
finance-tracker/
├── config/
│   └── database.go
├── controllers/
│   ├── auth_controller.go        # Register & Login
│   ├── transaction_controller.go # Transaction CRUD & Dashboard
│   ├── currency_controller.go    # Currency conversion via fxapi.app
│   └── email.go                  # Email confirmation via Resend
├── middleware/
│   └── jwt.go                    # JWT auth middleware
├── models/
│   ├── user.go                   # User model
│   ├── transaction.go            # Transaction model with activity types
│   └── email_token.go            # Email confirmation token model
├── routes/
│   └── routes.go
├── utils/
│   ├── hash.go                   # bcrypt helpers
│   └── jwt.go                    # Token generation
├── public/
│   └── index.html                # Landing page
├── .env                          # Never commit this
├── go.mod
├── go.sum
└── main.go
```

---

## Prerequisites

- `Go 1.21+` — [go.dev/dl](https://go.dev/dl/)
- `PostgreSQL` database — local or cloud via `Neon`
- `Resend` account for email — [resend.com](https://resend.com)
- `Git`

---

## PostgreSQL Setup

### Option A — Local PostgreSQL

1. Install from [postgresql.org](https://www.postgresql.org/download/)
2. Create a database:

```sql
CREATE DATABASE finance_tracker;
```

`GORM` auto-migrates all tables on first run — no SQL scripts needed.

### Option B — Neon (Recommended)

1. Sign up at [neon.tech](https://neon.tech)
2. Create a new project and database
3. Go to Dashboard → Connection Details
4. Copy the connection string:

```
postgresql://user:password@ep-xxxx.neon.tech/dbname?sslmode=require
```

---

## Local Setup

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

```env
DATABASE_URL=postgresql://user:password@host/dbname?sslmode=require
JWT_SECRET=your_secret_key_here
RESEND_API_KEY=re_xxxxxxxxxxxx
APP_URL=http://localhost:3000
```

### 4. Run the server

```bash
go run main.go
```

Server starts at `http://localhost:3000`

---

## Deployment on Vercel

This project uses the official `Gin` + `Vercel` template with embedded static files.

### 1. Push to GitHub

Make sure your repo is up to date.

### 2. Import on Vercel

1. Go to [vercel.com](https://vercel.com) → New Project → Import from GitHub
2. Select your repository
3. Vercel auto-detects `Go` from `go.mod`

### 3. Set environment variables

In Vercel dashboard → Settings → Environment Variables:

```
DATABASE_URL     = postgresql://user:password@host/dbname?sslmode=require
JWT_SECRET       = your_secret_key_here
RESEND_API_KEY   = re_xxxxxxxxxxxx
APP_URL          = https://your-app.vercel.app
```

### 4. Deploy

Push to `main` branch — Vercel redeploys automatically.

---

## API Reference

### Base URL

```
Local:      http://localhost:3000
Production: https://your-app.vercel.app
```

---

### Public Routes

#### `POST /register` — Create a new user

**Request Body:**
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "yourpassword",
  "salaryAmount": 5000000,
  "salaryCurrency": "IDR",
  "salaryFrequency": "monthly"
}
```

**Response `201`:**
```json
{
  "message": "User created. Please check your email to confirm your account.",
  "is_email_confirmed": false
}
```

A confirmation email is sent automatically via `Resend` on successful registration.

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
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "is_email_confirmed": false,
  "warning": "Your email is not confirmed. Some features may be restricted.",
  "user": {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com"
  }
}
```

**Error responses:**
- `401` — User not found
- `401` — Invalid credential

---

#### `GET /api/auth/confirm?token=xxx` — Confirm email address

Called automatically when the user clicks the link in their confirmation email.

**Response `200`:**
```json
{
  "message": "Email confirmed successfully. You now have full access."
}
```

---

### Protected Routes

All routes below require a `JWT` token in the `Authorization` header:

```
Authorization: Bearer <your_token>
```

---

#### `POST /api/auth/send-confirmation` — Resend confirmation email

Resends the confirmation email to the currently logged-in user.

**Response `200`:**
```json
{
  "message": "Confirmation email sent to john@example.com"
}
```

---

#### `POST /api/transactions` — Create a transaction

**Request Body:**
```json
{
  "title": "Freelance Payment",
  "amount": 500000,
  "type": "income",
  "category": "Freelance",
  "description": "Website project payment",
  "currency": "IDR",
  "status": "POSTED"
}
```

Supported `type` values: `BUY`, `SELL`, `DEPOSIT`, `WITHDRAWAL`, `TRANSFER_IN`, `TRANSFER_OUT`, `DIVIDEND`, `INTEREST`, `CREDIT`, `FEE`, `TAX`, `ADJUSTMENT`, `SPLIT`, `UNKNOWN`

Supported `status` values: `POSTED`, `PENDING`, `DRAFT`, `VOID`

**Response `201`:** Returns the created transaction object.

---

#### `GET /api/transactions` — Get all transactions for logged-in user

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
    "currency": "IDR",
    "status": "POSTED",
    "created_at": "2026-06-20T18:37:54Z"
  }
]
```

---

#### `GET /api/dashboard` — Income/expense summary

**Response `200`:**
```json
{
  "income": 1500000,
  "expense": 300000,
  "balance": 1200000
}
```

---

#### `GET /api/dashboard/convert?currency=USD` — Dashboard in any currency

Converts all transactions and salary to the requested currency using live rates from `fxapi.app`.

**Response `200`:**
```json
{
  "currency": "USD",
  "income": 93.75,
  "expense": 18.75,
  "balance": 75.00,
  "salary_amount": 312.50,
  "salary_frequency": "monthly",
  "salary_currency": "IDR"
}
```

---

#### `GET /api/currency/convert?from=IDR&to=USD&amount=500000` — Convert a specific amount

Uses live exchange rates from `fxapi.app`. No API key required.

**Response `200`:**
```json
{
  "from": "IDR",
  "to": "USD",
  "rate": 0.0000625,
  "amount": 500000,
  "converted": 31.25
}
```

---

#### `GET /api/currency/supported` — List all supported currencies

Returns all 170+ currencies supported by `fxapi.app`.

---

## PowerShell Examples (Windows)

**Register:**
```powershell
$body = @{
    name="John Doe"; email="john@example.com"; password="pass123"
    salaryAmount=5000000; salaryCurrency="IDR"; salaryFrequency="monthly"
} | ConvertTo-Json
Invoke-RestMethod -Uri "http://localhost:3000/register" -Method POST -ContentType "application/json" -Body $body
```

**Login and save token:**
```powershell
$body = @{ email="john@example.com"; password="pass123" } | ConvertTo-Json
$response = Invoke-RestMethod -Uri "http://localhost:3000/login" -Method POST -ContentType "application/json" -Body $body
$token = $response.token
```

**Create transaction:**
```powershell
$body = @{
    title="Salary"; amount=5000000; type="income"
    category="Work"; description="Monthly salary"; currency="IDR"; status="POSTED"
} | ConvertTo-Json
Invoke-RestMethod -Uri "http://localhost:3000/api/transactions" -Method POST -ContentType "application/json" -Headers @{ Authorization="Bearer $token" } -Body $body
```

**Dashboard in USD:**
```powershell
Invoke-RestMethod -Uri "http://localhost:3000/api/dashboard/convert?currency=USD" -Method GET -Headers @{ Authorization="Bearer $token" }
```

**Convert currency:**
```powershell
Invoke-RestMethod -Uri "http://localhost:3000/api/currency/convert?from=IDR&to=USD&amount=500000" -Method GET -Headers @{ Authorization="Bearer $token" }
```

---

## Database Models

### `users`
| Column | Type | Notes |
|---|---|---|
| `id` | `bigint` | Primary key |
| `name` | `text` | |
| `email` | `text` | Unique |
| `password` | `text` | `bcrypt` hashed, never returned in responses |
| `salary_amount` | `numeric` | |
| `salary_currency` | `text` | Default `USD` |
| `salary_frequency` | `text` | Default `monthly` |
| `is_email_confirmed` | `boolean` | Default `false` |

### `transactions`
| Column | Type | Notes |
|---|---|---|
| `id` | `bigint` | Primary key |
| `user_id` | `bigint` | Foreign key → `users` |
| `title` | `text` | |
| `amount` | `numeric` | |
| `type` | `text` | Activity type |
| `status` | `text` | Default `POSTED` |
| `category` | `text` | |
| `description` | `text` | |
| `currency` | `text` | Default `USD` |
| `fee` | `numeric` | Default `0` |
| `asset` | `text` | For `BUY`/`SELL` activities |
| `quantity` | `numeric` | For `BUY`/`SELL` activities |
| `unit_price` | `numeric` | For `BUY`/`SELL` activities |
| `subtype` | `text` | e.g. `DRIP`, `STAKING_REWARD` |
| `metadata` | `text` | `JSON` string for extra context |

### `email_tokens`
| Column | Type | Notes |
|---|---|---|
| `id` | `bigint` | Primary key |
| `user_email` | `text` | References user by email |
| `token` | `text` | Unique, 32-byte hex |
| `expires_at` | `timestamptz` | 24 hours from creation |

---

## Security Notes

- Passwords hashed with `bcrypt` (cost factor 10) — never stored in plain text
- `JWT` tokens expire after 24 hours
- `password` field excluded from all `JSON` responses via `json:"-"`
- Email tokens expire after 24 hours and are deleted after use
- Never commit `.env` — it is listed in `.gitignore`

---

## License

`MIT`