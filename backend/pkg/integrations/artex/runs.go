package artex

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type RunRequest struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	Goal            string `json:"goal"`
	TimeoutSeconds  int    `json:"timeout_seconds,omitempty"`
	BindingID       string `json:"binding_id"`
	BindingRevision string `json:"binding_revision"`
	CorrelationID   string `json:"correlation_id"`
}

type Run struct {
	ID              string     `json:"id"`
	Engine          string     `json:"engine"`
	UserID          int64      `json:"user_id"`
	EngineTaskID    string     `json:"engine_task_id,omitempty"`
	Name            string     `json:"name"`
	BindingID       string     `json:"binding_id"`
	BindingRevision string     `json:"binding_revision"`
	SubmitState     string     `json:"submit_state"`
	RawStatus       string     `json:"raw_status,omitempty"`
	Status          string     `json:"status"`
	Request         RunRequest `json:"request"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	LastSyncAt      *time.Time `json:"last_sync_at,omitempty"`
}

type UnifiedRun struct {
	ID              string    `json:"id"`
	Engine          string    `json:"engine"`
	EngineTaskID    string    `json:"engineTaskId,omitempty"`
	Title           string    `json:"title"`
	Status          string    `json:"status"`
	RawStatus       string    `json:"rawStatus,omitempty"`
	SubmitState     string    `json:"submitState,omitempty"`
	Model           string    `json:"model,omitempty"`
	BindingID       string    `json:"bindingId,omitempty"`
	BindingRevision string    `json:"bindingRevision,omitempty"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
	DetailPath      string    `json:"detailPath"`
}

type TaskSummary struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func ParseTaskSummaries(raw json.RawMessage) ([]TaskSummary, error) {
	var envelope struct {
		Tasks []TaskSummary `json:"tasks"`
	}
	if err := json.Unmarshal(raw, &envelope); err == nil && envelope.Tasks != nil {
		return envelope.Tasks, nil
	}
	var list []TaskSummary
	if err := json.Unmarshal(raw, &list); err != nil {
		return nil, ErrResponse
	}
	return list, nil
}

type SQLRunStore struct{ db *sql.DB }

var (
	ErrRunNotFound   = errors.New("engine run not found")
	ErrRunUnresolved = errors.New("remote run could not be uniquely reconciled")
	ErrRunAmbiguous  = errors.New("multiple remote tasks matched the run correlation")
)

func NewSQLRunStore(db *sql.DB) *SQLRunStore { return &SQLRunStore{db: db} }

func CorrelationMarker(id string) string { return "[pentagi-run:" + strings.TrimSpace(id) + "]" }

func HasCorrelation(text, id string) bool { return strings.Contains(text, CorrelationMarker(id)) }

func (s *SQLRunStore) CreatePending(ctx context.Context, uid int64, req RunRequest) (*Run, error) {
	if s == nil || s.db == nil || uid <= 0 || strings.TrimSpace(req.BindingID) == "" {
		return nil, ErrStoreUnavailable
	}
	id := uuid.NewString()
	if req.CorrelationID == "" {
		req.CorrelationID = id
	}
	raw, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	const q = `INSERT INTO engine_runs (id,engine_key,user_id,name,binding_id,binding_revision,submit_state,request) VALUES ($1,'artex',$2,$3,$4,$5,'pending',$6) RETURNING created_at,updated_at`
	var r Run
	r.ID = id
	r.Engine = "artex"
	r.UserID = uid
	r.Name = req.Name
	r.BindingID = req.BindingID
	r.BindingRevision = req.BindingRevision
	r.SubmitState = "pending"
	r.Status = "pending"
	r.Request = req
	if err := s.db.QueryRowContext(ctx, q, id, uid, req.Name, req.BindingID, req.BindingRevision, raw).Scan(&r.CreatedAt, &r.UpdatedAt); err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *SQLRunStore) MarkCreated(ctx context.Context, id, remote string) error {
	if s == nil || s.db == nil {
		return ErrStoreUnavailable
	}
	if strings.TrimSpace(id) == "" || strings.TrimSpace(remote) == "" {
		return ErrInvalidRequest
	}
	_, err := s.db.ExecContext(ctx, `UPDATE engine_runs SET engine_task_id=$2,submit_state='created',normalized_status='queued',updated_at=now(),last_sync_at=now() WHERE id=$1`, id, remote)
	return err
}

