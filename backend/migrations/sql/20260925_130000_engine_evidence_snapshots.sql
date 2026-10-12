-- +goose Up
CREATE TABLE engine_evidence_snapshots (
    id TEXT PRIMARY KEY,
    sequence BIGSERIAL UNIQUE NOT NULL,
    engine_key TEXT NOT NULL,
    run_id TEXT NOT NULL REFERENCES engine_runs(id),
    user_id BIGINT NOT NULL,
    remote_task_id TEXT NOT NULL,
    remote_record_id TEXT NOT NULL,
    parent_record_id TEXT NOT NULL DEFAULT '',
    representation TEXT NOT NULL CHECK (representation IN ('list','detail')),
    kind TEXT NOT NULL CHECK (kind IN ('asset','finding','activity','retest','report')),
    version INTEGER NOT NULL CHECK (version > 0),
    content TEXT NOT NULL CHECK (jsonb_typeof(content::jsonb) = 'object'),
    content_sha256 TEXT NOT NULL,
    redacted BOOLEAN NOT NULL DEFAULT false,
    captured_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(engine_key, run_id, remote_record_id, kind, representation, version)
);
CREATE INDEX engine_evidence_owner ON engine_evidence_snapshots(engine_key, user_id, run_id, kind, captured_at DESC);
CREATE INDEX engine_evidence_pagination ON engine_evidence_snapshots(engine_key, run_id, sequence DESC);

-- +goose Down
DROP TABLE engine_evidence_snapshots;
