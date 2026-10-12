package artex

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
)

// ProfileSpec contains only transferable execution settings, never credentials.
// PentAGI-specific agent prompts/sampling defaults are not copied implicitly.
type ProfileSpec struct {
	Format          string `json:"format"`
	BaseURL         string `json:"base_url"`
	Proxy           string `json:"proxy,omitempty"`
	Model           string `json:"model"`
	MaxTokens       int    `json:"max_tokens"`
	ReasoningEffort string `json:"reasoning_effort,omitempty"`
	MaxTokensField  string `json:"max_tokens_field,omitempty"`
	Streaming       bool   `json:"streaming"`
	PoolExclude     bool   `json:"pool_exclude"`
}

type RemoteProfile struct {
	ProfileSpec
	ID        string `json:"id"`
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default"`
}

func (s ProfileSpec) Validate() error {
	if s.Format != "openai" || strings.TrimSpace(s.Model) == "" || len(s.Model) > 512 || strings.ContainsAny(s.Model, "\r\n\x00") || s.MaxTokens < 0 || s.MaxTokens > 1000000 || !s.PoolExclude {
		return ErrInvalidRequest
	}
	for _, raw := range []string{s.BaseURL, s.Proxy} {
		if raw == "" && raw == s.Proxy {
			continue
		}
		u, err := url.Parse(raw)
		if err != nil || u.Hostname() == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
			return ErrInvalidRequest
		}
	}
	if s.BaseURL == "" {
		return ErrInvalidRequest
	}
	switch s.ReasoningEffort {
	case "", "low", "medium", "high", "xhigh", "max":
	default:
		return ErrInvalidRequest
	}
	if s.MaxTokensField != "" && s.MaxTokensField != "max_completion_tokens" {
		return ErrInvalidRequest
	}
	return nil
}

func (s ProfileSpec) Revision() string {
	raw, _ := json.Marshal(s)
	digest := sha256.Sum256(raw)
	return hex.EncodeToString(digest[:])
}

// InstanceKey changes when the configured engine endpoint changes. Raw control
// tokens are excluded so a login renewal does not orphan the saved mappings.
func (c *Client) InstanceKey(namespace string) string {
	if !c.Enabled() {
		return ""
	}
	sum := sha256.Sum256([]byte(namespace + "\x00" + c.baseURL.String()))
	return hex.EncodeToString(sum[:])
}

func (c *Client) ListProfiles(ctx context.Context) ([]RemoteProfile, error) {
	raw, err := c.do(ctx, http.MethodGet, "/api/llm/profiles", nil, nil)
	if err != nil {
		return nil, err
	}
	var result struct {
		Profiles []RemoteProfile `json:"profiles"`
	}
	if json.Unmarshal(raw, &result) != nil {
		return nil, ErrResponse
	}
	return result.Profiles, nil
}

func (c *Client) SaveProfile(ctx context.Context, id int64, name string, spec ProfileSpec, key string) (int64, error) {
	if id < 0 || name == "" || len(name) > 200 || key == "" || spec.Validate() != nil {
		return 0, ErrInvalidRequest
	}
	payload := struct {
		ProfileSpec
		ID     int64  `json:"id,omitempty"`
		Name   string `json:"name"`
		APIKey string `json:"api_key"`
	}{spec, id, name, key}
	raw, err := c.do(ctx, http.MethodPost, "/api/llm/profiles", nil, payload)
	if err != nil {
		return 0, err
	}
	var result struct {
		ID int64 `json:"id"`
	}
	if json.Unmarshal(raw, &result) != nil || result.ID <= 0 {
		return 0, ErrResponse
	}
	return result.ID, nil
}