func (s *SQLRunStore) MarkFailed(ctx context.Context, id string) error {
	if s == nil || s.db == nil {
		return ErrStoreUnavailable
	}
	_, err := s.db.ExecContext(ctx, `UPDATE engine_runs SET submit_state='failed',normalized_status='failed',updated_at=now(),last_sync_at=now() WHERE id=$1`, id)
	return err
}
func (s *SQLRunStore) MarkUncertain(ctx context.Context, id string) error {
	if s == nil || s.db == nil {
		return ErrStoreUnavailable
	}
	_, err := s.db.ExecContext(ctx, `UPDATE engine_runs SET submit_state='uncertain',normalized_status='sync_unknown',updated_at=now() WHERE id=$1`, id)
	return err
}

func (s *SQLRunStore) UpdateRemote(ctx context.Context, remoteID, rawStatus, status string) error {
	if s == nil || s.db == nil {
		return ErrStoreUnavailable
	}
	_, err := s.db.ExecContext(ctx, `UPDATE engine_runs SET raw_status=$2,normalized_status=$3,updated_at=now(),last_sync_at=now() WHERE engine_key='artex' AND engine_task_id=$1`, remoteID, rawStatus, status)
	return err
}

func (s *SQLRunStore) Get(ctx context.Context, uid int64, id string) (*Run, error) {
	if s == nil || s.db == nil {
		return nil, ErrStoreUnavailable
	}
	var r Run
	var raw []byte
	err := s.db.QueryRowContext(ctx, `SELECT id,user_id,engine_task_id,name,binding_id,binding_revision,submit_state,raw_status,normalized_status,request,created_at,updated_at,last_sync_at FROM engine_runs WHERE engine_key='artex' AND user_id=$1 AND id=$2`, uid, id).Scan(&r.ID, &r.UserID, &r.EngineTaskID, &r.Name, &r.BindingID, &r.BindingRevision, &r.SubmitState, &r.RawStatus, &r.Status, &raw, &r.CreatedAt, &r.UpdatedAt, &r.LastSyncAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrRunNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(raw, &r.Request); err != nil {
		return nil, errors.New("invalid engine run request")
	}
	r.Engine = "artex"
	return &r, nil
}

func (s *SQLRunStore) FindByRemote(ctx context.Context, remoteTaskID string) (*Run, error) {
	if s == nil || s.db == nil {
		return nil, ErrStoreUnavailable
	}
	var r Run
	var raw []byte
	err := s.db.QueryRowContext(ctx, `SELECT id,user_id,engine_task_id,name,binding_id,binding_revision,submit_state,raw_status,normalized_status,request,created_at,updated_at,last_sync_at FROM engine_runs WHERE engine_key='artex' AND engine_task_id=$1 ORDER BY updated_at DESC LIMIT 1`, remoteTaskID).Scan(&r.ID, &r.UserID, &r.EngineTaskID, &r.Name, &r.BindingID, &r.BindingRevision, &r.SubmitState, &r.RawStatus, &r.Status, &raw, &r.CreatedAt, &r.UpdatedAt, &r.LastSyncAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrRunNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(raw, &r.Request); err != nil {
		return nil, errors.New("invalid engine run request")
	}
	r.Engine = "artex"
	return &r, nil
}

func (s *SQLRunStore) FindByCorrelation(ctx context.Context, uid int64, correlation string) (*Run, error) {
	if s == nil || s.db == nil || uid <= 0 || strings.TrimSpace(correlation) == "" {
		return nil, ErrStoreUnavailable
	}
	var r Run
	var raw []byte
	err := s.db.QueryRowContext(ctx, `SELECT id,user_id,engine_task_id,name,binding_id,binding_revision,submit_state,raw_status,normalized_status,request,created_at,updated_at,last_sync_at FROM engine_runs WHERE engine_key='artex' AND user_id=$1 AND request->>'correlation_id'=$2 ORDER BY updated_at DESC LIMIT 1`, uid, correlation).Scan(&r.ID, &r.UserID, &r.EngineTaskID, &r.Name, &r.BindingID, &r.BindingRevision, &r.SubmitState, &r.RawStatus, &r.Status, &raw, &r.CreatedAt, &r.UpdatedAt, &r.LastSyncAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrRunNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(raw, &r.Request); err != nil {
		return nil, errors.New("invalid engine run request")
	}
	r.Engine = "artex"
	return &r, nil
}

func (s *SQLRunStore) FindByRecord(ctx context.Context, remoteRecordID, kind string) (*Run, error) {
	if s == nil || s.db == nil {
		return nil, ErrStoreUnavailable
	}
	var r Run
	var raw []byte
	err := s.db.QueryRowContext(ctx, `SELECT r.id,r.user_id,r.engine_task_id,r.name,r.binding_id,r.binding_revision,r.submit_state,r.raw_status,r.normalized_status,r.request,r.created_at,r.updated_at,r.last_sync_at FROM engine_runs r INNER JOIN engine_evidence_snapshots e ON e.run_id=r.id WHERE e.engine_key='artex' AND e.remote_record_id=$1 AND e.kind=$2 ORDER BY e.captured_at DESC LIMIT 1`, remoteRecordID, kind).Scan(&r.ID, &r.UserID, &r.EngineTaskID, &r.Name, &r.BindingID, &r.BindingRevision, &r.SubmitState, &r.RawStatus, &r.Status, &raw, &r.CreatedAt, &r.UpdatedAt, &r.LastSyncAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrRunNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(raw, &r.Request); err != nil {
		return nil, errors.New("invalid engine run request")
	}
	r.Engine = "artex"
	return &r, nil
}

func NormalizeTaskStatus(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "queued", "pending", "created", "planning":
		return "queued"
	case "running", "executing", "in_progress":
		return "running"
	case "paused", "pause":
		return "paused"
	case "completed", "complete", "done", "success":
		return "completed"
	case "timeout", "timed_out":
		return "timeout"
	case "failed", "failure", "error":
		return "failed"
	default:
		return "unknown"
	}
}
func (s *SQLRunStore) List(ctx context.Context, uid int64) ([]Run, error) {
	return s.list(ctx, &uid)
}

func (s *SQLRunStore) ListAll(ctx context.Context) ([]Run, error) { return s.list(ctx, nil) }

func (s *SQLRunStore) list(ctx context.Context, uid *int64) ([]Run, error) {
	if s == nil || s.db == nil {
		return nil, ErrStoreUnavailable
	}
	query := `SELECT id,user_id,engine_task_id,name,binding_id,binding_revision,submit_state,raw_status,normalized_status,request,created_at,updated_at,last_sync_at FROM engine_runs WHERE engine_key='artex'`
	args := []any{}
	if uid != nil {
		query += ` AND user_id=$1`
		args = append(args, *uid)
	}
	query += ` ORDER BY updated_at DESC`
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Run{}
	for rows.Next() {
		var r Run
		var raw []byte
		if err := rows.Scan(&r.ID, &r.UserID, &r.EngineTaskID, &r.Name, &r.BindingID, &r.BindingRevision, &r.SubmitState, &r.RawStatus, &r.Status, &raw, &r.CreatedAt, &r.UpdatedAt, &r.LastSyncAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(raw, &r.Request); err != nil {
			return nil, errors.New("invalid engine run request")
		}
		r.Engine = "artex"
		out = append(out, r)
	}
	return out, rows.Err()
}
