-- +goose Up
-- +goose StatementBegin

-- Create workflows table
CREATE TABLE IF NOT EXISTS workflows (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    event VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL,
    message VARCHAR(500),
    payload TEXT,
    steps SMALLINT NOT NULL,
    handler_url VARCHAR(500),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_workflows_event ON workflows(event);

-- Create workflow_instances table
CREATE TABLE IF NOT EXISTS workflow_instances (
    id SERIAL PRIMARY KEY,
    workflow_id INTEGER NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    execution_id VARCHAR(100) UNIQUE NOT NULL,
    status VARCHAR(50) NOT NULL,
    variables JSONB,
    current_step SMALLINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_workflow_instances_workflow_id ON workflow_instances(workflow_id);
CREATE INDEX IF NOT EXISTS idx_workflow_instances_status ON workflow_instances(status);

-- Create tasks table
CREATE TABLE IF NOT EXISTS tasks (
    id SERIAL PRIMARY KEY,
    instance_id INTEGER NOT NULL REFERENCES workflow_instances(id) ON DELETE CASCADE,
    step_id SMALLINT NOT NULL,
    type VARCHAR(50) NOT NULL,
    payload JSONB,
    output JSONB,
    status VARCHAR(50) NOT NULL,
    retries SMALLINT NOT NULL DEFAULT 0,
    timeout_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tasks_instance_id ON tasks(instance_id);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);

-- Create history_entries table
CREATE TABLE IF NOT EXISTS history_entries (
    id SERIAL PRIMARY KEY,
    instance_id INTEGER NOT NULL REFERENCES workflow_instances(id) ON DELETE CASCADE,
    event VARCHAR(100) NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    data JSONB
);

CREATE INDEX IF NOT EXISTS idx_history_entries_instance_id ON history_entries(instance_id);

-- Create workflow_registries table
CREATE TABLE IF NOT EXISTS workflow_registries (
    id SERIAL PRIMARY KEY,
    workflow_id INTEGER NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    trigger VARCHAR(255) UNIQUE NOT NULL,
    handler_url VARCHAR(500) NOT NULL,
    version VARCHAR(20) NOT NULL,
    lease_expires_at TIMESTAMP,
    last_heartbeat TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_workflow_registries_workflow_id ON workflow_registries(workflow_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop tables in reverse order (respecting foreign key constraints)
DROP TABLE IF EXISTS workflow_registries;
DROP TABLE IF EXISTS history_entries;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS workflow_instances;
DROP TABLE IF EXISTS workflows;

-- +goose StatementEnd
