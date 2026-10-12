package artex

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
)

var (
	ErrSourceUnavailable   = errors.New("model provider is unavailable or not owned by this user")
	ErrUnsupportedProvider = errors.New("model provider is not supported by this engine")
	ErrBindingMissing      = errors.New("model binding not found")
	ErrProfileDrift        = errors.New("managed ARTEX profile was changed or removed externally")
	ErrSyncUncertain       = errors.New("model synchronization outcome requires reconciliation")
	ErrCredentialChanged   = errors.New("provider credential changed; rotate the managed binding first")
	ErrEndpointChanged     = errors.New("provider endpoint changed; create a new configuration version")
	ErrBindingRevoked      = errors.New("model binding is revoked")
	ErrStoreUnavailable    = errors.New("model binding store is unavailable")
)

type ProviderChoice struct {
	Ref          string   `json:"ref"`
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	DefaultModel string   `json:"defaultModel"`
	Models       []string `json:"models"`
	Compatible   bool     `json:"compatible"`
	Reason       string   `json:"reason,omitempty"`
}

type ModelSelection struct {
	ProviderRef     string `json:"providerRef"`
	Model           string `json:"model"`
	MaxTokens       int    `json:"maxTokens"`
	ReasoningEffort string `json:"reasoningEffort,omitempty"`
	MaxTokensField  string `json:"maxTokensField,omitempty"`
}

type ResolvedProvider struct {
	BaseURL string
	Proxy   string
	Model   string
	APIKey  string `json:"-"`
}

type ModelSource interface {
	List(context.Context, int64) ([]ProviderChoice, error)
	Resolve(context.Context, int64, string, string) (ResolvedProvider, error)
}

type Binding struct {
	ID                    string      `json:"id"`
	EngineKey             string      `json:"-"`
	UserID                int64       `json:"-"`
	SourceRef             string      `json:"providerRef"`
	Revision              string      `json:"revision"`
	RemoteName            string      `json:"-"`
	RemoteID              int64       `json:"remoteProfileId"`
	Spec                  ProfileSpec `json:"spec"`
	CredentialFingerprint string      `json:"-"`
	PendingFingerprint    string      `json:"-"`
	State                 string      `json:"state"`
	CreatedAt             time.Time   `json:"createdAt"`
	UpdatedAt             time.Time   `json:"updatedAt"`
}

type BindingRepository interface {
	FindRevision(context.Context, string) (*Binding, error)
	Get(context.Context, string) (*Binding, error)
	Insert(context.Context, *Binding) error
	Save(context.Context, *Binding) error
}
type BindingStore interface {
	WithSourceLock(context.Context, string, int64, string, func(BindingRepository) error) error
	Get(context.Context, string, int64, string) (*Binding, error)
	List(context.Context, string, int64) ([]Binding, error)
}

type ModelSharing struct {
	client         *Client
	store          BindingStore
	source         ModelSource
	engineKey      string
	fingerprintKey []byte
}

func NewModelSharing(client *Client, store BindingStore, source ModelSource, namespace, secret string) *ModelSharing {
	if !client.Enabled() || store == nil || source == nil || namespace == "" || secret == "" {
		return nil
	}
	return &ModelSharing{client: client, store: store, source: source, engineKey: client.InstanceKey(namespace), fingerprintKey: []byte(secret)}
}
func (m *ModelSharing) ListProviders(ctx context.Context, uid int64) ([]ProviderChoice, error) {
	if m == nil {
		return nil, ErrStoreUnavailable
	}
	return m.source.List(ctx, uid)
}
func (m *ModelSharing) ListBindings(ctx context.Context, uid int64) ([]Binding, error) {
	if m == nil {
		return nil, ErrStoreUnavailable
	}
	return m.store.List(ctx, m.engineKey, uid)
}
func (m *ModelSharing) fingerprint(uid int64, ref, key string) string {
	hash := hmac.New(sha256.New, m.fingerprintKey)
	_, _ = hash.Write([]byte(strconv.FormatInt(uid, 10) + "\x00" + ref + "\x00" + key))
	return hex.EncodeToString(hash.Sum(nil))
}

