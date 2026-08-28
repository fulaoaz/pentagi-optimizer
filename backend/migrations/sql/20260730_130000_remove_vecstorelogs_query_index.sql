-- +goose Up
-- +goose StatementBegin
-- Vector store log queries are unbounded text. A B-tree index rejects values
-- above PostgreSQL's per-index-entry limit and prevents the log record itself
-- from being stored.
DROP INDEX IF EXISTS vecstorelogs_query_idx;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Deliberately left without a replacement: query values are arbitrary long
-- tool payloads and are not used as an equality lookup key by the application.
-- +goose StatementEnd
