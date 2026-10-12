package artex

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const MaxResponseBytes = 16 << 20

var (
	ErrDisabled       = errors.New("ARTEX integration is disabled")
	ErrWritesDisabled = errors.New("ARTEX write operations are disabled")
	ErrInvalidRequest = errors.New("invalid ARTEX request")
	ErrResponse       = errors.New("invalid or oversized ARTEX response")
	ErrUnavailable    = errors.New("ARTEX connection failed")
	ErrTimeout        = errors.New("ARTEX request timed out")
)

type Config struct {
	Enabled     bool
	BaseURL     string
	APIKey      string
	AllowWrites bool
	Timeout     time.Duration
	HTTPClient  *http.Client
}

type Client struct {
	baseURL     *url.URL
	apiKey      string
	allowWrites bool
	httpClient  *http.Client
}

// Upstream bodies and URLs can contain evidence or credentials. Errors carry
// only status codes, including when logged by the API layer.
type HTTPError struct{ StatusCode int }

func (e *HTTPError) Error() string { return fmt.Sprintf("ARTEX returned HTTP %d", e.StatusCode) }

type CreateTaskRequest struct {
	Name           string  `json:"name,omitempty"`
	Description    string  `json:"description"`
	Goal           string  `json:"goal"`
	TimeoutSeconds int     `json:"timeout_seconds,omitempty"`
	LLMProfileIDs  []int64 `json:"llm_profile_ids,omitempty"`
}
type ControlTaskRequest struct {
	Action string `json:"action"`
}
type RetestRequest struct {
	Notes string `json:"notes"`
}

// TaskID extracts only the remote task identifier from a create response.
func TaskID(raw json.RawMessage) string {
	var v struct {
		ID string `json:"id"`
	}
	if json.Unmarshal(raw, &v) != nil {
		return ""
	}
	return strings.TrimSpace(v.ID)
}

