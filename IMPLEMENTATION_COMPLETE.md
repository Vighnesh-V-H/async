# Event-Driven Audio Generation Setup Complete! 🎉

## Architecture Overview

```
User Request
     │
     ↓
┌────────────────────┐
│  Orchestrator API  │  POST /audio/generate
│  (Entry Point)     │
└─────────┬──────────┘
          │
          │ 1. Validates request
          │ 2. Creates WorkflowInstance (status: PENDING)
          │ 3. Publishes "extract_text" task → Kafka
          │
          ↓
┌─────────────────────┐
│      Kafka          │
│  - task-queue       │ ← Workers consume from here
│  - task-completions │ ← Orchestrator consumes from here
└─────────┬───────────┘
          │
          ↓
┌─────────────────────┐
│    Worker Pool      │
│  - extract_text     │
│  - generate_audio   │
│  - store_audio      │
└─────────┬───────────┘
          │
          │ Publishes completion → Kafka
          │
          ↓
┌─────────────────────┐
│ Orchestrator        │
│ Consumer            │
│ (State Machine)     │
│                     │
│ - Processes         │
│   completion        │
│ - Updates DB        │
│ - Publishes next    │
│   task (if any)     │
└─────────────────────┘
```

## Files Created

### 1. **Event System**

- `internal/events/producer.go` - Kafka producer with TaskEvent & CompletionEvent
- `internal/events/consumer.go` - Kafka consumer for orchestrator

### 2. **Audio Handler**

- `internal/handler/audio.go` - Audio generation endpoints
  - `POST /audio/generate` - Trigger workflow
  - `GET /audio/status/:execution_id` - Check status

### 3. **Router**

- `internal/router/audio.go` - Audio route setup

### 4. **Services**

- `internal/service/instance.go` - Workflow instance management
- `internal/service/workflow.go` - Added `GetWorkflowByEvent()`

### 5. **Repositories**

- `internal/repositories/instance.go` - Instance CRUD operations
- `internal/repositories/wokrflow.go` - Added `GetByEvent()`

### 6. **Orchestrator State Machine**

- `internal/orchestrator/state_machine.go` - Processes completions & publishes next tasks

### 7. **Main Application**

- `cmd/orchestrator/main.go` - Wired everything together:
  - HTTP server for API
  - Kafka consumer for completions (runs in goroutine)
  - Graceful shutdown

## Workflow Flow

### Step 1: User Triggers Workflow

```bash
POST http://localhost:8080/audio/generate
Content-Type: application/json

{
  "text": "Hello world, generate audio for this",
  "voice": "default",
  "metadata": {
    "user_id": "123",
    "priority": "high"
  }
}
```

**Response (202 Accepted):**

```json
{
  "execution_id": "550e8400-e29b-41d4-a716-446655440000",
  "workflow": "audio_generation",
  "status": "PENDING",
  "message": "Audio generation workflow triggered successfully"
}
```

**What Happens:**

1. Creates `WorkflowInstance` in DB (status: PENDING, step: 0)
2. Publishes `extract_text` task to Kafka topic `task-queue`

---

### Step 2: Worker Processes Task

Worker consumes from `task-queue`:

```json
{
  "execution_id": "550e8400-...",
  "workflow_id": 2,
  "task_type": "extract_text",
  "step": 1,
  "input": {
    "text": "Hello world...",
    "voice": "default"
  }
}
```

Worker:

1. Extracts/validates text
2. Publishes completion to `task-completions`:

```json
{
  "execution_id": "550e8400-...",
  "workflow_id": 2,
  "task_type": "extract_text",
  "step": 1,
  "status": "success",
  "output": {
    "extracted_text": "Hello world...",
    "char_count": 25
  }
}
```

---

### Step 3: Orchestrator Processes Completion

Orchestrator consumer receives completion:

1. Updates DB: `current_step = 1, status = COMPLETED`
2. Checks workflow definition (3 steps total)
3. Determines next task: `generate_audio` (step 2)
4. Publishes to `task-queue`:

```json
{
  "execution_id": "550e8400-...",
  "workflow_id": 2,
  "task_type": "generate_audio",
  "step": 2,
  "input": {
    "extracted_text": "Hello world...",
    "voice": "default"
  }
}
```

---

### Step 4: Loop Until Complete

- **Step 2:** Worker generates audio → Publishes completion
- Orchestrator → Publishes `store_audio` (step 3)
- **Step 3:** Worker stores audio → Publishes completion
- Orchestrator → No more steps → Updates DB: `status = COMPLETED`

---

