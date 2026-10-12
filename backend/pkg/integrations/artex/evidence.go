package artex

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

const MaxEvidenceBytes = 1 << 20

var ErrSnapshotNotFound = errors.New("evidence snapshot not found")

type EvidenceSnapshot struct {
	ID             string          `json:"id"`
	Sequence       int64           `json:"sequence,string"`
	Engine         string          `json:"engine"`
	RunID          string          `json:"runId"`
	RemoteTaskID   string          `json:"remoteTaskId"`
	RemoteRecordID string          `json:"remoteRecordId"`
	ParentRecordID string          `json:"parentRecordId,omitempty"`
	Representation string          `json:"representation"`
	Kind           string          `json:"kind"`
	Version        int             `json:"version"`
	Content        json.RawMessage `json:"content,omitempty"`
	ContentSHA256  string          `json:"contentSha256"`
	Redacted       bool            `json:"redacted"`
	CapturedAt     time.Time       `json:"capturedAt"`
}

type EvidenceRecord struct {
	ID             string
	ParentID       string
	Kind           string
	Representation string
	Content        json.RawMessage
}

type SnapshotPage struct {
	Items      []EvidenceSnapshot `json:"items"`
	NextCursor string             `json:"nextCursor,omitempty"`
	HasMore    bool               `json:"hasMore"`
}

func validEvidenceKind(kind string) bool {
	switch kind {
	case "asset", "finding", "activity", "retest", "report":
		return true
	}
	return false
}