func NewClient(cfg Config) (*Client, error) {
	if !cfg.Enabled {
		return &Client{}, nil
	}
	base, err := url.Parse(strings.TrimSpace(cfg.BaseURL))
	if err != nil || base == nil || (base.Scheme != "http" && base.Scheme != "https") || base.Hostname() == "" || base.User != nil || base.RawQuery != "" || base.ForceQuery || base.Fragment != "" {
		return nil, errors.New("ARTEX_URL must be an http(s) service URL without credentials, query or fragment")
	}
	if strings.Contains(base.Path, "..") || strings.ContainsAny(base.Path, "%\\") {
		return nil, errors.New("invalid ARTEX_URL path")
	}
	base.Path = strings.TrimRight(base.Path, "/")
	if strings.HasSuffix(base.Path, "/api") {
		return nil, errors.New("ARTEX_URL must be the service root, without /api")
	}
	key := strings.TrimSpace(cfg.APIKey)
	if key == "" || strings.ContainsAny(key, "\r\n") {
		return nil, errors.New("ARTEX_API_KEY must contain the ARTEX login token")
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	if timeout > 120*time.Second {
		return nil, errors.New("ARTEX_TIMEOUT must be at most 120 seconds")
	}
	client := http.Client{}
	if cfg.HTTPClient != nil {
		client = *cfg.HTTPClient
	}
	client.Timeout = timeout
	// Redirects must not escape the allow-list or forward the Bearer token.
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	return &Client{baseURL: base, apiKey: key, allowWrites: cfg.AllowWrites, httpClient: &client}, nil
}
func (c *Client) Enabled() bool      { return c != nil && c.baseURL != nil }
func (c *Client) AllowsWrites() bool { return c.Enabled() && c.allowWrites }
func validID(id string) bool {
	value, err := strconv.ParseInt(id, 10, 64)
	return err == nil && value > 0 && strconv.FormatInt(value, 10) == id
}
func (c *Client) Health(ctx context.Context) (json.RawMessage, error) {
	raw, err := c.do(ctx, http.MethodGet, "/api/health", nil, nil)
	if err != nil {
		return nil, err
	}
	var health struct {
		OK      bool   `json:"ok"`
		Service string `json:"service"`
	}
	if json.Unmarshal(raw, &health) != nil || !health.OK || health.Service != "artex" {
		return nil, ErrResponse
	}
	return raw, nil
}
func (c *Client) ListTasks(ctx context.Context) (json.RawMessage, error) {
	return c.do(ctx, http.MethodGet, "/api/tasks", nil, nil)
}
func (c *Client) GetTask(ctx context.Context, id string) (json.RawMessage, error) {
	if !validID(id) {
		return nil, ErrInvalidRequest
	}
	return c.do(ctx, http.MethodGet, "/api/tasks/"+id, nil, nil)
}
func (c *Client) CreateTask(ctx context.Context, input CreateTaskRequest) (json.RawMessage, error) {
	if len(input.LLMProfileIDs) > 1 || (len(input.LLMProfileIDs) == 1 && input.LLMProfileIDs[0] <= 0) {
		return nil, ErrInvalidRequest
	}
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	input.Goal = strings.TrimSpace(input.Goal)
	if input.Description == "" || input.Goal == "" || utf8.RuneCountInString(input.Name) > 200 || utf8.RuneCountInString(input.Description) > 8000 || utf8.RuneCountInString(input.Goal) > 8000 || input.TimeoutSeconds < 0 || input.TimeoutSeconds > 86400 {
		return nil, ErrInvalidRequest
	}
	return c.do(ctx, http.MethodPost, "/api/tasks", nil, input)
}
func (c *Client) ControlTask(ctx context.Context, id string, input ControlTaskRequest) (json.RawMessage, error) {
	if !validID(id) || (input.Action != "pause" && input.Action != "resume") {
		return nil, ErrInvalidRequest
	}
	return c.do(ctx, http.MethodPost, "/api/tasks/"+id+"/control", nil, input)
}
func (c *Client) ListAssets(ctx context.Context, taskID string, page int) (json.RawMessage, error) {
	if !validID(taskID) || page < 1 || page > 100000 {
		return nil, ErrInvalidRequest
	}
	query := url.Values{"task_id": {taskID}, "limit": {"50"}, "offset": {strconv.Itoa((page - 1) * 50)}}
	return c.do(ctx, http.MethodGet, "/api/assets", query, nil)
}
func (c *Client) ListFindings(ctx context.Context, taskID string, page int) (json.RawMessage, error) {
	if !validID(taskID) || page < 1 || page > 100000 {
		return nil, ErrInvalidRequest
	}
	// Paginated findings return stable finding IDs. ?task returns graph-node IDs.
	query := url.Values{"task_id": {taskID}, "page": {strconv.Itoa(page)}, "limit": {"50"}}
	return c.do(ctx, http.MethodGet, "/api/exploration/findings", query, nil)
}
func (c *Client) GetFinding(ctx context.Context, id string) (json.RawMessage, error) {
	if !validID(id) {
		return nil, ErrInvalidRequest
	}
	return c.do(ctx, http.MethodGet, "/api/exploration/findings/"+id, nil, nil)
}
func (c *Client) Activity(ctx context.Context, taskID, since string) (json.RawMessage, error) {
	if !validID(taskID) || (since != "0" && !validID(since)) {
		return nil, ErrInvalidRequest
	}
	return c.do(ctx, http.MethodGet, "/api/exploration/activity", url.Values{"task": {taskID}, "limit": {"100"}, "since": {since}}, nil)
}
func (c *Client) ListRetests(ctx context.Context, id string) (json.RawMessage, error) {
	if !validID(id) {
		return nil, ErrInvalidRequest
	}
	return c.do(ctx, http.MethodGet, "/api/exploration/findings/"+id+"/retests", nil, nil)
}
func (c *Client) StartRetest(ctx context.Context, id string, input RetestRequest) (json.RawMessage, error) {
	if !validID(id) || utf8.RuneCountInString(input.Notes) > 4000 {
		return nil, ErrInvalidRequest
	}
	return c.do(ctx, http.MethodPost, "/api/exploration/findings/"+id+"/retests", nil, input)
}
func (c *Client) do(ctx context.Context, method, endpoint string, query url.Values, input any) (json.RawMessage, error) {
	if !c.Enabled() {
		return nil, ErrDisabled
	}
	if method != http.MethodGet && !c.allowWrites {
		return nil, ErrWritesDisabled
	}
	requestURL := *c.baseURL
	requestURL.Path = c.baseURL.Path + endpoint
	requestURL.RawPath = ""
	requestURL.RawQuery = query.Encode()
	var body io.Reader
	if input != nil {
		payload, err := json.Marshal(input)
		if err != nil {
			return nil, ErrInvalidRequest
		}
		body = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(ctx, method, requestURL.String(), body)
	if err != nil {
		return nil, ErrInvalidRequest
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	if input != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		var netErr interface{ Timeout() bool }
		if errors.As(err, &netErr) && netErr.Timeout() {
			return nil, ErrTimeout
		}
		return nil, ErrUnavailable
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &HTTPError{StatusCode: resp.StatusCode}
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, MaxResponseBytes+1))
	if err != nil {
		return nil, ErrUnavailable
	}
	if len(data) > MaxResponseBytes || !json.Valid(data) {
		return nil, ErrResponse
	}
	return json.RawMessage(data), nil
}
