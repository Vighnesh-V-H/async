# Postman Testing Guide for Audio Generation API

## Prerequisites

1. Make sure the orchestrator is running: `task run:orc`
2. Kafka should be running (docker-compose up -d)
3. Base URL: `http://localhost:8080`
4. **Workflow Definition** must already exist in the database (pre-seeded):
   ```
   Workflow ID: 2
   Name: "audio_generation"
   Event: "generate_audio"
   Steps: 3
   ```

---

## 🎯 Important Concept

### Workflow vs Workflow Instance

- **Workflow Definition** (ONE per workflow type)

  - Stored once in the `workflows` table
  - Defines: name, event, steps, handler_url
  - Example: `audio_generation` workflow with 3 steps
  - **Pre-configured in your database**

- **Workflow Instance** (ONE per user request)
  - Created automatically when you call `/audio/generate`
  - Stored in `workflow_instances` table
  - Tracks: execution_id, status, current_step, input/output
  - Example: User request generates instance with unique execution_id

**You DON'T manually create instances - the API does it for you!**

---

## API Endpoints

### 1. Health Check (GET)

**Purpose:** Verify the orchestrator is running

**Endpoint:** `GET http://localhost:8080/health`

**Headers:** None required

**Response (200 OK):**

```json
{
  "service": "orchestrator",
  "status": "ok"
}
```

---

### 2. Trigger Audio Generation (POST) - Main Endpoint

**Purpose:** Trigger the audio generation workflow (automatically creates a workflow instance)

**Endpoint:** `POST http://localhost:8080/audio/generate`

**What Happens Behind the Scenes:**

1. ✅ Validates your request
2. ✅ Fetches the `audio_generation` workflow definition from DB
3. ✅ **Creates a NEW workflow instance** with unique execution_id
4. ✅ Saves instance to DB (status: PENDING, step: 0)
5. ✅ Publishes first task (`extract_text`) to Kafka
6. ✅ Returns execution_id for tracking

**Headers:**

```
Content-Type: application/json
```

**Body (raw JSON):**

```json
{
  "text": "Hello world, this is a test for audio generation using AI",
  "voice": "default",
  "metadata": {
    "user_id": "user_123",
    "priority": "high",
    "language": "en-US"
  }
}
```

**Minimal Body (only text is required):**

```json
{
  "text": "Simple test message"
}
```

**Response (202 Accepted):**

```json
{
  "execution_id": "550e8400-e29b-41d4-a716-446655440000",
  "message": "Audio generation workflow triggered successfully",
  "status": "PENDING",
  "workflow": "audio_generation"
}
```

**Error Responses:**

**400 Bad Request** (missing required field):

```json
{
  "error": "Key: 'GenerateAudioRequest.Text' Error:Field validation for 'Text' failed on the 'required' tag"
}
```

**500 Internal Server Error** (workflow not found):

```json
{
  "error": "workflow not found"
}
```

_Note: This means the `audio_generation` workflow doesn't exist in your database. See "Admin Endpoints" section below._

---

### 3. Check Workflow Instance Status (GET)

**Purpose:** Check the current status of a workflow instance execution

**Endpoint:** `GET http://localhost:8080/audio/status/:execution_id`

**Example:** `GET http://localhost:8080/audio/status/550e8400-e29b-41d4-a716-446655440000`

**Headers:** None required

**What This Returns:**

- Current status of the **workflow instance**
- Which step it's on (0-3)
- When it was created/updated
- The workflow definition ID it belongs to

**Response (200 OK):**

```json
{
  "created_at": "2025-10-19T18:50:00Z",
  "current_step": 1,
  "execution_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "PENDING",
  "updated_at": "2025-10-19T18:50:05Z",
  "workflow_id": 2
}
```

**Status Values:**

- `PENDING` - Workflow instance created, task(s) in progress
- `COMPLETED` - All 3 steps completed successfully
- `FAILED` - Workflow failed at some step

**Error Response (404 Not Found):**

```json
{
  "error": "execution not found"
}
```

---

## 🔧 Admin Endpoints (One-Time Setup)

