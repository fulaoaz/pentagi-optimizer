package artex

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

type memoryBindings struct {
	mu    sync.Mutex
	items map[string]Binding
}
type memoryBindingScope struct {
	s      *memoryBindings
	engine string
	uid    int64
	source string
}

func (s *memoryBindings) WithSourceLock(ctx context.Context, engine string, uid int64, source string, fn func(BindingRepository) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return fn(&memoryBindingScope{s, engine, uid, source})
}
func (s *memoryBindings) Get(ctx context.Context, engine string, uid int64, id string) (*Binding, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[id]
	if !ok || item.EngineKey != engine || item.UserID != uid {
		return nil, ErrBindingMissing
	}
	return &item, nil
}
func (s *memoryBindings) List(ctx context.Context, engine string, uid int64) ([]Binding, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []Binding{}
	for _, item := range s.items {
		if item.EngineKey == engine && item.UserID == uid {
			out = append(out, item)
		}
	}
	return out, nil
}
func (r *memoryBindingScope) FindRevision(ctx context.Context, rev string) (*Binding, error) {
	for _, item := range r.s.items {
		if item.EngineKey == r.engine && item.UserID == r.uid && item.SourceRef == r.source && item.Revision == rev {
			return &item, nil
		}
	}
	return nil, ErrBindingMissing
}
func (r *memoryBindingScope) Get(ctx context.Context, id string) (*Binding, error) {
	item, ok := r.s.items[id]
	if !ok || item.EngineKey != r.engine || item.UserID != r.uid || item.SourceRef != r.source {
		return nil, ErrBindingMissing
	}
	return &item, nil
}
func (r *memoryBindingScope) Insert(ctx context.Context, item *Binding) error {
	item.CreatedAt = time.Now()
	return r.Save(ctx, item)
}
func (r *memoryBindingScope) Save(ctx context.Context, item *Binding) error {
	item.UpdatedAt = time.Now()
	r.s.items[item.ID] = *item
	return nil
}

type sharingSource struct{ key, base string }

func (s *sharingSource) List(context.Context, int64) ([]ProviderChoice, error) { return nil, nil }
func (s *sharingSource) Resolve(ctx context.Context, uid int64, ref, model string) (ResolvedProvider, error) {
	if uid != 1 || ref != "saved:1" {
		return ResolvedProvider{}, ErrSourceUnavailable
	}
	return ResolvedProvider{BaseURL: s.base, Model: model, APIKey: s.key}, nil
}

type profileHarness struct {
	mu              sync.Mutex
	profiles        map[int64]RemoteProfile
	keys            map[int64]string
	next            int64
	creates, writes int
	loseResponse    bool
}

