//go:build artexintegration

package artex

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"pentagi/migrations"
)

func TestSQLModelSharingLive(t *testing.T) {
	dsn := os.Getenv("ARTEX_LAB_PG_DSN")
	password := os.Getenv("ARTEX_LAB_ADMIN_PASSWORD")
	oldKey, newKey := os.Getenv("ARTEX_LAB_KEY_A"), os.Getenv("ARTEX_LAB_KEY_ROTATED")
	if dsn == "" || password == "" || oldKey == "" || newKey == "" {
		t.Fatal("isolated ARTEX lab credentials are required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	api := func(method, path string, body any, token string) json.RawMessage {
		t.Helper()
		raw, _ := json.Marshal(body)
		request, _ := http.NewRequestWithContext(ctx, method, "http://artex:8787"+path, bytes.NewReader(raw))
		request.Header.Set("Content-Type", "application/json")
		if token != "" {
			request.Header.Set("Authorization", "Bearer "+token)
		}
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Fatal("real ARTEX connection failed")
		}
		defer response.Body.Close()
		data, _ := io.ReadAll(io.LimitReader(response.Body, MaxResponseBytes))
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			t.Fatalf("ARTEX %s %s status %d", method, path, response.StatusCode)
		}
		return data
	}
	var auth struct {
		Token string `json:"token"`
	}
	_ = json.Unmarshal(api("POST", "/api/auth/login", map[string]string{"username": "ARTEX", "password": password}, ""), &auth)
	if auth.Token == "" {
		t.Fatal("ARTEX lab login failed")
	}
	root, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal("lab database configuration failed")
	}
	defer root.Close()
	schema := "pentagi_share_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err = root.ExecContext(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal("create isolated metadata schema failed")
	}
	defer func() { _, _ = root.ExecContext(context.Background(), "DROP SCHEMA "+schema+" CASCADE") }()
	endpoint, err := url.Parse(dsn)
	if err != nil {
		t.Fatal("lab DSN parse failed")
	}
	query := endpoint.Query()
	query.Set("search_path", schema)
	endpoint.RawQuery = query.Encode()
	db, err := sql.Open("postgres", endpoint.String())
	if err != nil {
		t.Fatal("open metadata database failed")
	}
	defer db.Close()
	db.SetMaxOpenConns(8)
	for _, migrationName := range []string{"sql/20260924_200000_engine_provider_bindings.sql", "sql/20260925_120000_engine_runs.sql"} {
		migration, readErr := migrations.EmbedMigrations.ReadFile(migrationName)
		if readErr != nil {
			t.Fatal(readErr)
		}
		if _, execErr := db.ExecContext(ctx, strings.Split(string(migration), "-- +goose Down")[0]); execErr != nil {
			t.Fatalf("apply migration %s: %v", migrationName, execErr)
		}
	}
	client, err := NewClient(Config{Enabled: true, AllowWrites: true, BaseURL: "http://artex:8787", APIKey: auth.Token})
	if err != nil {
		t.Fatal(err)
	}
	source := &sharingSource{key: oldKey, base: "http://artex:18999/v1"}
	store := NewSQLBindingStore(db)
	manager := NewModelSharing(client, store, source, schema, "synthetic-fingerprint-salt")
	selection := ModelSelection{ProviderRef: "saved:1", Model: "pentagi-shared-fixture", MaxTokens: 128}
	var wg sync.WaitGroup
	results := make(chan *Binding, 4)
	failures := make(chan error, 4)
	for n := 0; n < 4; n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			binding, err := manager.Sync(ctx, 1, selection)
			results <- binding
			failures <- err
		}()
	}
	wg.Wait()
	close(results)
	close(failures)
	for err := range failures {
		if err != nil {
			t.Fatal(err)
		}
	}
	var binding *Binding
	for candidate := range results {
		if binding != nil && binding.ID != candidate.ID {
			t.Fatal("concurrent sync created duplicate local mapping")
		}
		binding = candidate
	}
	t.Log("PASS persistent_concurrent_sync_reuses_binding")
	defer func() {
		if binding != nil {
			api("DELETE", fmt.Sprintf("/api/llm/profiles/%d", binding.RemoteID), nil, auth.Token)
		}
	}()
	// Simulate process reconstruction from PostgreSQL, with no in-memory cache.
	reconstructed := NewModelSharing(client, NewSQLBindingStore(db), source, schema, "synthetic-fingerprint-salt")
	restored, err := reconstructed.Sync(ctx, 1, selection)
	if err != nil || restored.ID != binding.ID || restored.RemoteID != binding.RemoteID {
		t.Fatal("persistent mapping was not restored")
	}
	t.Log("PASS new_manager_restores_remote_identity")
	selection.MaxTokens = 256
	newer, err := manager.Sync(ctx, 1, selection)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { api("DELETE", fmt.Sprintf("/api/llm/profiles/%d", newer.RemoteID), nil, auth.Token) }()
	if newer.RemoteID == binding.RemoteID || newer.Revision == binding.Revision {
		t.Fatal("config revision overwrote original")
	}
	remote, err := manager.findRemote(ctx, binding)
	if err != nil || remote == nil || remote.MaxTokens != 128 {
		t.Fatal("original remote version changed")
	}
	t.Log("PASS changed_parameters_create_distinct_remote_version")
	source.key = newKey
	if _, err = manager.ForTask(ctx, 1, binding.ID); err != ErrCredentialChanged {
		t.Fatal("stale key was accepted")
	}
	rotated, err := manager.RotateKey(ctx, 1, binding.ID)
	if err != nil || rotated.RemoteID != binding.RemoteID {
		t.Fatal("real key rotation failed")
	}
	t.Log("PASS real_profile_key_rotation_preserves_id")
	if _, err = manager.ForTask(ctx, 2, binding.ID); err != ErrBindingMissing {
		t.Fatal("other user could select binding")
	}
	t.Log("PASS foreign_owner_cannot_select_binding")
	selected, err := manager.ForTask(ctx, 1, binding.ID)
	if err != nil {
		t.Fatal(err)
	}
	runStore := NewSQLRunStore(db)
	run, err := runStore.CreatePending(ctx, 1, RunRequest{Name: "live run mapping", Description: "fixture", Goal: "fixture", BindingID: binding.ID, BindingRevision: binding.Revision})
	if err != nil {
		t.Fatal("local run record failed: ", err)
	}
	if err = runStore.MarkCreated(ctx, run.ID, "live-task-id"); err != nil {
		t.Fatal(err)
	}
	runs, err := runStore.List(ctx, 1)
	if err != nil || len(runs) != 1 || runs[0].EngineTaskID != "live-task-id" || runs[0].Status != "queued" {
		t.Fatal("local run mapping was not durable")
	}
	if err = runStore.MarkUncertain(ctx, run.ID); err != nil {
		t.Fatal(err)
	}
	runs, err = runStore.List(ctx, 1)
	if err != nil || len(runs) != 1 || runs[0].SubmitState != "uncertain" || runs[0].Status != "sync_unknown" {
		t.Fatal("uncertain submission state was not preserved")
	}
	t.Log("PASS local_run_mapping_and_uncertain_submission")
	raw, err := client.CreateTask(ctx, CreateTaskRequest{Name: "PentAGI lab shared manager", Description: "Local integration fixture only.", Goal: "Acknowledge protocol without external tools.", TimeoutSeconds: 3600, LLMProfileIDs: []int64{selected.RemoteID}})
	if err != nil {
		t.Fatal("managed task creation failed")
	}
	var task struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(raw, &task)
	if !validID(task.ID) {
		t.Fatal("invalid task response")
	}
	// Delete only this test-owned task before removing its profiles.
	defer func() { api("DELETE", "/api/tasks/"+task.ID, nil, auth.Token) }()
	wantedDigest := sha256.Sum256([]byte("Bearer " + newKey))
	matched := false
	for end := time.Now().Add(20 * time.Second); time.Now().Before(end); {
		response, err := http.Get("http://artex:18999/diagnostics")
		if err != nil {
			t.Fatal("fixture diagnostics unavailable")
		}
		var diag struct {
			Calls []struct {
				Model  string `json:"model"`
				Digest string `json:"credentialDigest"`
			} `json:"calls"`
		}
		_ = json.NewDecoder(response.Body).Decode(&diag)
		response.Body.Close()
		for _, call := range diag.Calls {
			if call.Model == selection.Model && call.Digest == hex.EncodeToString(wantedDigest[:]) {
				matched = true
			}
		}
		if matched {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !matched {
		t.Fatal("real engine did not call the synchronized model/key")
	}
	if _, err = client.ControlTask(ctx, task.ID, ControlTaskRequest{Action: "pause"}); err != nil {
		t.Fatal("managed task pause failed")
	}
	t.Log("PASS real_artex_uses_synchronized_model_and_rotated_key")
	encoded, _ := json.Marshal(rotated)
	if strings.Contains(string(encoded), oldKey) || strings.Contains(string(encoded), newKey) {
		t.Fatal("binding response leaked credentials")
	}
	t.Log("PASS binding_response_contains_no_key")
}
