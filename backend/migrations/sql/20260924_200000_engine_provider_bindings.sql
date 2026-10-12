-- +goose Up
CREATE TABLE engine_provider_bindings (
    id TEXT PRIMARY KEY,
    engine_key TEXT NOT NULL,
    user_id BIGINT NOT NULL,
    source_ref TEXT NOT NULL,
    revision TEXT NOT NULL,
    remote_name TEXT NOT NULL,
    remote_id BIGINT NOT NULL DEFAULT 0,
    spec JSONB NOT NULL,
    credential_fingerprint TEXT NOT NULL,
    pending_fingerprint TEXT NOT NULL DEFAULT '',
    state TEXT NOT NULL CHECK (state IN ('pending','ready','uncertain','rotating','revoked')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(engine_key,user_id,source_ref,revision),
    UNIQUE(engine_key,remote_name)
);
-- Intentionally retain records after provider/user deletion for credential
-- cleanup and execution provenance. Authorization always checks the live user.
CREATE INDEX engine_provider_bindings_owner ON engine_provider_bindings(engine_key,user_id);

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN
    IF EXISTS (SELECT 1 FROM engine_provider_bindings) THEN
        RAISE EXCEPTION 'engine provider bindings exist: export/revoke mappings before removing this schema';
    END IF;
END $$;
-- +goose StatementEnd
DROP TABLE engine_provider_bindings;