func (h *profileHarness) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if r.URL.Path != "/api/llm/profiles" {
		w.WriteHeader(404)
		return
	}
	if r.Method == http.MethodGet {
		items := []RemoteProfile{}
		for _, item := range h.profiles {
			items = append(items, item)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"profiles": items})
		return
	}
	var body struct {
		ProfileSpec
		ID     int64  `json:"id"`
		Name   string `json:"name"`
		APIKey string `json:"api_key"`
	}
	if json.NewDecoder(r.Body).Decode(&body) != nil {
		w.WriteHeader(400)
		return
	}
	if body.ID == 0 {
		h.next++
		body.ID = h.next
		h.creates++
	}
	h.writes++
	h.profiles[body.ID] = RemoteProfile{ProfileSpec: body.ProfileSpec, ID: strconv.FormatInt(body.ID, 10), Name: body.Name}
	h.keys[body.ID] = body.APIKey
	if h.loseResponse {
		h.loseResponse = false
		w.WriteHeader(502)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"id": body.ID})
}
func newSharingHarness(t *testing.T) (*ModelSharing, *profileHarness, *sharingSource, *memoryBindings) {
	t.Helper()
	upstream := &profileHarness{profiles: map[int64]RemoteProfile{}, keys: map[int64]string{}}
	server := httptest.NewServer(upstream)
	t.Cleanup(server.Close)
	client, err := NewClient(Config{Enabled: true, AllowWrites: true, BaseURL: server.URL, APIKey: "control-fixture"})
	if err != nil {
		t.Fatal(err)
	}
	source := &sharingSource{key: "fixture-first-key", base: "https://model.invalid/v1"}
	store := &memoryBindings{items: map[string]Binding{}}
	return NewModelSharing(client, store, source, "test-installation", "fingerprint-fixture"), upstream, source, store
}
func TestModelSharingPersistsVersionsAndOwnership(t *testing.T) {
	manager, upstream, _, store := newSharingHarness(t)
	ctx := context.Background()
	selection := ModelSelection{ProviderRef: "saved:1", Model: "test-model", MaxTokens: 100}
	first, err := manager.Sync(ctx, 1, selection)
	if err != nil {
		t.Fatal(err)
	}
	// A new manager over the same durable store must reuse the remote identity.
	restarted := NewModelSharing(manager.client, store, manager.source, "test-installation", "fingerprint-fixture")
	again, err := restarted.Sync(ctx, 1, selection)
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != again.ID || first.RemoteID != again.RemoteID || upstream.creates != 1 {
		t.Fatal("duplicate binding")
	}
	selection.MaxTokens = 200
	second, err := manager.Sync(ctx, 1, selection)
	if err != nil {
		t.Fatal(err)
	}
	if second.RemoteID == first.RemoteID || first.Revision == second.Revision || upstream.profiles[first.RemoteID].MaxTokens != 100 {
		t.Fatal("old configuration mutated")
	}
	if _, err := manager.ForTask(ctx, 2, first.ID); !errors.Is(err, ErrBindingMissing) {
		t.Fatal("cross-user binding accepted")
	}
	if _, err := manager.Sync(ctx, 2, selection); !errors.Is(err, ErrSourceUnavailable) {
		t.Fatal("cross-user source accepted")
	}
	encoded, _ := json.Marshal(first)
	if strings.Contains(string(encoded), "fixture-first-key") || strings.Contains(string(encoded), "Fingerprint") {
		t.Fatal("credential in binding output")
	}
}
func TestModelSharingReconcilesLostCreationResponse(t *testing.T) {
	manager, upstream, _, store := newSharingHarness(t)
	upstream.loseResponse = true
	selection := ModelSelection{ProviderRef: "saved:1", Model: "test-model"}
	if _, err := manager.Sync(context.Background(), 1, selection); !errors.Is(err, ErrSyncUncertain) {
		t.Fatal(err)
	}
	if len(store.items) != 1 {
		t.Fatal("recovery journal was lost")
	}
	binding, err := manager.Sync(context.Background(), 1, selection)
	if err != nil {
		t.Fatal(err)
	}
	if upstream.creates != 1 || binding.State != "ready" {
		t.Fatal("uncertain create duplicated")
	}
}
func TestModelSharingConcurrentSyncCreatesOnce(t *testing.T) {
	manager, upstream, _, _ := newSharingHarness(t)
	var wg sync.WaitGroup
	failures := make(chan error, 8)
	for n := 0; n < 8; n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := manager.Sync(context.Background(), 1, ModelSelection{ProviderRef: "saved:1", Model: "test-model"})
			failures <- err
		}()
	}
	wg.Wait()
	close(failures)
	for err := range failures {
		if err != nil {
			t.Fatal(err)
		}
	}
	if upstream.creates != 1 {
		t.Fatalf("created %d profiles", upstream.creates)
	}
}
func TestModelSharingRequiresExplicitRotationAndProtectsEndpoint(t *testing.T) {
	manager, upstream, source, _ := newSharingHarness(t)
	ctx := context.Background()
	selection := ModelSelection{ProviderRef: "saved:1", Model: "test-model"}
	binding, err := manager.Sync(ctx, 1, selection)
	if err != nil {
		t.Fatal(err)
	}
	source.key = "fixture-rotated-key"
	if _, err = manager.ForTask(ctx, 1, binding.ID); !errors.Is(err, ErrCredentialChanged) {
		t.Fatal(err)
	}
	if _, err = manager.Sync(ctx, 1, selection); !errors.Is(err, ErrCredentialChanged) {
		t.Fatal(err)
	}
	rotated, err := manager.RotateKey(ctx, 1, binding.ID)
	if err != nil {
		t.Fatal(err)
	}
	if rotated.RemoteID != binding.RemoteID || upstream.keys[binding.RemoteID] != source.key {
		t.Fatal("rotation replaced binding or missed key")
	}
	if _, err = manager.ForTask(ctx, 1, binding.ID); err != nil {
		t.Fatal(err)
	}
	source.base = "https://different.invalid/v1"
	before := upstream.writes
	if _, err = manager.RotateKey(ctx, 1, binding.ID); !errors.Is(err, ErrEndpointChanged) {
		t.Fatal(err)
	}
	if upstream.writes != before {
		t.Fatal("key sent after endpoint mismatch")
	}
}
func TestModelSharingRejectsExternalProfileReplacement(t *testing.T) {
	manager, upstream, _, _ := newSharingHarness(t)
	binding, err := manager.Sync(context.Background(), 1, ModelSelection{ProviderRef: "saved:1", Model: "test-model"})
	if err != nil {
		t.Fatal(err)
	}
	profile := upstream.profiles[binding.RemoteID]
	profile.Name = "human-owned-profile"
	upstream.profiles[binding.RemoteID] = profile
	writes := upstream.writes
	if _, err = manager.ForTask(context.Background(), 1, binding.ID); !errors.Is(err, ErrProfileDrift) {
		t.Fatal(err)
	}
	if _, err = manager.RotateKey(context.Background(), 1, binding.ID); !errors.Is(err, ErrProfileDrift) {
		t.Fatal(err)
	}
	if writes != upstream.writes {
		t.Fatal("external profile overwritten")
	}
}
