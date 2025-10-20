# Async Workflow Orchestrator

[![Go Version](https://img.shields.io/badge/Go-1.25.1-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-Apache%202.0-green.svg)](LICENSE)

A high-performance, event-driven workflow orchestration system built with Go, designed to handle complex asynchronous task processing at scale. This system leverages Apache Kafka for reliable event streaming, PostgreSQL for persistent storage, and Redis for caching, providing a robust foundation for building distributed workflow applications.

## 🎯 Overview

Async Workflow Orchestrator is a production-ready backend system that enables you to define, trigger, and manage complex workflows through a declarative DSL. It implements a state machine pattern to orchestrate multi-step processes, providing features like automatic retries, failure handling, task chaining, and comprehensive observability.

**Key Features:**
- 🔄 **Event-Driven Architecture**: Built on Apache Kafka for reliable, scalable message processing
- 🎭 **State Machine Orchestration**: Intelligent workflow coordination with automatic step transitions
- 🔁 **Automatic Retry Logic**: Built-in retry mechanisms with configurable policies
- 📊 **Workflow Instance Tracking**: Full visibility into workflow execution state and history
- 🔌 **Extensible Task System**: Easy integration of custom task executors
- 📝 **Declarative Workflow DSL**: Define workflows using YAML configuration
- 🏗️ **Microservices Architecture**: Separate API, Orchestrator, and Worker services
- 📈 **Production-Ready**: Structured logging, health checks, and graceful shutdown

## 🏗️ Architecture

The system follows a microservices architecture with three core components:

```
┌─────────────┐         ┌──────────────┐         ┌─────────────┐
│             │  HTTP   │              │  Kafka  │             │
│  API Server │◄───────►│ Orchestrator │◄───────►│   Workers   │
│             │         │              │         │             │
└──────┬──────┘         └──────┬───────┘         └──────┬──────┘
       │                       │                        │
       │                       │                        │
       └───────────────────────┴────────────────────────┘
                               │
                    ┌──────────┴──────────┐
                    │                     │
            ┌───────▼────┐        ┌──────▼──────┐
            │ PostgreSQL │        │    Redis    │
            │            │        │             │
            └────────────┘        └─────────────┘
```

### Components

1. **API Service** (`cmd/api/`): RESTful API endpoints for workflow management
2. **Orchestrator Service** (`cmd/orchestrator/`): Core workflow coordination and state machine management
3. **Worker Service** (`cmd/worker/`): Task execution and processing

### Data Flow

1. Client submits workflow request via REST API
2. API creates workflow instance and publishes initial task event to Kafka
3. Workers consume task events, execute tasks, and publish completion events
4. Orchestrator consumes completions, updates state, and triggers next tasks
5. Process continues until workflow completes or fails

## 🚀 Getting Started

### Prerequisites

- **Go**: 1.25.1 or higher
- **Docker**: For running dependencies
- **Docker Compose**: For local development environment
- **Task**: Task runner (optional, for convenience commands)

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/Vighnesh-V-H/async.git
   cd async
   ```

2. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

   Required environment variables:
   ```env
   # Database
   DATABASE_URL=postgres://pg:pradyuman@localhost:5432/db?sslmode=disable

   # Kafka
   KAFKA_BROKERS=localhost:29092

   # Redis
   REDIS_URL=localhost:6379

   # Application
   PORT=8080
   LOG_LEVEL=info
   LOG_FORMAT=json
   ENVIRONMENT=development
   ```

3. **Start infrastructure services**
   ```bash
   docker-compose up -d
   ```

   This will start:
   - PostgreSQL (port 5432)
   - Adminer (port 7080) - Database management UI
   - Redis (port 6379)
   - Kafka (ports 9092, 29092)
   - Kafka UI (port 8081)

4. **Run database migrations**
   ```bash
   task migrations:up
   ```

5. **Start the orchestrator service**
   ```bash
   task run:orc
   # Or manually: go run cmd/orchestrator/main.go
   ```

### Quick Start Example

1. **Create a workflow**
   ```bash
   curl -X POST http://localhost:8080/workflow/create \
     -H "Content-Type: application/json" \
     -d '{
       "name": "audio-generation",
       "event": "generate_audio",
       "message": "Audio generation workflow",
       "steps": 3,
       "handler_url": "http://worker:8082/tasks"
     }'
   ```

2. **Trigger audio generation**
   ```bash
   curl -X POST http://localhost:8080/audio/generate \
     -H "Content-Type: application/json" \
     -d '{
       "text": "Hello, world!",
       "voice": "en-US-Standard-A",
       "metadata": {
         "format": "mp3",
         "sample_rate": 44100
       }
     }'
   ```

3. **Check execution status**
   ```bash
   curl http://localhost:8080/audio/status/{execution_id}
   ```

## 📁 Project Structure

```
async/
├── cmd/                        # Application entry points
│   ├── api/                    # REST API server
│   ├── orchestrator/           # Workflow orchestrator service
│   └── worker/                 # Task worker service
├── internal/                   # Private application code
│   ├── events/                 # Kafka producer/consumer
│   ├── handler/                # HTTP request handlers
│   ├── logger/                 # Structured logging
│   ├── models/                 # Domain models and entities
│   ├── orchestrator/           # State machine logic
│   ├── repositories/           # Data access layer
│   ├── router/                 # HTTP route definitions
│   ├── service/                # Business logic layer
│   └── workers/                # Task executors
├── pkg/                        # Public reusable packages
│   ├── cache/                  # Redis caching
│   ├── database/               # PostgreSQL connection and migrations
│   ├── dsl/                    # Workflow DSL parser (planned)
│   ├── kafka/                  # Kafka client initialization
│   └── observability/          # Metrics and tracing (planned)
├── configs/                    # Configuration files
├── deploy/                     # Deployment manifests
├── workflows/                  # Workflow definitions (YAML)
└── docker-compose.yml          # Local development stack
```

## 🔧 Development

### Available Tasks

Using [Task](https://taskfile.dev/), you can run:

```bash
task help              # List all available tasks
task run:orc           # Run orchestrator service
task migrations:new    # Create new migration
task migrations:up     # Apply migrations
task migrations:down   # Rollback last migration
task tidy              # Format code and tidy dependencies
```

### Code Organization

The project follows clean architecture principles:

- **Models** (`internal/models`): Core domain entities with GORM annotations
- **Repositories** (`internal/repositories`): Data persistence layer
- **Services** (`internal/service`): Business logic and orchestration
- **Handlers** (`internal/handler`): HTTP request/response handling
- **Events** (`internal/events`): Kafka event publishing and consumption

## 📊 Database Schema

The system uses PostgreSQL with the following core tables:

- `workflows`: Workflow definitions and metadata
- `workflow_instances`: Individual workflow execution instances
- `tasks`: Task execution records
- `history_entries`: Audit trail of workflow events
- `workflow_registries`: Worker registration and health tracking

## 🎭 Workflow DSL

Workflows are defined using a YAML-based DSL. Example workflow structure:

```yaml
name: audio-generation-workflow
version: 1.0
triggers:
  - type: webhook
    event: "generate_audio"