func (m *ModelSharing) Sync(ctx context.Context, uid int64, selection ModelSelection) (*Binding, error) {
	if m == nil {
		return nil, ErrStoreUnavailable
	}
	if !m.client.AllowsWrites() {
		return nil, ErrWritesDisabled
	}
	if uid <= 0 || selection.ProviderRef == "" || len(selection.ProviderRef) > 128 {
		return nil, ErrInvalidRequest
	}
	var binding *Binding
	err := m.store.WithSourceLock(ctx, m.engineKey, uid, selection.ProviderRef, func(repo BindingRepository) error {
		source, err := m.source.Resolve(ctx, uid, selection.ProviderRef, selection.Model)
		if err != nil {
			return err
		}
		spec := ProfileSpec{Format: "openai", BaseURL: source.BaseURL, Proxy: source.Proxy, Model: source.Model, MaxTokens: selection.MaxTokens, ReasoningEffort: selection.ReasoningEffort, MaxTokensField: selection.MaxTokensField, Streaming: true, PoolExclude: true}
		if spec.Validate() != nil || source.APIKey == "" {
			return ErrInvalidRequest
		}
		fingerprint := m.fingerprint(uid, selection.ProviderRef, source.APIKey)
		binding, err = repo.FindRevision(ctx, spec.Revision())
		if errors.Is(err, ErrBindingMissing) {
			id := uuid.NewString()
			binding = &Binding{ID: id, EngineKey: m.engineKey, UserID: uid, SourceRef: selection.ProviderRef, Revision: spec.Revision(), RemoteName: "pentagi-managed-" + id, Spec: spec, CredentialFingerprint: fingerprint, PendingFingerprint: fingerprint, State: "pending"}
			if err = repo.Insert(ctx, binding); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		if binding.State == "revoked" {
			return ErrBindingRevoked
		}
		if binding.State == "rotating" || (binding.State == "uncertain" && binding.RemoteID > 0) {
			return ErrSyncUncertain
		}
		if binding.CredentialFingerprint != fingerprint {
			return ErrCredentialChanged
		}
		remote, err := m.findRemote(ctx, binding)
		if err != nil {
			return err
		}
		if binding.State == "ready" {
			if remote == nil || remote.ProfileSpec != binding.Spec {
				return ErrProfileDrift
			}
			return nil
		}
		if remote != nil {
			if remote.ProfileSpec != binding.Spec {
				return ErrProfileDrift
			}
			binding.RemoteID, _ = strconv.ParseInt(remote.ID, 10, 64)
		}
		// The journal row is committed before an external write. On cancellation or
		// response loss it stays pending/uncertain and retains its unguessable name.
		binding.State = "pending"
		if err = repo.Save(ctx, binding); err != nil {
			return err
		}
		remoteID, err := m.client.SaveProfile(ctx, binding.RemoteID, binding.RemoteName, binding.Spec, source.APIKey)
		if err != nil {
			binding.State = "uncertain"
			cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 3*time.Second)
			defer cancel()
			_ = repo.Save(cleanup, binding)
			return ErrSyncUncertain
		}
		binding.RemoteID = remoteID
		binding.State = "ready"
		binding.PendingFingerprint = ""
		return repo.Save(ctx, binding)
	})
	if err != nil {
		return nil, err
	}
	return binding, nil
}

func (m *ModelSharing) RotateKey(ctx context.Context, uid int64, id string) (*Binding, error) {
	if m == nil {
		return nil, ErrStoreUnavailable
	}
	if !m.client.AllowsWrites() {
		return nil, ErrWritesDisabled
	}
	initial, err := m.store.Get(ctx, m.engineKey, uid, id)
	if err != nil {
		return nil, err
	}
	var binding *Binding
	err = m.store.WithSourceLock(ctx, m.engineKey, uid, initial.SourceRef, func(repo BindingRepository) error {
		var err error
		binding, err = repo.Get(ctx, id)
		if err != nil {
			return err
		}
		if binding.State == "revoked" {
			return ErrBindingRevoked
		}
		if binding.RemoteID <= 0 {
			return ErrSyncUncertain
		}
		source, err := m.source.Resolve(ctx, uid, binding.SourceRef, binding.Spec.Model)
		if err != nil {
			return err
		}
		if source.BaseURL != binding.Spec.BaseURL || source.Proxy != binding.Spec.Proxy {
			return ErrEndpointChanged
		}
		fp := m.fingerprint(uid, binding.SourceRef, source.APIKey)
		if binding.PendingFingerprint != "" && binding.PendingFingerprint != fp {
			return ErrSyncUncertain
		}
		remote, err := m.findRemote(ctx, binding)
		if err != nil {
			return err
		}
		if remote == nil || remote.ProfileSpec != binding.Spec {
			return ErrProfileDrift
		}
		binding.State = "rotating"
		binding.PendingFingerprint = fp
		if err = repo.Save(ctx, binding); err != nil {
			return err
		}
		_, err = m.client.SaveProfile(ctx, binding.RemoteID, binding.RemoteName, binding.Spec, source.APIKey)
		if err != nil {
			return ErrSyncUncertain
		}
		binding.CredentialFingerprint = fp
		binding.PendingFingerprint = ""
		binding.State = "ready"
		return repo.Save(ctx, binding)
	})
	if err != nil {
		return nil, err
	}
	return binding, nil
}

func (m *ModelSharing) ForTask(ctx context.Context, uid int64, id string) (*Binding, error) {
	if m == nil {
		return nil, ErrStoreUnavailable
	}
	binding, err := m.store.Get(ctx, m.engineKey, uid, id)
	if err != nil {
		return nil, err
	}
	if binding.State != "ready" {
		return nil, ErrSyncUncertain
	}
	source, err := m.source.Resolve(ctx, uid, binding.SourceRef, binding.Spec.Model)
	if err != nil {
		return nil, err
	}
	if source.BaseURL != binding.Spec.BaseURL || source.Proxy != binding.Spec.Proxy {
		return nil, ErrEndpointChanged
	}
	if m.fingerprint(uid, binding.SourceRef, source.APIKey) != binding.CredentialFingerprint {
		return nil, ErrCredentialChanged
	}
	remote, err := m.findRemote(ctx, binding)
	if err != nil {
		return nil, err
	}
	if remote == nil || remote.ProfileSpec != binding.Spec {
		return nil, ErrProfileDrift
	}
	return binding, nil
}

func (m *ModelSharing) findRemote(ctx context.Context, binding *Binding) (*RemoteProfile, error) {
	profiles, err := m.client.ListProfiles(ctx)
	if err != nil {
		return nil, err
	}
	for _, profile := range profiles {
		if profile.ID == strconv.FormatInt(binding.RemoteID, 10) || profile.Name == binding.RemoteName {
			if profile.Name != binding.RemoteName || !validID(profile.ID) || profile.IsDefault {
				return nil, ErrProfileDrift
			}
			if binding.RemoteID > 0 && profile.ID != strconv.FormatInt(binding.RemoteID, 10) {
				return nil, ErrProfileDrift
			}
			return &profile, nil
		}
	}
	return nil, nil
}
