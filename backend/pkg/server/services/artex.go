package services

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"pentagi/pkg/database"
	"pentagi/pkg/integrations/artex"
	"pentagi/pkg/server/response"
)

type ArtexService struct {
	client  *artex.Client
	sharing *artex.ModelSharing
	runs    *artex.SQLRunStore
	native  database.Querier
}

func NewArtexService(client *artex.Client, sharing ...*artex.ModelSharing) *ArtexService {
	s := &ArtexService{client: client}
	if len(sharing) > 0 {
		s.sharing = sharing[0]
	}
	return s
}

func (s *ArtexService) SetRunStore(store *artex.SQLRunStore)        { s.runs = store }
func (s *ArtexService) SetNativeFlowSource(source database.Querier) { s.native = source }

func (s *ArtexService) ListRuns(c *gin.Context) {
	if !s.canView(c) {
		return
	}
	if s.runs == nil {
		response.Error(c, artexError(artex.ErrStoreUnavailable), nil)
		return
	}
	if raw, err := s.client.ListTasks(c.Request.Context()); err == nil {
		for _, task := range artexTaskSummaries(raw) {
			_ = s.runs.UpdateRemote(c.Request.Context(), task.ID, task.Status, artex.NormalizeTaskStatus(task.Status))
		}
	}
	items, err := s.runs.List(c.Request.Context(), int64(c.GetUint64("uid")))
	if err != nil {
		response.Error(c, artexError(err), nil)
		return
	}
	response.Success(c, http.StatusOK, items)
}

func (s *ArtexService) ListUnifiedRuns(c *gin.Context) {
	uid := int64(c.GetUint64("uid"))
	privs := c.GetStringSlice("prm")
	admin := slices.Contains(privs, "flows.admin")
	if !admin && !slices.Contains(privs, "flows.view") {
		response.Error(c, response.NewHttpError(403, "Runs.Forbidden", "Flow viewing permission required"), nil)
		return
	}
	items := make([]artex.UnifiedRun, 0)
	if s.native != nil {
		var flows []database.Flow
		var err error
		if admin {
			flows, err = s.native.GetFlows(c.Request.Context())
		} else {
			flows, err = s.native.GetUserFlows(c.Request.Context(), uid)
		}
		if err != nil {
			response.Error(c, response.NewHttpError(500, "Runs.NativeUnavailable", "Unable to load PentAGI executions"), nil)
			return
		}
		for _, flow := range flows {
			created, updated := time.Time{}, time.Time{}
			if flow.CreatedAt.Valid {
				created = flow.CreatedAt.Time
			}
			if flow.UpdatedAt.Valid {
				updated = flow.UpdatedAt.Time
			}
			items = append(items, artex.UnifiedRun{ID: "pentagi:" + strconv.FormatInt(flow.ID, 10), Engine: "pentagi", EngineTaskID: strconv.FormatInt(flow.ID, 10), Title: flow.Title, Status: normalizeNativeRunStatus(string(flow.Status)), RawStatus: string(flow.Status), Model: flow.Model, CreatedAt: created, UpdatedAt: updated, DetailPath: "/flows/" + strconv.FormatInt(flow.ID, 10)})
		}
	}
	if admin && s.runs != nil && s.client.Enabled() {
		if raw, err := s.client.ListTasks(c.Request.Context()); err == nil {
			for _, task := range artexTaskSummaries(raw) {
				_ = s.runs.UpdateRemote(c.Request.Context(), task.ID, task.Status, artex.NormalizeTaskStatus(task.Status))
			}
		}
		runs, err := s.runs.ListAll(c.Request.Context())
		if err != nil {
			response.Error(c, artexError(err), nil)
			return
		}
		for _, run := range runs {
			items = append(items, artex.UnifiedRun{ID: "artex:" + run.ID, Engine: "artex", EngineTaskID: run.EngineTaskID, Title: run.Name, Status: run.Status, RawStatus: run.RawStatus, SubmitState: run.SubmitState, BindingID: run.BindingID, BindingRevision: run.BindingRevision, CreatedAt: run.CreatedAt, UpdatedAt: run.UpdatedAt, DetailPath: "/artex"})
		}
	}
	response.Success(c, http.StatusOK, gin.H{"items": items, "total": len(items)})
}

func normalizeNativeRunStatus(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "created":
		return "queued"
	case "running":
		return "running"
	case "waiting":
		return "waiting"
	case "finished":
		return "completed"
	case "failed":
		return "failed"
	default:
		return "unknown"
	}
}

