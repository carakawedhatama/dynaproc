# DynaProc - Dynamics x Procurement Integration

A Go-based service concept that processes purchase orders by consuming messages from RabbitMQ and synchronizing them with Dynamics 365.

## Features

- RabbitMQ integration for reliable message queuing
- Dynamics 365 API integration for purchase order synchronization
- PostgreSQL database for order persistence
- Error reporting to GlitchTip for monitoring
- Comprehensive test coverage with mocks

## Prerequisites

- Go 1.22.0 or higher
- PostgreSQL
- RabbitMQ
- Access to Dynamics 365 API
- GlitchTip account (optional, for error reporting)

## Configuration

The service uses environment variables for configuration:

```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=dynaproc
DB_SSL_MODE=disable

# RabbitMQ
RABBITMQ_URL=amqp://guest:guest@localhost:5672/

# Dynamics 365
DYNAMICS_API_URL=https://your-dynamics-instance.com/api

# GlitchTip (optional)
GLITCHTIP_API_URL=https://your-glitchtip-instance.com/api
```

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/carakawedhatama/dynaproc.git
   cd dynaproc
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Build the application:
   ```bash
   go build
   ```

## Usage

1. Start the service:
   ```bash
   ./dynaproc
   ```

2. The service will:
   - Connect to RabbitMQ and start consuming messages
   - Process purchase orders from the database
   - Sync orders with Dynamics 365
   - Report any errors to GlitchTip

## Testing

Run the test suite:
```bash
go test -v ./...
```

The tests cover:
- RabbitMQ message publishing and consumption
- Dynamics 365 API integration
- Database operations
- Error reporting
- Edge cases and error handling

## Architecture

### Components

1. **RabbitMQ Integration** (`rabbit_mq.go`):
   - Handles message queue operations
   - Provides reliable message delivery
   - Implements retry logic for failed operations

2. **Database Layer** (`database.go`):
   - Manages PostgreSQL connections
   - Handles purchase order persistence
   - Tracks sync status

3. **Dynamics 365 Integration** (`sync.go`):
   - Implements API client for Dynamics 365
   - Handles purchase order synchronization
   - Manages API authentication

4. **Error Reporting** (`glitchtip.go`):
   - Sends error reports to GlitchTip
   - Provides error tracking and monitoring
   - Helps with debugging and issue resolution

