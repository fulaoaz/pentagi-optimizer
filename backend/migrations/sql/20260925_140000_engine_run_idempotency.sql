-- +goose Up
CREATE UNIQUE INDEX engine_runs_idempotency
    ON engine_runs(engine_key, user_id, (request->>'correlation_id'))
    WHERE request->>'correlation_id' <> '';

-- +goose Down
DROP INDEX engine_runs_idempotency;
