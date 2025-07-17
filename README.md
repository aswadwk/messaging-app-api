# Messaging Task Go

Sebuah layanan API yang dibangun dengan bahasa pemrograman Go (Golang) untuk menangani operasi Messaging. Layanan ini menyediakan endpoint API yang terdokumentasi dengan baik menggunakan OpenAPI/Swagger.

## Features

- **OpenAPI Documentation**: Terdokumentasi dengan baik menggunakan Swagger.
- **Database Migrations**: Menggunakan `golang-migrate` untuk mengelola migrasi database.
- **Hot Reloading**: Mendukung hot reloading untuk pengembangan yang lebih cepat.
- **Modular Architecture**: Terstruktur dengan baik untuk memisahkan logika
- **RabbitMQ Integration**: Menggunakan RabbitMQ untuk message queuing
- **PostgreSQL Partitioning**: Partisi database berdasarkan tenant_id
- **Worker Pool Pattern**: Menggunakan worker pool untuk concurrent processing

## Prerequisites

- Go 1.24 or later
- PostgreSQL 13 or later
- RabbitMQ 3.8 or later
- Swag CLI for documentation generation

## Installation

### Install dependencies:

```bash
go mod download
go install github.com/swaggo/swag/cmd/swag@latest
```

## Development

Project Structure

```
.
├── Makefile
├── air.toml                  # Air configuration for hot reload
├── README.md
├── bin                       # Binary location
│   └── app
├── cmd                       # Application entry points
│   └── server                # Main server implementation
│       └── main.go
├── db                        # Database related files
│   └── migrations            # Database migrations
│       ├── 000001_create_messages_table.up.sql
│       └── 000001_create_messages_table.down.sql
├── docker-compose.yml        # Docker compose configuration
├── dto                       # Data Transfer Objects
│   ├── message_dto.go
│   ├── query_dto.go
│   ├── response.go
│   └── tenant_dto.go
├── go.mod
├── go.sum
├── internal                  # Internal application code
│   ├── config                # Configuration files
│   │   ├── config.go
│   │   └── gorm.go
│   ├── docs                  # Generated OpenAPI documentation
│   │   ├── docs.go
│   │   ├── index.html
│   │   ├── swagger.json
│   │   ├── swagger.yaml
│   │   └── v3
│   │       ├── openapi.json
│   │       └── README.md
│   ├── handlers              # API handlers
│   │   ├── message_handler.go
│   │   └── tenant_handler.go
│   ├── middleware            # Middleware
│   │   └── auth.go
│   ├── models                # Data models
│   │   ├── message.go
│   │   └── user.go
│   ├── repositories          # Repository pattern
│   │   └── message_respository.go
│   ├── routes                # API routes
│   │   ├── message_route.go
│   │   ├── routes.go
│   │   └── tenant_route.go
│   ├── services              # Business logic
│   │   ├── jwt_service.go
│   │   ├── publisher_service.go
│   │   ├── rabbitmq_service.go
│   │   ├── tenant_manager_service.go
│   │   └── worker_pool_service.go
│   └── utils                 # Utilities
│       ├── constanta.go
│       ├── jwt.go
│       ├── query.go
│       └── response.go
├── modd.conf                 # Modd configuration
├── static                    # Static files
│   └── favicon.ico
├── storage                   # Storage directory
│   └── app
└── tmp                       # Temporary files
```

## Generating OpenAPI Documentation

To generate or update OpenAPI/Swagger documentation:

Run with make command:

```bash
make docs
```
or
```bash
swag init --dir ./cmd/server --output ./internal/docs
```

The documentation will be available when the server is running at
- [/docs](http://127.0.0.1:8081/docs/) or 
- [/swagger](http://127.0.0.1:8081/swagger/) when the server is running.

## Database Migrations

Install `golang-migrate`:

```bash
brew install golang-migrate
```

Generate Migrations:

```bash
migrate create -ext sql -dir db/migrations -seq create_name_table_table
```

Run Migrations Up:
On Production:

```bash
./bin/app migrate
```

On Development:

```bash
go run cmd/server/main.go migrate
```

```bash
migrate -path db/migrations -database database_url up
```

Run Migrations Down:

```bash
migrate -path db/migrations -database database_url down
```

## Running the Application

Choose one of these hot-reload tools for development:

### Using Gow

```bash
go install github.com/mitranim/gow@latest
```

Run the application:

```bash
gow run cmd/server/main.go
```

## Building and Deployment

Using Makefile

We provide a Makefile for easy build automation:

| Command      | Description                      |
| ------------ | -------------------------------- |
| `make dev`   | Run with hot reload              |
| `make docs`  | Generate OpenAPI documentation   |
| `make build` | Build binary to `bin/app`        |
| `make tidy`  | Clean and sync dependencies      |

Example build:

```bash
make build
```

The resulting binary will be located at `bin/app`.

## Configuration

Configuration is managed through environment variables. Copy `.env.example` to `.env` and modify as needed:

```env
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=messaging_db

# RabbitMQ Configuration
RABBITMQ_URL=amqp://guest:guest@localhost:5672/

# JWT Configuration
JWT_SECRET=your-secret-key

# Server Configuration
SERVER_PORT=8080
```

## API Endpoints

### Tenant Management
- `POST /tenants` - Create tenant and partition
- `DELETE /tenants/:id` - Delete tenant and partition

### Message Management
- `POST /messages` - Send message to queue
- `GET /messages` - Get messages with pagination

## Architecture

### Key Components

1. **Tenant Manager**: Manages tenant creation and PostgreSQL partitioning
2. **Publisher Service**: Handles message publishing to RabbitMQ
3. **Worker Pool**: Concurrent message processing
4. **Repository Pattern**: Database abstraction layer
5. **JWT Middleware**: Authentication and authorization

### Database Design

- **Partitioned Tables**: Messages are partitioned by `tenant_id`
- **Primary Key**: Composite key of `(id, tenant_id)`
- **JSONB Storage**: Message payload stored as JSONB for flexibility

## Deployment

To deploy the service:

Build the binary:

```bash
make build
```

Run the binary:

```bash
./bin/app
```

## Testing

Run tests:

```bash
go test ./...
```
