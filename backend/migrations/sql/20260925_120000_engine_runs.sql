-- +goose Up
CREATE TABLE engine_runs (
    id TEXT PRIMARY KEY,
    engine_key TEXT NOT NULL,
    user_id BIGINT NOT NULL,
    engine_task_id TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL,
    binding_id TEXT NOT NULL,
    binding_revision TEXT NOT NULL,
    submit_state TEXT NOT NULL CHECK (submit_state IN ('pending','created','uncertain','failed')),
    raw_status TEXT NOT NULL DEFAULT '',
    normalized_status TEXT NOT NULL DEFAULT 'pending',
    request JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_sync_at TIMESTAMPTZ
);
CREATE INDEX engine_runs_owner ON engine_runs(engine_key, user_id, updated_at DESC);
CREATE UNIQUE INDEX engine_runs_remote ON engine_runs(engine_key, user_id, engine_task_id) WHERE engine_task_id <> '';

-- +goose Down
DROP TABLE engine_runs;