### Create Workflow Definition (POST)

**Purpose:** Create a new workflow definition (only needed once per workflow type)

**⚠️ Important:** You typically do this ONCE during initial setup, not per user request!

**Endpoint:** `POST http://localhost:8080/workflow/create`

**Headers:**

```
Content-Type: application/json
```

**Body (raw JSON):**

```json
{
  "name": "audio_generation",
  "event": "generate_audio",
  "message": "audio_generation_pipeline",
  "handler_url": "http://localhost:8080/audio/generate",
  "steps": 3
}
```

**Response (201 Created):**

```json
{
  "id": 2,
  "name": "audio_generation",
  "event": "generate_audio",
  "status": "active",
  "message": "audio_generation_pipeline",
  "payload": null,
  "steps": 3,
  "handler_url": "http://localhost:8080/audio/generate",
  "created_at": "2025-10-19T18:43:17.063176Z",
  "updated_at": "2025-10-19T18:43:17.063176Z"
}
```

**When to Use:**

- First time setup
- Creating new workflow types
- You already have this in your database (ID: 2)

---

## Postman Collection Setup

### Step-by-Step Instructions:

#### 1. **Create a New Collection**

- Open Postman
- Click **"New"** → **"Collection"**
- Name it: `Audio Generation API`
- Description: `Event-driven audio workflow orchestrator`

#### 2. **Set Collection Variables**

- Click on your collection
- Go to **"Variables"** tab
- Add these variables:

| Variable       | Initial Value           | Current Value           |
| -------------- | ----------------------- | ----------------------- |
| `base_url`     | `http://localhost:8080` | `http://localhost:8080` |
| `execution_id` | (leave empty)           | (leave empty)           |

#### 3. **Add Requests**

##### Request 1: Health Check

- **Name:** `Health Check`
- **Method:** `GET`
- **URL:** `{{base_url}}/health`
- **Tests Tab** (add this script):

```javascript
pm.test("Status code is 200", function () {
  pm.response.to.have.status(200);
});

pm.test("Response has correct structure", function () {
  var jsonData = pm.response.json();
  pm.expect(jsonData).to.have.property("status");
  pm.expect(jsonData.status).to.eql("ok");
});
```

##### Request 2: Trigger Audio Generation

- **Name:** `Trigger Audio Generation`
- **Method:** `POST`
- **URL:** `{{base_url}}/audio/generate`
- **Headers:**
  - `Content-Type: application/json`
- **Body:** (raw JSON)

```json
{
  "text": "Generate audio for this test message",
  "voice": "default",
  "metadata": {
    "user_id": "test_user_123",
    "priority": "high"
  }
}
```

- **Tests Tab:**

```javascript
pm.test("Status code is 202", function () {
  pm.response.to.have.status(202);
});

pm.test("Response contains execution_id", function () {
  var jsonData = pm.response.json();
  pm.expect(jsonData).to.have.property("execution_id");
  pm.expect(jsonData.status).to.eql("PENDING");

  // Save execution_id for next request
  pm.collectionVariables.set("execution_id", jsonData.execution_id);
});

pm.test("Workflow name is correct", function () {
  var jsonData = pm.response.json();
  pm.expect(jsonData.workflow).to.eql("audio_generation");
});

console.log("✅ Workflow instance created!");
console.log("Execution ID:", pm.response.json().execution_id);
```

##### Request 3: Check Status

- **Name:** `Check Workflow Status`
- **Method:** `GET`
- **URL:** `{{base_url}}/audio/status/{{execution_id}}`
- **Tests Tab:**

```javascript
pm.test("Status code is 200", function () {
  pm.response.to.have.status(200);
});

pm.test("Response has workflow details", function () {
  var jsonData = pm.response.json();
  pm.expect(jsonData).to.have.property("execution_id");
  pm.expect(jsonData).to.have.property("status");
  pm.expect(jsonData).to.have.property("current_step");
  pm.expect(jsonData).to.have.property("workflow_id");
});

console.log("Status:", pm.response.json().status);
console.log("Current Step:", pm.response.json().current_step);
```

