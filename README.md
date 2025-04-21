# API Gateway Sample

A robust API Gateway implementation in Go using Clean Architecture principles. This gateway serves as a single entry point for client applications to access various backend services, providing essential features like authentication, rate limiting, request routing, and monitoring.

## Features

- 🔒 JWT Authentication
- ⚡ Rate Limiting
- 🔄 Request Routing
- 📝 Logging
- 🚦 CORS Support
- 💾 Redis Caching
- 📊 Service Discovery
- ⚖️ Load Balancing
- 🔍 Request/Response Transformation
- 🛡️ Circuit Breaking

## Architecture

The project follows Clean Architecture principles with the following layers:

```
api-gateway/
├── cmd/                          # Application entry points
│   └── api/                     
│       └── main.go              
├── internal/                     # Private application code
│   ├── domain/                   # Business entities & interfaces
│   ├── application/              # Use cases & DTOs
│   ├── infrastructure/           # External implementations
│   └── interfaces/               # API handlers & middleware
├── pkg/                          # Public libraries
│   ├── logger/                   
│   ├── config/                   
│   └── errors/                   
└── configs/                      # Configuration files
```

### Application Flow

```mermaid
graph TD
    Client[Client] -->|HTTP Request| Gateway[API Gateway]
    Gateway -->|1. Logging| Log[Log Request]
    Gateway -->|2. Authentication| Auth[JWT Auth]
    Gateway -->|3. Rate Limiting| Rate[Rate Limiter]
    Gateway -->|4. Route| Route[Service Discovery]
    Gateway -->|5. Transform| Transform[Request Transform]
    Gateway -->|6. Cache Check| Cache[Redis Cache]
    Cache -->|Cache Hit| Response[Response Transform]
    Cache -->|Cache Miss| Backend[Backend Service]
    Backend --> Response
    Response --> Client
```

## Prerequisites

- Go 1.24 or higher
- Docker and Docker Compose
- PostgreSQL (for service registry)
- Redis (for caching and rate limiting)

## Getting Started

1. Clone the repository:
```bash
git clone https://github.com/yourusername/api-gateway-sample.git
cd api-gateway-sample
```

2. Start the services using Docker Compose:
```bash
docker-compose up -d
```

The API Gateway will be available at `http://localhost:8080`.

## Configuration

Configuration is managed through environment variables. Here are the key configurations:

```yaml
# Server Configuration
API_GATEWAY_SERVER_PORT: 8080
API_GATEWAY_SERVER_READTIMEOUT: 30s
API_GATEWAY_SERVER_WRITETIMEOUT: 30s

# Database Configuration
API_GATEWAY_DATABASE_HOST: postgres
API_GATEWAY_DATABASE_PORT: 5432
API_GATEWAY_DATABASE_USER: postgres
API_GATEWAY_DATABASE_PASSWORD: postgres
API_GATEWAY_DATABASE_DATABASE: api_gateway

# Redis Configuration
API_GATEWAY_REDIS_ADDRESS: redis:6379
API_GATEWAY_REDIS_PASSWORD: ""
API_GATEWAY_REDIS_DB: 0

# Auth Configuration
API_GATEWAY_AUTH_SECRETKEY: your-secret-key
API_GATEWAY_AUTH_ISSUER: api-gateway
API_GATEWAY_AUTH_EXPIRATION: 24h

# Logging Configuration
API_GATEWAY_LOGGING_LEVEL: info
API_GATEWAY_LOGGING_DEVELOPMENT: true
```

## API Usage Examples

### 1. Authentication

Get a JWT token:
```bash
curl -X POST http://localhost:8080/auth \
  -H "Content-Type: application/json" \
  -d '{"username": "user", "password": "pass"}'
```

Response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

### 2. Using Protected Endpoints

Make a request to a protected endpoint:
```bash
curl http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

### 3. Service Registration

Register a new backend service:
```bash
curl -X POST http://localhost:8080/api/services \
  -H "Content-Type: application/json" \
  -d '{
    "name": "users-service",
    "baseURL": "http://users-service:8080",
    "endpoints": [
      {
        "path": "/api/v1/users",
        "methods": ["GET", "POST"],
        "rateLimit": 100,
        "authRequired": true
      }
    ]
  }'
```

## Development

### Running Tests

Run unit tests:
```bash
go test ./...
```

Run integration tests:
```bash
go test -tags=integration ./...
```

### Adding a New Service

1. Register the service using the API
2. Configure rate limits and authentication requirements
3. Update the service discovery configuration if needed

## Monitoring

The API Gateway provides several endpoints for monitoring:

- `/health` - Health check endpoint
- `/metrics` - Prometheus metrics (if enabled)
- `/debug/pprof` - Go profiling endpoints (in development)

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Roadmap

- [ ] Implement service repository
- [ ] Add cache repository implementation
- [ ] Complete application layer use cases
- [ ] Add comprehensive test coverage
- [ ] Add service discovery integration
- [ ] Implement circuit breaker pattern
- [ ] Add metrics collection
- [ ] Improve documentation
- [ ] Add OpenAPI/Swagger documentation
- [ ] Add performance benchmarks