// CanonicalEvidence retains numeric precision and hashes the exact bytes stored.
// Credential fields are excluded from the product archive; originals stay in ARTEX.
func CanonicalEvidence(raw json.RawMessage) (json.RawMessage, bool, error) {
	if len(raw) == 0 || len(raw) > MaxEvidenceBytes || !json.Valid(raw) {
		return nil, false, ErrResponse
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value map[string]any
	if err := decoder.Decode(&value); err != nil || value == nil {
		return nil, false, ErrResponse
	}
	redacted := redactEvidenceValue(value)
	canonical, err := json.Marshal(value)
	if err != nil {
		return nil, false, ErrResponse
	}
	return canonical, redacted, nil
}

func redactEvidenceValue(value any) bool {
	changed := false
	switch v := value.(type) {
	case map[string]any:
		for key, child := range v {
			normalized := strings.ToLower(strings.NewReplacer("_", "", "-", "").Replace(key))
			switch normalized {
			case "apikey", "authorization", "proxyauthorization", "password", "secret", "token", "accesstoken", "refreshtoken", "cookie", "setcookie", "credentials", "auth":
				v[key] = "[REDACTED]"
				changed = true
			default:
				if redactEvidenceValue(child) {
					changed = true
				}
			}
		}
	case []any:
		for _, child := range v {
			if redactEvidenceValue(child) {
				changed = true
			}
		}
	}
	return changed
}

const snapshotColumns = `id,sequence,run_id,remote_task_id,remote_record_id,parent_record_id,representation,kind,version,content_sha256,redacted,captured_at`

func scanSnapshot(row interface{ Scan(...any) error }, withContent bool) (*EvidenceSnapshot, error) {
	item := &EvidenceSnapshot{Engine: "artex"}
	args := []any{&item.ID, &item.Sequence, &item.RunID, &item.RemoteTaskID, &item.RemoteRecordID, &item.ParentRecordID, &item.Representation, &item.Kind, &item.Version, &item.ContentSHA256, &item.Redacted, &item.CapturedAt}
	var content string
	if withContent {
		args = append(args, &content)
	}
	if err := row.Scan(args...); err != nil {
		return nil, err
	}
	if withContent {
		item.Content = json.RawMessage(content)
	}
	return item, nil
}

// SaveEvidence serializes version allocation with the run row. An identical
// latest observation reuses its original ID/time; A -> B -> A is three versions.
func (s *SQLRunStore) SaveEvidence(ctx context.Context, uid int64, runID, remoteTaskID string, record EvidenceRecord) (*EvidenceSnapshot, error) {
	if s == nil || s.db == nil {
		return nil, ErrStoreUnavailable
	}
	if uid <= 0 || strings.TrimSpace(runID) == "" || !validID(remoteTaskID) || !validID(record.ID) || !validEvidenceKind(record.Kind) || (record.ParentID != "" && !validID(record.ParentID)) || (record.Representation != "list" && record.Representation != "detail") {
		return nil, ErrInvalidRequest
	}
	content, redacted, err := CanonicalEvidence(record.Content)
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256(content)
	hash := hex.EncodeToString(digest[:])
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var owner string
	err = tx.QueryRowContext(ctx, `SELECT id FROM engine_runs WHERE engine_key='artex' AND id=$1 AND user_id=$2 AND engine_task_id=$3 FOR UPDATE`, runID, uid, remoteTaskID).Scan(&owner)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrRunNotFound
	}
	if err != nil {
		return nil, err
	}
	latest, err := scanSnapshot(tx.QueryRowContext(ctx, `SELECT `+snapshotColumns+`,content FROM engine_evidence_snapshots WHERE engine_key='artex' AND run_id=$1 AND remote_record_id=$2 AND kind=$3 AND representation=$4 ORDER BY version DESC LIMIT 1`, runID, record.ID, record.Kind, record.Representation), true)
	version := 1
	if err == nil {
		if latest.ParentRecordID != record.ParentID {
			return nil, ErrResponse
		}
		if latest.ContentSHA256 == hash {
			return latest, tx.Commit()
		}
		version = latest.Version + 1
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	item, err := scanSnapshot(tx.QueryRowContext(ctx, `INSERT INTO engine_evidence_snapshots (id,engine_key,run_id,user_id,remote_task_id,remote_record_id,parent_record_id,representation,kind,version,content,content_sha256,redacted) VALUES ($1,'artex',$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) RETURNING `+snapshotColumns+`,content`, uuid.NewString(), runID, uid, remoteTaskID, record.ID, record.ParentID, record.Representation, record.Kind, version, string(content), hash, redacted), true)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return item, nil
}

// History pages contain bounded metadata only. Content is fetched explicitly.
// A monotonic cursor avoids offset skips when a new observation arrives.
func (s *SQLRunStore) SnapshotPage(ctx context.Context, runID, kind, before string, limit int) (*SnapshotPage, error) {
	if s == nil || s.db == nil {
		return nil, ErrStoreUnavailable
	}
	if _, err := uuid.Parse(runID); err != nil {
		return nil, ErrInvalidRequest
	}
	if (kind != "" && !validEvidenceKind(kind)) || limit < 1 || limit > 100 {
		return nil, ErrInvalidRequest
	}
	var cursor int64
	if before != "" {
		var err error
		cursor, err = strconv.ParseInt(before, 10, 64)
		if err != nil || cursor <= 0 {
			return nil, ErrInvalidRequest
		}
	}
	query := `SELECT ` + snapshotColumns + ` FROM engine_evidence_snapshots WHERE engine_key='artex' AND run_id=$1 AND ($2='' OR kind=$2) AND ($3::bigint=0 OR sequence<$3) ORDER BY sequence DESC LIMIT $4`
	rows, err := s.db.QueryContext(ctx, query, runID, kind, cursor, limit+1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	page := &SnapshotPage{Items: []EvidenceSnapshot{}}
	for rows.Next() {
		item, err := scanSnapshot(rows, false)
		if err != nil {
			return nil, err
		}
		page.Items = append(page.Items, *item)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	if len(page.Items) > limit {
		page.HasMore = true
		page.Items = page.Items[:limit]
		page.NextCursor = strconv.FormatInt(page.Items[limit-1].Sequence, 10)
	}
	return page, nil
}

func (s *SQLRunStore) GetSnapshot(ctx context.Context, runID, id string) (*EvidenceSnapshot, error) {
	if s == nil || s.db == nil {
		return nil, ErrStoreUnavailable
	}
	item, err := scanSnapshot(s.db.QueryRowContext(ctx, `SELECT `+snapshotColumns+`,content FROM engine_evidence_snapshots WHERE engine_key='artex' AND run_id=$1 AND id=$2`, runID, id), true)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSnapshotNotFound
	}
	return item, err
}