##### Request 4: [ADMIN] Create Workflow Definition

- **Name:** `[ADMIN] Create Workflow Definition`
- **Method:** `POST`
- **URL:** `{{base_url}}/workflow/create`
- **Headers:**
  - `Content-Type: application/json`
- **Body:** (raw JSON)

```json
{
  "name": "audio_generation",
  "event": "generate_audio",
  "message": "audio_generation_pipeline",
  "handler_url": "{{base_url}}/audio/generate",
  "steps": 3
}
```

- **Tests Tab:**

```javascript
pm.test("Status code is 201", function () {
  pm.response.to.have.status(201);
});

pm.test("Workflow definition created", function () {
  var jsonData = pm.response.json();
  pm.expect(jsonData).to.have.property("id");
  pm.expect(jsonData.name).to.eql("audio_generation");
  pm.expect(jsonData.status).to.eql("active");
});

console.log("⚠️ Only run this ONCE for initial setup!");
```

---

## Testing Workflow

### Complete Test Sequence:

1. **First Time Setup (ONLY ONCE):**

   ```
   1. Health Check → Verify orchestrator is running
   2. [ADMIN] Create Workflow Definition → Creates workflow in DB (if not exists)
   ```

2. **Normal Testing Flow (Every Time):**

   ```
   1. Health Check → Verify system is up
   2. Trigger Audio Generation → Creates workflow instance, returns execution_id
   3. Check Status → Monitor instance progress (run multiple times)
   ```

3. **Expected Instance Flow:**
   ```
   POST /audio/generate (creates instance)
   ↓
   Response: execution_id = "abc-123..."
   ↓
   Database: WorkflowInstance created (Status: PENDING, Step: 0)
   ↓
   Kafka: Task "extract_text" published
   ↓
   GET /audio/status/abc-123 → Status: PENDING, Step: 0
   ↓
   (Worker processes extract_text)
   ↓
   Database: Instance updated (Step: 1)
   ↓
   Kafka: Task "generate_audio" published
   ↓
   GET /audio/status/abc-123 → Status: PENDING, Step: 1
   ↓
   (Worker processes generate_audio)
   ↓
   Database: Instance updated (Step: 2)
   ↓
   Kafka: Task "store_audio" published
   ↓
   GET /audio/status/abc-123 → Status: PENDING, Step: 2
   ↓
   (Worker processes store_audio)
   ↓
   Database: Instance updated (Status: COMPLETED, Step: 3)
   ↓
   GET /audio/status/abc-123 → Status: COMPLETED, Step: 3
   ```

---

## Import Ready Postman Collection (JSON)

Save this as `audio-generation-api.postman_collection.json`:

```json
{
  "info": {
    "name": "Audio Generation API",
    "description": "Event-driven audio workflow orchestrator",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "variable": [
    {
      "key": "base_url",
      "value": "http://localhost:8080",
      "type": "string"
    },
    {
      "key": "execution_id",
      "value": "",
      "type": "string"
    }
  ],
  "item": [
    {
      "name": "Health Check",
      "request": {
        "method": "GET",
        "header": [],
        "url": {
          "raw": "{{base_url}}/health",
          "host": ["{{base_url}}"],
          "path": ["health"]
        }
      }
    },
    {
      "name": "Create Workflow",
      "request": {
        "method": "POST",
        "header": [
          {
            "key": "Content-Type",
            "value": "application/json"
          }
        ],
        "body": {
          "mode": "raw",
          "raw": "{\n  \"name\": \"audio_generation\",\n  \"event\": \"generate_audio\",\n  \"message\": \"audio_generation_pipeline\",\n  \"handler_url\": \"{{base_url}}/audio/generate\",\n  \"steps\": 3\n}"
        },
        "url": {
          "raw": "{{base_url}}/workflow/create",
          "host": ["{{base_url}}"],
          "path": ["workflow", "create"]
        }
      }
    },
    {
      "name": "Trigger Audio Generation",
      "request": {
        "method": "POST",
        "header": [
          {
            "key": "Content-Type",
            "value": "application/json"
          }
        ],
        "body": {
          "mode": "raw",
          "raw": "{\n  \"text\": \"Generate audio for this test message\",\n  \"voice\": \"default\",\n  \"metadata\": {\n    \"user_id\": \"test_user_123\",\n    \"priority\": \"high\"\n  }\n}"
        },
        "url": {
          "raw": "{{base_url}}/audio/generate",
          "host": ["{{base_url}}"],
          "path": ["audio", "generate"]
        }
      },
      "event": [
        {
          "listen": "test",
          "script": {
            "exec": [
              "pm.test(\"Status code is 202\", function () {",
              "    pm.response.to.have.status(202);",
              "});",
              "",
              "pm.test(\"Response contains execution_id\", function () {",
              "    var jsonData = pm.response.json();",
              "    pm.expect(jsonData).to.have.property('execution_id');",
              "    pm.collectionVariables.set(\"execution_id\", jsonData.execution_id);",
              "});"
            ],
            "type": "text/javascript"
          }
        }
      ]
    },
    {
      "name": "Check Workflow Status",
      "request": {
        "method": "GET",
        "header": [],
        "url": {
          "raw": "{{base_url}}/audio/status/{{execution_id}}",
          "host": ["{{base_url}}"],
          "path": ["audio", "status", "{{execution_id}}"]
        }
      }
    }
  ]
}
```

