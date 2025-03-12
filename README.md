# Port Service

A microservice that manages port data with memory-efficient processing of large JSON files.

## Features

- Stream processing of large JSON files
- In-memory database for port data
- RESTful API for port management
- Docker support
- Memory-efficient design (200MB limit)
- Signal handling for graceful shutdown
- Thread-safe concurrent operations
- Comprehensive validation and error handling
- Detailed statistics tracking

## Project Structure

The project follows hexagonal architecture principles:

```
.
├── cmd/                    # Application entry points
│   └── portservice/       # Main service executable
├── internal/              # Private application code
│   ├── domain/           # Domain models and interfaces
│   ├── ports/            # Ports (interfaces) layer
│   │   ├── in/          # Input ports (application services)
│   │   └── out/         # Output ports (repository interfaces)
│   ├── adapters/         # Adapters layer
│   │   ├── primary/     # Primary/Driving adapters (HTTP, gRPC)
│   │   └── secondary/   # Secondary/Driven adapters (DB, external services)
│   └── core/            # Core business logic
└── pkg/                  # Public library code
    └── common/          # Shared utilities

```

## Requirements

- Go 1.21 or higher
- Docker

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/portservice.git
cd portservice
```

2. Install dependencies:
```bash
go mod download
```

## Running the Service

1. Build and run using Docker:
```bash
docker build -t portservice .
docker run -p 8080:8080 portservice
```

2. Run locally:
```bash
go run cmd/portservice/main.go
```

## Testing

1. Run all tests:
```bash
go test ./...
```

2. Run tests with race detection:
```bash
go test -race ./...
```

3. Run tests with coverage:
```bash
go test -cover ./...
```

4. Run tests with verbose output and coverage:
```bash
go test -v -cover ./...
```

### Test Coverage

Current test coverage by package:
- Domain Layer: 100%
- Repository Layer: 96.8%
- Service Layer: 84.5%

## API Documentation

### Endpoints

#### Create or Update Port
- Method: `POST`
- Path: `/api/v1/ports`
- Content-Type: `application/json`
- Request Body: Port object
```json
{
    "id": "AEAJM",
    "name": "Ajman",
    "city": "Ajman",
    "country": "United Arab Emirates",
    "coordinates": [55.5136433, 25.4052165],
    "province": "Ajman",
    "timezone": "Asia/Dubai",
    "unlocs": ["AEAJM"],
    "code": "52000"
}
```
- Response: 200 OK on success

#### Get Port by ID
- Method: `GET`
- Path: `/api/v1/ports/{id}`
- Response: Port object or 404 Not Found

#### Process Ports File
- Method: `POST`
- Path: `/api/v1/ports/file`
- Content-Type: `multipart/form-data`
- Form Field: `file` (JSON file)
- Response: 200 OK on success

### Error Responses
- 400 Bad Request: Invalid input data
- 404 Not Found: Port not found
- 500 Internal Server Error: Server-side error

## Configuration

The service can be configured using environment variables:

- `PORT` - HTTP server port (default: 8080)
- `MAX_MEMORY_MB` - Maximum memory limit in MB (default: 200)
- `READ_TIMEOUT` - HTTP read timeout in seconds (default: 30)
- `WRITE_TIMEOUT` - HTTP write timeout in seconds (default: 30)
- `SHUTDOWN_TIMEOUT` - Graceful shutdown timeout in seconds (default: 30)

## Performance

The service is designed to handle large JSON files efficiently:
- Stream processing to minimize memory usage
- Concurrent file processing for better performance
- Thread-safe operations for concurrent access
- Memory-efficient data structures
- Graceful handling of malformed data

### Benchmarks
- Can process 1000 ports in under 100ms
- Can handle port data with large fields (100KB+ names)
- Maintains consistent performance under concurrent access

## Error Handling

The service includes comprehensive error handling:
- Input validation for all port fields
- Coordinate validation (longitude: -180 to 180, latitude: -90 to 90)
- Malformed JSON detection
- Duplicate port handling
- Context cancellation support
- Resource cleanup on errors

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Development

### Using Make Commands

The project includes a Makefile for common development tasks. View all available commands:
```bash
make help
```

Common commands:
```bash
# Build the application
make build

# Run tests
make test

# Run tests with race detection
make test-race

# Generate test coverage report
make coverage

# Run linter
make lint

# Run all checks (lint, test, coverage)
make check

# Build and run with Docker
make docker-build
make docker-run

# Start all services with docker-compose
make compose-up

# Install development tools
make tools
```

### Development Tools

Required development tools are installed via:
```bash
make tools
```

This installs:
- golangci-lint (code linting)

## License

This project is licensed under the MIT License - see the LICENSE file for details.