type artexTaskSummary struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func artexTaskSummaries(raw json.RawMessage) []artexTaskSummary {
	items, err := artex.ParseTaskSummaries(raw)
	if err != nil {
		return nil
	}
	result := make([]artexTaskSummary, len(items))
	for i, item := range items {
		result[i] = artexTaskSummary{ID: item.ID, Status: item.Status}
	}
	return result
}

func (s *ArtexService) Status(c *gin.Context) {
	if !s.canView(c) {
		return
	}
	status := gin.H{"enabled": s.client.Enabled(), "healthy": false, "canCreate": false, "canControl": false}
	status["modelSharingAvailable"] = s.sharing != nil
	if s.client == nil {
		status["errorCode"] = "Artex.Configuration"
	}
	if !s.client.Enabled() {
		response.Success(c, http.StatusOK, status)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	health, err := s.client.Health(ctx)
	if err == nil {
		_, err = s.client.ListTasks(ctx)
	} // Health alone is public; verify the login token too.
	if err != nil {
		status["errorCode"] = artexError(err).Code()
	} else {
		var version struct {
			Version string `json:"version"`
		}
		_ = json.Unmarshal(health, &version)
		status["version"] = version.Version
		status["healthy"] = true
		status["canCreate"] = s.sharing != nil && s.client.AllowsWrites() && slices.Contains(c.GetStringSlice("prm"), "flows.create")
		status["canShareModels"] = s.sharing != nil && s.client.AllowsWrites() && slices.Contains(c.GetStringSlice("prm"), "settings.providers.edit")
		status["canControl"] = s.client.AllowsWrites() && slices.Contains(c.GetStringSlice("prm"), "flows.edit")
	}
	response.Success(c, http.StatusOK, status)
}
func (s *ArtexService) ListTasks(c *gin.Context) { s.proxyRead(c, s.client.ListTasks) }
func (s *ArtexService) GetTask(c *gin.Context) {
	s.proxyRead(c, func(ctx context.Context) (json.RawMessage, error) { return s.client.GetTask(ctx, c.Param("id")) })
}
func (s *ArtexService) ListAssets(c *gin.Context) {
	s.readEvidence(c, func(ctx context.Context) (json.RawMessage, error) {
		return s.client.ListAssets(ctx, c.Query("task"), artexPage(c))
	}, c.Query("task"), "asset", "list", "")
}
func (s *ArtexService) ListFindings(c *gin.Context) {
	s.readEvidence(c, func(ctx context.Context) (json.RawMessage, error) {
		return s.client.ListFindings(ctx, c.Query("task"), artexPage(c))
	}, c.Query("task"), "finding", "list", "")
}
func (s *ArtexService) GetFinding(c *gin.Context) {
	s.readEvidence(c, func(ctx context.Context) (json.RawMessage, error) { return s.client.GetFinding(ctx, c.Param("id")) }, "", "finding", "detail", "")
}
func (s *ArtexService) Activity(c *gin.Context) {
	s.readEvidence(c, func(ctx context.Context) (json.RawMessage, error) {
		return s.client.Activity(ctx, c.Query("task"), c.DefaultQuery("since", "0"))
	}, c.Query("task"), "activity", "list", "")
}
func (s *ArtexService) ListRetests(c *gin.Context) {
	if !s.canView(c) {
		return
	}
	// Resolve the source finding directly; parallel frontend requests must not
	// depend on another request having already written a snapshot.
	finding, err := s.client.GetFinding(c.Request.Context(), c.Param("id"))
	if err != nil {
		s.writeResult(c, nil, err)
		return
	}
	s.readEvidence(c, func(ctx context.Context) (json.RawMessage, error) { return s.client.ListRetests(ctx, c.Param("id")) }, artex.EvidenceTaskID(finding), "retest", "list", c.Param("id"))
}
func (s *ArtexService) CreateTask(c *gin.Context) {
	if !s.canWrite(c, "flows.create") {
		return
	}
	var input struct {
		Name           string `json:"name"`
		Description    string `json:"description"`
		Goal           string `json:"goal"`
		TimeoutSeconds int    `json:"timeout_seconds"`
		BindingID      string `json:"bindingId"`
	}
	if !bindArtexRequest(c, &input) {
		return
	}
	idempotencyKey := strings.TrimSpace(c.GetHeader("Idempotency-Key"))
	if utf8.RuneCountInString(idempotencyKey) > 128 {
		s.writeResult(c, nil, artex.ErrInvalidRequest)
		return
	}
	replayExisting := func(run *artex.Run) bool {
		c.Header("X-PentAGI-Run-ID", run.ID)
		switch run.SubmitState {
		case "created":
			if run.EngineTaskID == "" {
				s.writeResult(c, nil, artex.ErrSyncUncertain)
				return true
			}
			body, err := s.client.GetTask(c.Request.Context(), run.EngineTaskID)
			s.writeMutation(c, "replay_create_task", body, err)
			return true
		case "pending", "uncertain":
			s.writeResult(c, nil, artex.ErrSyncUncertain)
			return true
		default:
			s.writeResult(c, nil, artex.ErrRunUnresolved)
			return true
		}
	}
	if idempotencyKey != "" && s.runs != nil {
		run, findErr := s.runs.FindByCorrelation(c.Request.Context(), int64(c.GetUint64("uid")), idempotencyKey)
		if findErr == nil {
			replayExisting(run)
			return
		}
		if !errors.Is(findErr, artex.ErrRunNotFound) {
			s.writeResult(c, nil, findErr)
			return
		}
	}
	if input.BindingID == "" {
		s.writeResult(c, nil, artex.ErrBindingMissing)
		return
	}
	binding, err := s.sharing.ForTask(c.Request.Context(), int64(c.GetUint64("uid")), input.BindingID)
	if err != nil {
		s.writeResult(c, nil, err)
		return
	}
	var run *artex.Run
	if s.runs != nil {
		run, err = s.runs.CreatePending(c.Request.Context(), int64(c.GetUint64("uid")), artex.RunRequest{Name: input.Name, Description: input.Description, Goal: input.Goal, TimeoutSeconds: input.TimeoutSeconds, BindingID: binding.ID, BindingRevision: binding.Revision, CorrelationID: idempotencyKey})
		if err != nil {
			if idempotencyKey != "" {
				if existing, findErr := s.runs.FindByCorrelation(c.Request.Context(), int64(c.GetUint64("uid")), idempotencyKey); findErr == nil {
					replayExisting(existing)
					return
				}
			}
			s.writeResult(c, nil, err)
			return
		}
	}
	remoteName, remoteDescription := input.Name, input.Description
	if run != nil {
		marker := artex.CorrelationMarker(run.ID)
		remoteName = appendCorrelation(remoteName, marker, 200)
		remoteDescription = appendCorrelation(remoteDescription, marker, 8000)
		c.Header("X-PentAGI-Run-ID", run.ID)
	}
	body, err := s.client.CreateTask(c.Request.Context(), artex.CreateTaskRequest{Name: remoteName, Description: remoteDescription, Goal: input.Goal, TimeoutSeconds: input.TimeoutSeconds, LLMProfileIDs: []int64{binding.RemoteID}})
	if s.runs != nil && run != nil {
		if err != nil {
			if submissionIsDefinitelyRejected(err) {
				_ = s.runs.MarkFailed(context.Background(), run.ID)
			} else {
				_ = s.runs.MarkUncertain(context.Background(), run.ID)
			}
		} else if remoteID := artex.TaskID(body); remoteID != "" {
			_ = s.runs.MarkCreated(context.Background(), run.ID, remoteID)
		} else {
			_ = s.runs.MarkUncertain(context.Background(), run.ID)
		}
	}
	s.writeMutation(c, "create_task", body, err)
}

func appendCorrelation(value, marker string, limit int) string {
	value = strings.TrimSpace(value)
	if strings.Contains(value, marker) {
		return value
	}
	separator := " "
	if value == "" {
		separator = ""
	}
	runes := []rune(value + separator + marker)
	if len(runes) <= limit {
		return string(runes)
	}
	markerRunes := []rune(marker)
	if len(markerRunes) >= limit {
		return string(markerRunes[:limit])
	}
	prefix := limit - len(markerRunes) - len([]rune(separator))
	if prefix < 0 {
		prefix = 0
	}
	return string(runes[:prefix]) + separator + marker
}

func submissionIsDefinitelyRejected(err error) bool {
	if errors.Is(err, artex.ErrInvalidRequest) || errors.Is(err, artex.ErrWritesDisabled) || errors.Is(err, artex.ErrDisabled) {
		return true
	}
	var upstream *artex.HTTPError
	if errors.As(err, &upstream) {
		return upstream.StatusCode >= 400 && upstream.StatusCode < 500 && upstream.StatusCode != 429
	}
	return false
}

func (s *ArtexService) ReconcileRun(c *gin.Context) {
	if !s.canWrite(c, "flows.edit") {
		return
	}
	if s.runs == nil {
		response.Error(c, artexError(artex.ErrStoreUnavailable), nil)
		return
	}
	run, err := s.runs.Get(c.Request.Context(), int64(c.GetUint64("uid")), c.Param("runID"))
	if err != nil {
		response.Error(c, artexError(err), nil)
		return
	}
	if run.SubmitState == "created" && run.EngineTaskID != "" {
		response.Success(c, http.StatusOK, run)
		return
	}
	if run.SubmitState == "failed" {
		response.Error(c, artexError(artex.ErrRunUnresolved), nil)
		return
	}
	raw, err := s.client.ListTasks(c.Request.Context())
	if err != nil {
		response.Error(c, artexError(err), nil)
		return
	}
	tasks, err := artex.ParseTaskSummaries(raw)
	if err != nil {
		response.Error(c, artexError(err), nil)
		return
	}
	matches := make([]artex.TaskSummary, 0, 1)
	for _, task := range tasks {
		if artex.HasCorrelation(task.Name, run.ID) || artex.HasCorrelation(task.Description, run.ID) {
			matches = append(matches, task)
		}
	}
	if len(matches) == 0 {
		response.Error(c, artexError(artex.ErrRunUnresolved), nil)
		return
	}
	if len(matches) > 1 {
		response.Error(c, artexError(artex.ErrRunAmbiguous), nil)
		return
	}
	if err := s.runs.MarkCreated(c.Request.Context(), run.ID, matches[0].ID); err != nil {
		response.Error(c, artexError(err), nil)
		return
	}
	_ = s.runs.UpdateRemote(c.Request.Context(), matches[0].ID, matches[0].Status, artex.NormalizeTaskStatus(matches[0].Status))
	updated, err := s.runs.Get(c.Request.Context(), int64(c.GetUint64("uid")), run.ID)
	if err != nil {
		response.Error(c, artexError(err), nil)
		return
	}
	response.Success(c, http.StatusOK, updated)
}

func (s *ArtexService) ControlTask(c *gin.Context) {
	if !s.canWrite(c, "flows.edit") {
		return
	}
	var input artex.ControlTaskRequest
	if !bindArtexRequest(c, &input) {
		return
	}
	body, err := s.client.ControlTask(c.Request.Context(), c.Param("id"), input)
	s.writeMutation(c, "control_task", body, err)
}
func (s *ArtexService) StartRetest(c *gin.Context) {
	if !s.canWrite(c, "flows.edit") {
		return
	}
	var input artex.RetestRequest
	if !bindArtexRequest(c, &input) {
		return
	}
	body, err := s.client.StartRetest(c.Request.Context(), c.Param("id"), input)
	s.writeMutation(c, "start_retest", body, err)
}
func artexPage(c *gin.Context) int {
	raw := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(raw)
	if err != nil {
		return 0
	}
	return page
}
func bindArtexRequest(c *gin.Context, dst any) bool {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 64<<10)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(dst)
	if err == nil {
		if decoder.Decode(new(any)) != io.EOF {
			err = artex.ErrInvalidRequest
		}
	}
	if err != nil {
		response.Error(c, artexError(artex.ErrInvalidRequest), nil)
		return false
	}
	return true
}
func (s *ArtexService) proxyRead(c *gin.Context, call func(context.Context) (json.RawMessage, error)) {
	if !s.canView(c) {
		return
	}
	body, err := call(c.Request.Context())
	s.writeResult(c, body, err)
}
func (s *ArtexService) writeMutation(c *gin.Context, operation string, body json.RawMessage, err error) {
	outcome := "success"
	if err != nil {
		outcome = "error"
	}
	logrus.WithFields(logrus.Fields{"component": "artex-adapter", "operation": operation, "uid": c.GetUint64("uid"), "outcome": outcome}).Info("ARTEX operation")
	s.writeResult(c, body, err)
}
func (s *ArtexService) writeResult(c *gin.Context, body json.RawMessage, err error) {
	if err != nil {
		response.Error(c, artexError(err), nil)
		return
	}
	response.Success(c, http.StatusOK, body) // Preserve numbers and the original JSON shape.
}
func (s *ArtexService) canView(c *gin.Context) bool {
	// One upstream JWT has instance-wide access. Until task ownership is mapped,
	// restrict this adapter to administrators rather than leak other users' tasks.
	if c.GetUint64("uid") > 0 && slices.Contains(c.GetStringSlice("prm"), "flows.admin") {
		return true
	}
	response.Error(c, response.NewHttpError(403, "Artex.Forbidden", "ARTEX requires flow administrator access"), nil)
	return false
}
func (s *ArtexService) canWrite(c *gin.Context, permission string) bool {
	if !s.canView(c) {
		return false
	}
	if !s.client.Enabled() {
		response.Error(c, artexError(artex.ErrDisabled), nil)
		return false
	}
	if !s.client.AllowsWrites() {
		response.Error(c, artexError(artex.ErrWritesDisabled), nil)
		return false
	}
	if !slices.Contains(c.GetStringSlice("prm"), permission) {
		response.Error(c, response.NewHttpError(403, "Artex.Forbidden", "ARTEX operation is not permitted"), nil)
		return false
	}
	return true
}
func artexError(err error) *response.HttpError {
	switch {
	case errors.Is(err, artex.ErrSourceUnavailable):
		return response.NewHttpError(404, "Artex.ProviderUnavailable", "The selected model provider is unavailable or not owned by this user")
	case errors.Is(err, artex.ErrUnsupportedProvider):
		return response.NewHttpError(400, "Artex.ProviderUnsupported", "This provider protocol is not supported by ARTEX sharing")
	case errors.Is(err, artex.ErrBindingMissing):
		return response.NewHttpError(400, "Artex.BindingRequired", "Select and synchronize an owned model provider first")
	case errors.Is(err, artex.ErrProfileDrift):
		return response.NewHttpError(409, "Artex.ProfileChanged", "The managed ARTEX profile was changed externally; reconcile it before running tasks")
	case errors.Is(err, artex.ErrSyncUncertain):
		return response.NewHttpError(409, "Artex.SyncUncertain", "Model synchronization is pending or uncertain; reconcile the same configuration before continuing")
	case errors.Is(err, artex.ErrCredentialChanged):
		return response.NewHttpError(409, "Artex.CredentialChanged", "The provider key changed; synchronize the managed binding with Rotate key")
	case errors.Is(err, artex.ErrEndpointChanged):
		return response.NewHttpError(409, "Artex.EndpointChanged", "The provider endpoint changed; create a new model configuration version")
	case errors.Is(err, artex.ErrBindingRevoked):
		return response.NewHttpError(409, "Artex.BindingRevoked", "The model binding is revoked")
	case errors.Is(err, artex.ErrStoreUnavailable):
		return response.NewHttpError(503, "Artex.BindingStorage", "Model sharing is unavailable; check database migrations and signing configuration")
	case errors.Is(err, artex.ErrSnapshotNotFound):
		return response.NewHttpError(404, "Artex.SnapshotNotFound", "The requested evidence snapshot does not exist for this run")
	case errors.Is(err, artex.ErrRunNotFound):
		return response.NewHttpError(404, "Artex.RunNotFound", "The local execution record was not found or is not owned by this user")
	case errors.Is(err, artex.ErrRunUnresolved):
		return response.NewHttpError(409, "Artex.RunUnresolved", "No unique remote task matched this submission; inspect ARTEX before retrying")
	case errors.Is(err, artex.ErrRunAmbiguous):
		return response.NewHttpError(409, "Artex.RunAmbiguous", "Multiple remote tasks matched this submission; resolve the mapping manually")
	case errors.Is(err, artex.ErrDisabled):
		return response.NewHttpError(503, "Artex.Disabled", "ARTEX is not configured or enabled")
	case errors.Is(err, artex.ErrWritesDisabled):
		return response.NewHttpError(403, "Artex.ReadOnly", "ARTEX is connected in read-only mode")
	case errors.Is(err, artex.ErrInvalidRequest):
		return response.NewHttpError(400, "Artex.InvalidRequest", "Check the task ID, page and input limits")
	case errors.Is(err, artex.ErrResponse):
		return response.NewHttpError(502, "Artex.Response", "ARTEX returned an invalid or oversized response")
	case errors.Is(err, artex.ErrTimeout), errors.Is(err, context.DeadlineExceeded):
		return response.NewHttpError(504, "Artex.Timeout", "ARTEX request timed out")
	}
	var upstream *artex.HTTPError
	if errors.As(err, &upstream) {
		switch upstream.StatusCode {
		case 401, 403:
			return response.NewHttpError(502, "Artex.Authentication", "ARTEX login token is invalid or expired; update ARTEX_API_KEY and restart PentAGI")
		case 404:
			return response.NewHttpError(404, "Artex.NotFound", "ARTEX task or finding was not found")
		case 409:
			return response.NewHttpError(409, "Artex.Conflict", "ARTEX cannot perform this action in the current state; check task and retester settings")
		case 400, 422:
			return response.NewHttpError(400, "Artex.Rejected", "ARTEX rejected the request; check its task and model settings")
		case 429:
			return response.NewHttpError(503, "Artex.Busy", "ARTEX is busy; try again later")
		}
	}
	return response.NewHttpError(502, "Artex.Unavailable", "Unable to reach ARTEX; check the service address and connection")
}
