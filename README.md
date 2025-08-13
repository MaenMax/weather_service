# Weather Service

A production-ready, scalable weather service built in Go

## Project Overview

This weather service provides forecasted weather information using the **National Weather Service API** (as required by the assessment). It demonstrates:

- **Clean Architecture** with proper separation of concerns
- **Production-Ready Features** including caching, metrics, logging, and graceful shutdown
- **Scalability** through configuration management and connection pooling
- **Enterprise Standards** with comprehensive error handling, validation, and monitoring

##  Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   HTTP Layer    │    │  Service Layer  │    │  External APIs  │
│  (Handlers)     │◄──►│ (Weather Svc)   │◄──►│   (NWS API)     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │
         ▼                       ▼
┌─────────────────┐    ┌─────────────────┐
│   Middleware    │    │     Cache       │
│ (Logging, CORS) │    │  (In-Memory)    │
└─────────────────┘    └─────────────────┘
```

##  Key Features

- **National Weather Service API** integration (not OpenWeather)
- **Forecasted weather** for today and upcoming periods
- **Temperature characterization** (hot/cold/moderate)
- **Latitude/Longitude** coordinate support

### Other features

- 🔧 **Configuration Management** with Viper
-  **Metrics & Monitoring** with Prometheus
-  **Intelligent Caching** with TTL support
-  **Structured Logging** with Logrus
-  **Error Handling** with proper HTTP status codes
-  **Graceful Shutdown** with context management
-  **Comprehensive Testing** with mocks
-  **Clean Code Structure** with proper packages
-  **CORS Support** for web applications
-  **Health Check Endpoints**

## Quick Start

### Prerequisites
- Go 1.21+
- Git

### Installation & Running

1. **Clone and navigate to the project:**
```bash
git clone <your-repo-url>
cd weather_service
```

2. **Install dependencies:**
```bash
go mod tidy
```

3. **Run the service:**
```bash
go run cmd/server/main.go
```

The service will start on port 8080 with default configuration.

### Configuration

The service uses `config/config.yaml` for configuration. Key settings:

```yaml
server:
  port: "8080"
  read_timeout: "15s"
  write_timeout: "15s"

weather:
  base_url: "https://api.weather.gov"
  request_timeout: "10s"
  cache_ttl: "15m"

logging:
  level: "info"
  format: "json"

metrics:
  enabled: true
  path: "/metrics"
```

Environment variables override config file settings:
```bash
export WEATHER_BASE_URL="https://api.weather.gov"
export SERVER_PORT="9090"
```

## API Usage

### Get Weather Forecast

```bash
curl "http://localhost:8080/weather?lat=40.7128&lon=-74.0060"
```

**Response:**
```json
{
  "location": {
    "latitude": 40.7128,
    "longitude": -74.0060
  },
  "current_weather": {
    "temperature_celsius": 23.89,
    "temperature_fahrenheit": 75,
    "condition": "Partly Cloudy",
    "description": "Partly sunny with a high near 75."
  },
  "forecast": [
    {
      "period": "Today",
      "start_time": "2024-01-01T06:00:00-05:00",
      "end_time": "2024-01-01T18:00:00-05:00",
      "temperature_celsius": 23.89,
      "condition": "Partly Cloudy",
      "description": "Partly sunny with a high near 75."
    }
  ],
  "temperature_type": "moderate",
  "generated_at": "2024-01-01T12:00:00Z"
}
```

### Health Check

```bash
curl "http://localhost:8080/health"
```

### Metrics (if enabled)

```bash
curl "http://localhost:8080/metrics"
```

##  Testing

Run the test suite:

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/services
```

##  Project Structure

```
weather_service/
├── cmd/
│   └── server/           # Application entry point
├── config/               # Configuration management
├── internal/             # Private application code
│   ├── cache/            # Caching layer
│   ├── handlers/         # HTTP handlers
│   ├── models/           # Data models
│   └── services/         # Business logic
├── config.yaml           # Configuration file
├── go.mod                # Go module file
└── README.md             # This file
```

## Development

### Adding New Features

1. **Models**: Add new data structures in `internal/models/`
2. **Services**: Implement business logic in `internal/services/`
3. **Handlers**: Create HTTP endpoints in `internal/handlers/`
4. **Tests**: Write comprehensive tests for new functionality

### Code Quality

- **Go Modules**: Modern dependency management
- **Go Lint**: Code quality enforcement
- **Go Vet**: Static analysis
- **Go Test**: Comprehensive testing

## Monitoring & Observability

### Logging
- **Structured JSON logging** for production
- **Configurable log levels** (debug, info, warn, error)
- **Request/response logging** with timing information

### Metrics
- **Prometheus metrics** endpoint
- **Request counts, durations, and error rates**
- **Cache hit/miss statistics**

### Health Checks
- **Liveness probe** endpoint
- **Readiness check** for load balancers

## Production Deployment

### Docker Support
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o main cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main"]
```

### Environment Variables
```bash
# Production settings
export LOGGING_LEVEL="warn"
export METRICS_ENABLED="true"
export WEATHER_CACHE_TTL="30m"
export SERVER_READ_TIMEOUT="30s"
```

### Load Balancing
- **Health check endpoint** for load balancer integration
- **Graceful shutdown** support for zero-downtime deployments
- **Connection pooling** for external API calls

## Security Features

- **Input validation** for coordinates
- **Rate limiting** through connection pooling
- **CORS configuration** for web security
- **User-Agent headers** for API compliance
- **Error message sanitization** to prevent information leakage

## Scalability Considerations

### Current Implementation
- **In-memory caching** with TTL (suitable for single instances)
- **Connection pooling** for external API calls
- **Configurable timeouts** and retry logic

### Future Enhancements
- **Redis caching** for distributed deployments
- **Circuit breaker** pattern for external API resilience
- **Horizontal scaling** with load balancers
- **Database persistence** for historical data

## Contributing

1. Fork the repository
2. Create a feature branch
3. Implement changes with tests
4. Submit a pull request

## License

This project is created for assessment purposes.

## Assessment Notes


### Shortcuts Taken

- **In-memory cache**: For simplicity, production would use Redis
- **Basic metrics**: Could add custom business metrics
- **Simple retry logic**: Could implement exponential backoff
- **Single instance**: Could add clustering and load balancing

