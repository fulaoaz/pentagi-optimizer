package artex

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"database/sql/driver"
	"encoding/binary"
	"encoding/json"
	"errors"
	"strconv"
	"time"
)

// SQLBindingStore requires a direct PostgreSQL connection or session-affine
// pool. The dedicated connection holds an advisory lock while journal writes
// autocommit, so an uncertain HTTP outcome cannot roll back its recovery row.
type SQLBindingStore struct{ db *sql.DB }

func NewSQLBindingStore(db *sql.DB) *SQLBindingStore { return &SQLBindingStore{db: db} }

type sqlBindingRepository struct {
	db interface {
		QueryRowContext(context.Context, string, ...any) *sql.Row
		ExecContext(context.Context, string, ...any) (sql.Result, error)
	}
	engine string
	uid    int64
	source string
}

const bindingColumns = `id,engine_key,user_id,source_ref,revision,remote_name,remote_id,spec,credential_fingerprint,pending_fingerprint,state,created_at,updated_at`

func scanBinding(row interface{ Scan(...any) error }) (*Binding, error) {
	var item Binding
	var raw []byte
	err := row.Scan(&item.ID, &item.EngineKey, &item.UserID, &item.SourceRef, &item.Revision, &item.RemoteName, &item.RemoteID, &raw, &item.CredentialFingerprint, &item.PendingFingerprint, &item.State, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrBindingMissing
	}
	if err != nil {
		return nil, ErrStoreUnavailable
	}
	if json.Unmarshal(raw, &item.Spec) != nil {
		return nil, ErrStoreUnavailable
	}
	return &item, nil
}

func (s *SQLBindingStore) WithSourceLock(ctx context.Context, engine string, uid int64, source string, fn func(BindingRepository) error) error {
	if s == nil || s.db == nil {
		return ErrStoreUnavailable
	}
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return ErrStoreUnavailable
	}
	sum := sha256.Sum256([]byte(engine + "\x00" + strconv.FormatInt(uid, 10) + "\x00" + source))
	lockID := int64(binary.BigEndian.Uint64(sum[:8]))
	defer func() {
		release, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		var unlocked bool
		if err := conn.QueryRowContext(release, "SELECT pg_advisory_unlock($1)", lockID).Scan(&unlocked); err != nil || !unlocked {
			// A pooled connection must never retain an invisible advisory lock.
			_ = conn.Raw(func(any) error { return driver.ErrBadConn })
		}
		_ = conn.Close()
	}()
	var ignored any
	if conn.QueryRowContext(ctx, "SELECT pg_advisory_lock($1)", lockID).Scan(&ignored) != nil {
		return ErrStoreUnavailable
	}
	return fn(&sqlBindingRepository{db: conn, engine: engine, uid: uid, source: source})
}

func (s *SQLBindingStore) Get(ctx context.Context, engine string, uid int64, id string) (*Binding, error) {
	if s == nil || s.db == nil {
		return nil, ErrStoreUnavailable
	}
	return scanBinding(s.db.QueryRowContext(ctx, "SELECT "+bindingColumns+" FROM engine_provider_bindings WHERE engine_key=$1 AND user_id=$2 AND id=$3", engine, uid, id))
}
func (s *SQLBindingStore) List(ctx context.Context, engine string, uid int64) ([]Binding, error) {
	if s == nil || s.db == nil {
		return nil, ErrStoreUnavailable
	}
	rows, err := s.db.QueryContext(ctx, "SELECT "+bindingColumns+" FROM engine_provider_bindings WHERE engine_key=$1 AND user_id=$2 ORDER BY created_at DESC", engine, uid)
	if err != nil {
		return nil, ErrStoreUnavailable
	}
	defer rows.Close()
	items := []Binding{}
	for rows.Next() {
		item, err := scanBinding(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	if rows.Err() != nil {
		return nil, ErrStoreUnavailable
	}
	return items, nil
}
func (r *sqlBindingRepository) FindRevision(ctx context.Context, revision string) (*Binding, error) {
	return scanBinding(r.db.QueryRowContext(ctx, "SELECT "+bindingColumns+" FROM engine_provider_bindings WHERE engine_key=$1 AND user_id=$2 AND source_ref=$3 AND revision=$4", r.engine, r.uid, r.source, revision))
}
func (r *sqlBindingRepository) Get(ctx context.Context, id string) (*Binding, error) {
	return scanBinding(r.db.QueryRowContext(ctx, "SELECT "+bindingColumns+" FROM engine_provider_bindings WHERE engine_key=$1 AND user_id=$2 AND source_ref=$3 AND id=$4", r.engine, r.uid, r.source, id))
}
func (r *sqlBindingRepository) Insert(ctx context.Context, item *Binding) error {
	spec, err := json.Marshal(item.Spec)
	if err != nil {
		return ErrInvalidRequest
	}
	err = r.db.QueryRowContext(ctx, `INSERT INTO engine_provider_bindings
 (id,engine_key,user_id,source_ref,revision,remote_name,remote_id,spec,credential_fingerprint,pending_fingerprint,state)
 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING created_at,updated_at`,
		item.ID, r.engine, r.uid, r.source, item.Revision, item.RemoteName, item.RemoteID, string(spec), item.CredentialFingerprint, item.PendingFingerprint, item.State).Scan(&item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return ErrStoreUnavailable
	}
	return nil
}
func (r *sqlBindingRepository) Save(ctx context.Context, item *Binding) error {
	err := r.db.QueryRowContext(ctx, `UPDATE engine_provider_bindings SET remote_id=$5,credential_fingerprint=$6,pending_fingerprint=$7,state=$8,updated_at=now()
 WHERE engine_key=$1 AND user_id=$2 AND source_ref=$3 AND id=$4 RETURNING updated_at`,
		r.engine, r.uid, r.source, item.ID, item.RemoteID, item.CredentialFingerprint, item.PendingFingerprint, item.State).Scan(&item.UpdatedAt)
	if err != nil {
		return ErrStoreUnavailable
	}
	return nil
}