## Environment Setup

### Update `.env` file:

```properties
DATABASE_URL=postgresql://neondb_owner:npg_WF5dKUomxDN4@ep-young-meadow-adkxbzao-pooler.c-2.us-east-1.aws.neon.tech/neondb?sslmode=require&channel_binding=require
KAFKA_BROKERS=localhost:9092
LOG_LEVEL=info
LOG_FORMAT=json
PORT=8080
```

### Install Dependencies:

```bash
go get github.com/google/uuid
go mod tidy
```

---

## Kafka Topics Required

Create these topics in your Kafka setup:

```bash
# Task queue for workers
kafka-topics.sh --create --topic task-queue --bootstrap-server localhost:9092 --partitions 3 --replication-factor 1

# Completion events for orchestrator
kafka-topics.sh --create --topic task-completions --bootstrap-server localhost:9092 --partitions 3 --replication-factor 1
```

---

## API Endpoints

### 1. Trigger Audio Generation

```bash
POST /audio/generate
```

### 2. Check Workflow Status

```bash
GET /audio/status/:execution_id
```

**Example:**

```bash
curl http://localhost:8080/audio/status/550e8400-e29b-41d4-a716-446655440000
```

**Response:**

```json
{
  "execution_id": "550e8400-...",
  "workflow_id": 2,
  "status": "PENDING",
  "current_step": 1,
  "created_at": "2025-10-19T10:30:00Z",
  "updated_at": "2025-10-19T10:30:05Z"
}
```

### 3. Health Check

```bash
GET /health
```

### 4. Create Workflow (Admin)

```bash
POST /workflow/create
```

---

## Database Schema

The workflow instance table tracks execution:

```sql
SELECT * FROM workflow_instances WHERE execution_id = '550e8400...';
```

| Column       | Value                        |
| ------------ | ---------------------------- |
| execution_id | 550e8400-...                 |
| workflow_id  | 2                            |
| status       | PENDING / COMPLETED / FAILED |
| current_step | 0-3                          |
| variables    | JSON with input/output       |
| created_at   | Timestamp                    |
| updated_at   | Timestamp                    |

---

## Running the System

### 1. Start Kafka & Postgres (if using docker-compose):

```bash
docker-compose up -d
```

### 2. Start Orchestrator:

```bash
cd cmd/orchestrator
go run main.go
```

**Expected logs:**

```
INFO Database initialized successfully
INFO Kafka producer initialized successfully
INFO Kafka consumer initialized successfully
INFO Event producer and consumer initialized
INFO Repositories initialized
INFO Services initialized
INFO Orchestrator state machine initialized
INFO Handlers initialized
INFO Starting Kafka consumer for task completions
INFO Routes configured
INFO Starting HTTP server port=8080
INFO Orchestrator started, waiting for shutdown signal
```

### 3. Test the API:

```bash
curl -X POST http://localhost:8080/audio/generate \
  -H "Content-Type: application/json" \
  -d '{
    "text": "Test audio generation",
    "voice": "default"
  }'
```

---

## Next Steps

### 1. **Implement Workers** (Phase 5 from todo.md)

Create `cmd/worker/main.go` that:

- Consumes from `task-queue`
- Executes tasks based on `task_type`
- Publishes completions to `task-completions`

### 2. **Add Task Executors**

In `internal/workers/`:

- `extract_text.go` - Text extraction logic
- `generate_audio.go` - AI/TTS integration
- `store_audio.go` - S3/storage logic

### 3. **Error Handling**

- Retry logic for failed tasks
- Dead-letter queue for permanent failures
- Timeout handling

### 4. **Monitoring**

- Prometheus metrics
- Structured logging with trace IDs
- Dashboard for workflow visualization

---

## Key Features Implemented ✅

1. ✅ Event-driven architecture with Kafka
2. ✅ Orchestrator as entry point (validates & creates instance)
3. ✅ State machine to process completions
4. ✅ Automatic task progression (step 1 → 2 → 3)
5. ✅ Separate handlers per endpoint type
6. ✅ Graceful shutdown for both HTTP & Kafka consumer
7. ✅ Proper repository & service layers
8. ✅ Status tracking API

---

## Architecture Benefits

🎯 **Scalability:** Workers can scale independently  
🎯 **Reliability:** Kafka ensures message delivery  
🎯 **Observability:** Each step is tracked in DB  
🎯 **Flexibility:** Easy to add new workflow types  
🎯 **Fault Tolerance:** Failed tasks can be retried

---

**All 3 tasks completed! The system is ready for testing.** 🚀
