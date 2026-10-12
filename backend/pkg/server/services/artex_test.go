package services

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/gin-gonic/gin"
	"pentagi/pkg/integrations/artex"
)

func artexTestRouter(t *testing.T, writes bool, privs []string, upstream http.HandlerFunc) (*gin.Engine, *atomic.Int32) {
	t.Helper()
	calls := new(atomic.Int32)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1); upstream(w, r) }))
	t.Cleanup(server.Close)
	client, err := artex.NewClient(artex.Config{Enabled: true, BaseURL: server.URL, APIKey: "upstream-test-token", AllowWrites: writes})
	if err != nil {
		t.Fatal(err)
	}
	svc := NewArtexService(client)
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set("uid", uint64(1)); c.Set("prm", privs); c.Next() })
	router.GET("/status", svc.Status)
	router.GET("/tasks", svc.ListTasks)
	router.POST("/tasks", svc.CreateTask)
	router.GET("/assets", svc.ListAssets)
	router.POST("/tasks/:id/control", svc.ControlTask)
	router.POST("/findings/:id/retest", svc.StartRetest)
	return router, calls
}
func artexRequest(router http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Language", "zh-CN")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}
func TestArtexRejectsRegularUsersAndReadOnlyWrites(t *testing.T) {
	cases := []struct {
		name               string
		privs              []string
		writes             bool
		method, path, body string
		code               string
	}{
		{"reader", []string{"settings.view", "flows.view"}, true, "GET", "/tasks", "", "Artex.Forbidden"},
		{"creator", []string{"flows.create"}, true, "POST", "/tasks", `{"description":"d","goal":"g"}`, "Artex.Forbidden"},
		{"read only", []string{"flows.admin", "flows.create"}, false, "POST", "/tasks", `{"description":"d","goal":"g"}`, "Artex.ReadOnly"},
		{"no edit", []string{"flows.admin", "flows.create"}, true, "POST", "/tasks/1/control", `{"action":"pause"}`, "Artex.Forbidden"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			router, calls := artexTestRouter(t, tt.writes, tt.privs, func(http.ResponseWriter, *http.Request) { t.Error("unauthorized call") })
			rec := artexRequest(router, tt.method, tt.path, tt.body)
			if rec.Code != 403 || !strings.Contains(rec.Body.String(), tt.code) || calls.Load() != 0 {
				t.Fatalf("%d %s calls=%d", rec.Code, rec.Body, calls.Load())
			}
		})
	}
}
func TestArtexStatusChecksProtectedEndpoint(t *testing.T) {
	router, calls := artexTestRouter(t, true, []string{"flows.admin"}, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/health" {
			_, _ = io.WriteString(w, `{"ok":true,"service":"artex","version":"test"}`)
			return
		}
		w.WriteHeader(401)
		_, _ = io.WriteString(w, "upstream-test-token secret-evidence")
	})
	rec := artexRequest(router, "GET", "/status", "")
	if rec.Code != 200 || calls.Load() != 2 || !strings.Contains(rec.Body.String(), "Artex.Authentication") || !strings.Contains(rec.Body.String(), `"healthy":false`) || strings.Contains(rec.Body.String(), "secret") {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
	rec = artexRequest(router, "GET", "/tasks", "")
	if rec.Code != 502 || !strings.Contains(rec.Body.String(), "Artex.Authentication") {
		t.Fatalf("upstream auth must not log out PentAGI: %d %s", rec.Code, rec.Body)
	}
}
func TestArtexMutationValidationAndProtocol(t *testing.T) {
	router, calls := artexTestRouter(t, true, []string{"flows.admin", "flows.create", "flows.edit"}, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/exploration/findings/123/retests" {
			t.Error(r.URL.Path)
		}
		var req map[string]string
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req["notes"] != "local fix" || len(req) != 1 {
			t.Error(req)
		}
		_, _ = io.WriteString(w, `{"created":true,"retest":{"id":9007199254740993}}`)
	})
	for _, body := range []string{`{"input":"wrong field"}`, `{} {}`, `{"notes":"` + strings.Repeat("x", 70<<10) + `"}`} {
		rec := artexRequest(router, "POST", "/findings/123/retest", body)
		if rec.Code != 400 {
			t.Fatalf("%d %s", rec.Code, rec.Body)
		}
	}
	if calls.Load() != 0 {
		t.Fatal("invalid request forwarded")
	}
	rec := artexRequest(router, "POST", "/findings/123/retest", `{"notes":"local fix"}`)
	if rec.Code != 200 || calls.Load() != 1 || !strings.Contains(rec.Body.String(), "9007199254740993") {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
}
func TestArtexRequiresExplicitTaskScope(t *testing.T) {
	router, calls := artexTestRouter(t, false, []string{"flows.admin"}, func(w http.ResponseWriter, r *http.Request) { t.Error("unscoped upstream call") })
	for _, path := range []string{"/assets", "/assets?task=active", "/assets?task=1&page=-1"} {
		rec := artexRequest(router, "GET", path, "")
		if rec.Code != 400 {
			t.Fatalf("%s: %d", path, rec.Code)
		}
	}
	if calls.Load() != 0 {
		t.Fatal("invalid scope forwarded")
	}
}

func TestArtexSubmissionCorrelationIsBoundedAndReversible(t *testing.T) {
	marker := "[pentagi-run:12345678]"
	if got := appendCorrelation("goal", marker, 80); got != "goal [pentagi-run:12345678]" {
		t.Fatalf("unexpected correlation %q", got)
	}
	long := appendCorrelation(strings.Repeat("x", 200), marker, 40)
	if len([]rune(long)) > 40 || !strings.HasSuffix(long, marker) {
		t.Fatalf("bounded correlation lost marker: %q", long)
	}
	if !strings.Contains(long, marker) {
		t.Fatal("marker missing")
	}
}
