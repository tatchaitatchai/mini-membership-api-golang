# Mini Membership API

ระบบ API สำหรับจัดการคะแนนสะสมสมาชิก (Membership Points System) ที่พัฒนาด้วย Golang

## Features

- 🔐 **Authentication & Authorization** - JWT-based authentication with branch-level access control
- 👥 **Member Management** - Create, read, update member information
- 💰 **Points Transaction** - Add, deduct, redeem, adjust points with product type tracking
- 🏢 **Branch Isolation** - Staff can only access members in their assigned branch
- 📊 **Point Categories** - Track points by product type (1.0L, 1.5L) and milestone scores

## Tech Stack

- **Framework**: Gin (Go web framework)
- **Database**: PostgreSQL 15
- **ORM/Query**: sqlx
- **Authentication**: JWT (golang-jwt/jwt)
- **Password Hashing**: bcrypt
- **Container**: Docker & Docker Compose

## Project Structure

```
api/
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── config/
│   └── config.go                # Configuration management
├── internal/
│   ├── domain/                  # Domain models
│   │   ├── member.go
│   │   ├── staff_user.go
│   │   └── transaction.go
│   ├── handler/                 # HTTP handlers
│   │   ├── auth_handler.go
│   │   ├── member_handler.go
│   │   └── transaction_handler.go
│   ├── middleware/              # HTTP middleware
│   │   ├── auth.go
│   │   └── cors.go
│   ├── repository/              # Data access layer
│   │   ├── member_repository.go
│   │   ├── staff_user_repository.go
│   │   └── transaction_repository.go
│   └── service/                 # Business logic layer
│       ├── auth_service.go
│       ├── member_service.go
│       └── transaction_service.go
├── migrations/
│   └── 001_initial_schema.sql   # Database schema
├── pkg/
│   └── database/
│       └── postgres.go          # Database connection
├── .env                         # Environment variables
├── .env.example                 # Example environment variables
├── docker-compose.yml           # Docker Compose configuration
├── go.mod                       # Go module dependencies
├── Makefile                     # Build commands
└── README.md
```

## Getting Started

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- Make (optional)

### Installation

1. **Clone the repository**
```bash
cd mini-membership-api-golang
```

2. **Install dependencies**
```bash
go mod tidy
```

3. **Setup environment variables**
```bash
cp .env.example .env
# Edit .env and set JWT_SECRET
```

4. **Start PostgreSQL**
```bash
docker-compose up -d
# or
make docker-up
```

5. **Run migrations**
```bash
make migrate-up
# or
psql postgresql://mini:1234@localhost:5432/mini_membership -f migrations/001_initial_schema.sql
```

6. **Run the application**
```bash
make run
# or
go run cmd/api/main.go
```

The API will be available at `http://localhost:8080`

## API Endpoints

### Authentication

- `POST /api/v1/auth/register` - Register new staff user
- `POST /api/v1/auth/login` - Login and get JWT token

### Members (Protected)

- `GET /api/v1/members` - List all members in staff's branch
- `POST /api/v1/members` - Create new member
- `GET /api/v1/members/:id` - Get member details
- `PUT /api/v1/members/:id` - Update member information

### Transactions (Protected)

- `POST /api/v1/transactions` - Create new point transaction
- `GET /api/v1/transactions/member/:member_id` - Get member's transaction history
- `GET /api/v1/transactions/branch` - Get all transactions in staff's branch

### Health Check

- `GET /health` - Check API health status

## Example Usage

### 1. Register Staff User

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "staff@thecoffee.com",
    "password": "password123",
    "branch": "Bangkok"
  }'
```

### 2. Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "staff@thecoffee.com",
    "password": "password123"
  }'
```

### 3. Create Member

```bash
curl -X POST http://localhost:8080/api/v1/members \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "name": "John Doe",
    "last4": "1234",
    "membership_number": "MB001"
  }'
```

### 4. Add Points

```bash
curl -X POST http://localhost:8080/api/v1/transactions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "member_id": "MEMBER_UUID",
    "action": "add",
    "product_type": "1.0_liter",
    "points": 10,
    "receipt_text": "Purchase #12345"
  }'
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| SERVER_PORT | API server port | 8080 |
| GIN_MODE | Gin mode (debug/release) | debug |
| POSTGRES_HOST | PostgreSQL host | localhost |
| POSTGRES_PORT | PostgreSQL port | 5432 |
| POSTGRES_USER | PostgreSQL user | mini |
| POSTGRES_PASSWORD | PostgreSQL password | - |
| POSTGRES_DB | PostgreSQL database | mini_membership |
| JWT_SECRET | JWT secret key (required) | - |
| JWT_EXPIRATION | JWT expiration in seconds | 86400 (24h) |

## Database Schema

### staff_users
- Primary authentication table for staff
- Each staff belongs to one branch
- Can only manage members in their branch

### members
- Customer/member information
- Track total points and category-specific points
- Branch association for access control

### member_point_transactions
- Point transaction history
- Actions: add, deduct, redeem, adjust
- Product types: 1.0_liter, 1.5_liter, other

## Development

```bash
# Run tests
make test

# Build binary
make build

# Clean build artifacts
make clean

# View database logs
make docker-logs
```

## License

Private - Mini Membership System