**To Import:**

1. Open Postman
2. Click **"Import"** button
3. Drag and drop the JSON file or paste the JSON content
4. Click **"Import"**

---

## Quick Test Commands (Alternative - cURL)

If you prefer testing from terminal:

```bash
# 1. Health Check
curl http://localhost:8080/health

# 2. Create Workflow
curl -X POST http://localhost:8080/workflow/create \
  -H "Content-Type: application/json" \
  -d '{
    "name": "audio_generation",
    "event": "generate_audio",
    "message": "audio_generation_pipeline",
    "handler_url": "http://localhost:8080/audio/generate",
    "steps": 3
  }'

# 3. Trigger Audio Generation
curl -X POST http://localhost:8080/audio/generate \
  -H "Content-Type: application/json" \
  -d '{
    "text": "Test audio generation",
    "voice": "default"
  }'

# 4. Check Status (replace with your execution_id)
curl http://localhost:8080/audio/status/YOUR-EXECUTION-ID-HERE
```

---

## Troubleshooting

### Issue: "workflow not found"

**Solution:** Run the "Create Workflow" request first

### Issue: "execution not found"

**Solution:**

- Make sure you copied the correct `execution_id` from the trigger response
- Check if the variable `{{execution_id}}` is set in Postman

### Issue: Status stays at PENDING

**Cause:** Workers are not running yet
**Note:** This is expected! Workers will be implemented in Phase 5. For now, you'll see the workflow stuck at PENDING status, which proves the orchestrator is working correctly.

---

## What You're Testing

✅ **API Layer:** HTTP endpoints are working  
✅ **Database:**

- Workflow definitions (pre-seeded)
- **Workflow instances** (created per request)
  ✅ **Kafka Producer:** Tasks are being published to Kafka  
  ✅ **Orchestrator Logic:**
- Request validation
- **Instance creation** (automatic)
- Event publishing

🔄 **Not Yet Testable:** Task completion flow (requires workers to be implemented)

---

## Key Takeaways

### What Happens When You Call `/audio/generate`:

1. ✅ Finds the `audio_generation` **workflow definition** (already in DB)
2. ✅ **Creates a NEW workflow instance** with unique `execution_id`
3. ✅ Stores instance in `workflow_instances` table
4. ✅ Publishes first task to Kafka
5. ✅ Returns `execution_id` for tracking

### You DON'T Need To:

- ❌ Manually create workflow instances
- ❌ Call `/workflow/create` for every request
- ❌ Manage instance IDs yourself

### You DO Need To:

- ✅ Call `/audio/generate` for each audio generation request
- ✅ Use the returned `execution_id` to check status
- ✅ Have the workflow definition in DB (one-time setup)

---

Happy Testing! 🚀