states:
  - id: extract_text
    type: ai_task
    model: "gpt-4"
    on_success: convert_to_audio
    retries: 2
    timeout: 60s

  - id: convert_to_audio
    type: task
    action: "tts_generate"
    on_success: upload_audio
    
  - id: upload_audio
    type: http_call
    method: POST
    url: "https://storage.api/store/audio"
    on_success: send_notification
    retries: 3
```

## 🔍 Monitoring & Observability

### Health Checks

```bash
curl http://localhost:8080/health
```

### Logs

All services use structured JSON logging with zerolog:
- Configurable log levels (debug, info, warn, error)
- Request tracing with execution IDs
- Contextual metadata for debugging

### Kafka UI

Access Kafka UI at http://localhost:8081 to monitor:
- Topic messages
- Consumer groups
- Broker health

### Database Admin

Access Adminer at http://localhost:7080 to:
- View database tables
- Run SQL queries
- Monitor workflow instances

## 📋 TODO & Roadmap

### High Priority - Core Features
- [ ] **Authentication & Authorization**: Implement JWT-based auth middleware for API endpoints
- [ ] **DSL Parser Implementation**: Complete YAML workflow DSL parser in `pkg/dsl/`
- [ ] **Worker Service**: Implement actual worker service in `cmd/worker/main.go`
- [ ] **Task Executors**: Build pluggable task executor system with sample implementations
- [ ] **Retry & Timeout Logic**: Implement exponential backoff and task timeout handling
- [ ] **Dead Letter Queue**: Add DLQ for failed tasks with manual intervention support

### Medium Priority - Production Readiness
- [ ] **API Rate Limiting**: Add rate limiting middleware to prevent abuse
- [ ] **Request Validation**: Comprehensive input validation for all API endpoints
- [ ] **Error Handling**: Standardized error responses and error codes
- [ ] **Graceful Degradation**: Circuit breakers for external service calls
- [ ] **Database Connection Pooling**: Optimize database connections with proper pooling
- [ ] **Metrics & Tracing**: Integrate Prometheus metrics and distributed tracing
- [ ] **API Documentation**: Generate OpenAPI/Swagger documentation
- [ ] **Unit & Integration Tests**: Comprehensive test coverage (target: >80%)

### Low Priority - Enhanced Features
- [ ] **Webhook Support**: Implement webhook notifications for workflow events
- [ ] **Conditional Branching**: Support for complex workflow conditions
- [ ] **Parallel Task Execution**: Execute multiple tasks concurrently
- [ ] **Scheduled Workflows**: Cron-based workflow triggers
- [ ] **Workflow Versioning**: Support multiple versions of same workflow
- [ ] **Admin Dashboard**: Web UI for workflow management
- [ ] **CLI Tool**: Command-line tool for workflow operations
- [ ] **Multi-tenancy**: Support for isolated tenant workspaces

### Infrastructure & DevOps
- [ ] **Kubernetes Deployment**: Helm charts for K8s deployment
- [ ] **CI/CD Pipeline**: GitHub Actions for automated testing and deployment
- [ ] **Docker Images**: Multi-stage Docker builds for all services
- [ ] **Performance Testing**: Load testing and benchmarking
- [ ] **Security Audit**: Dependency scanning and security best practices
- [ ] **Backup & Recovery**: Automated database backup strategy

## 🤝 Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the Apache License 2.0 - see the LICENSE file for details.

## 🙏 Acknowledgments

Built with these excellent open-source projects:
- [Gin](https://github.com/gin-gonic/gin) - HTTP web framework
- [GORM](https://gorm.io/) - ORM library
- [Confluent Kafka Go](https://github.com/confluentinc/confluent-kafka-go) - Kafka client
- [Zerolog](https://github.com/rs/zerolog) - Structured logging
- [Redis](https://redis.io/) - In-memory data store
- [PostgreSQL](https://www.postgresql.org/) - Relational database

## 📧 Contact

For questions and support, please open an issue on GitHub.

---

**Note**: This is an active development project. APIs and features may change. Check the TODO section for planned improvements and current limitations.