package artex

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func testClient(t *testing.T, handler http.HandlerFunc, write bool) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	client, err := NewClient(Config{Enabled: true, BaseURL: server.URL + "/engine", APIKey: "test-token", AllowWrites: write, Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	return client
}
func TestClientRequestContract(t *testing.T) {
	type request struct{ method, path, query, auth, body string }
	calls := make(chan request, 1)
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		calls <- request{r.Method, r.URL.Path, r.URL.RawQuery, r.Header.Get("Authorization"), string(b)}
		_, _ = io.WriteString(w, "{}")
	}, true)
	ctx := context.Background()
	tests := []struct {
		name, method, path, query, body string
		call                            func() (json.RawMessage, error)
	}{
		{"tasks", "GET", "/api/tasks", "", "", func() (json.RawMessage, error) { return client.ListTasks(ctx) }},
		{"task", "GET", "/api/tasks/42", "", "", func() (json.RawMessage, error) { return client.GetTask(ctx, "42") }},
		{"assets", "GET", "/api/assets", "limit=50&offset=50&task_id=42", "", func() (json.RawMessage, error) { return client.ListAssets(ctx, "42", 2) }},
		{"findings", "GET", "/api/exploration/findings", "limit=50&page=2&task_id=42", "", func() (json.RawMessage, error) { return client.ListFindings(ctx, "42", 2) }},
		{"finding", "GET", "/api/exploration/findings/99", "", "", func() (json.RawMessage, error) { return client.GetFinding(ctx, "99") }},
		{"activity", "GET", "/api/exploration/activity", "limit=100&since=100&task=42", "", func() (json.RawMessage, error) { return client.Activity(ctx, "42", "100") }},
		{"retests", "GET", "/api/exploration/findings/99/retests", "", "", func() (json.RawMessage, error) { return client.ListRetests(ctx, "99") }},
		{"create", "POST", "/api/tasks", "", `{"description":"local lab","goal":"inventory","timeout_seconds":1800}`, func() (json.RawMessage, error) {
			return client.CreateTask(ctx, CreateTaskRequest{Description: " local lab ", Goal: "inventory", TimeoutSeconds: 1800})
		}},
		{"pause", "POST", "/api/tasks/42/control", "", `{"action":"pause"}`, func() (json.RawMessage, error) {
			return client.ControlTask(ctx, "42", ControlTaskRequest{Action: "pause"})
		}},
		{"retest", "POST", "/api/exploration/findings/99/retests", "", `{"notes":"fixed build"}`, func() (json.RawMessage, error) {
			return client.StartRetest(ctx, "99", RetestRequest{Notes: "fixed build"})
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := tt.call(); err != nil {
				t.Fatal(err)
			}
			got := <-calls
			if got.method != tt.method || got.path != "/engine"+tt.path || got.query != tt.query || got.body != tt.body || got.auth != "Bearer test-token" {
				t.Fatalf("unexpected request: %+v", got)
			}
		})
	}
}
func TestClientValidatesBeforeNetwork(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) { t.Error("invalid call reached upstream") }, true)
	for _, id := range []string{"", "0", "-1", "01", "../settings", "%2F", "1/../../auth", "9999999999999999999999"} {
		if _, err := client.GetTask(context.Background(), id); !errors.Is(err, ErrInvalidRequest) {
			t.Errorf("id %q: %v", id, err)
		}
	}
	if _, err := client.ControlTask(context.Background(), "1", ControlTaskRequest{Action: "delete"}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatal(err)
	}
	if _, err := client.CreateTask(context.Background(), CreateTaskRequest{Description: "x", Goal: " "}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatal(err)
	}
	if _, err := client.ListAssets(context.Background(), "1", 0); !errors.Is(err, ErrInvalidRequest) {
		t.Fatal(err)
	}
	if _, err := client.StartRetest(context.Background(), "1", RetestRequest{Notes: strings.Repeat("中", 4001)}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatal(err)
	}
}
func TestClientBlocksWritesAndDisabled(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) { t.Error("blocked call reached upstream") }, false)
	if _, err := client.CreateTask(context.Background(), CreateTaskRequest{Description: "d", Goal: "g"}); !errors.Is(err, ErrWritesDisabled) {
		t.Fatal(err)
	}
	var disabled *Client
	if _, err := disabled.ListTasks(context.Background()); !errors.Is(err, ErrDisabled) {
		t.Fatal(err)
	}
}
func TestClientDoesNotFollowRedirectsOrLeakErrors(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/engine/api/tasks" {
			t.Error("followed redirect")
		}
		w.Header().Set("Location", "/api/settings")
		w.WriteHeader(302)
		_, _ = io.WriteString(w, "secret-evidence")
	}, false)
	_, err := client.ListTasks(context.Background())
	var upstream *HTTPError
	if !errors.As(err, &upstream) || upstream.StatusCode != 302 || strings.Contains(err.Error(), "secret") {
		t.Fatal(err)
	}
	client.httpClient.Transport = errorTransport{}
	_, err = client.ListTasks(context.Background())
	if !errors.Is(err, ErrUnavailable) || strings.Contains(err.Error(), "secret") {
		t.Fatal(err)
	}
}

type errorTransport struct{}

func (errorTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("secret internal URL")
}
func TestClientRejectsInvalidAndOversizedJSON(t *testing.T) {
	for _, body := range []string{"not json", "{}" + strings.Repeat(" ", MaxResponseBytes)} {
		client := testClient(t, func(w http.ResponseWriter, r *http.Request) { _, _ = io.WriteString(w, body) }, false)
		if _, err := client.ListTasks(context.Background()); !errors.Is(err, ErrResponse) {
			t.Fatal(err)
		}
	}
}
func TestClientHealthAndTimeout(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"ok":true,"service":"other"}`)
	}, false)
	if _, err := client.Health(context.Background()); !errors.Is(err, ErrResponse) {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := client.ListTasks(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
func TestClientRejectsCredentialURLs(t *testing.T) {
	for _, address := range []string{"", "file:///tmp", "http://u:p@127.0.0.1", "https://example.test?key=secret", "https://example.test/#token", "https://example.test/api", "http://example.test/a/../b"} {
		_, err := NewClient(Config{Enabled: true, BaseURL: address, APIKey: "token"})
		if err == nil {
			t.Fatalf("accepted %q", address)
		}
		var urlError *url.Error
		if errors.As(err, &urlError) || strings.Contains(err.Error(), "secret") {
			t.Error("configuration error contains URL")
		}
	}
}

func TestTaskIDExtractsOnlyRemoteID(t *testing.T) {
	if got := TaskID(json.RawMessage(`{"id":"task-42","api_key":"must-not-be-retained"}`)); got != "task-42" {
		t.Fatalf("got %q", got)
	}
	if got := TaskID(json.RawMessage(`{"id":123}`)); got != "" {
		t.Fatalf("numeric id accepted: %q", got)
	}
}
