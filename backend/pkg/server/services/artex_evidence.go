package services

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"pentagi/pkg/integrations/artex"
	"pentagi/pkg/server/response"
)

func (s *ArtexService) readEvidence(c *gin.Context, call func(context.Context) (json.RawMessage, error), taskID, kind, representation, parentID string) {
	if !s.canView(c) {
		return
	}
	body, err := call(c.Request.Context())
	if err != nil {
		s.writeResult(c, nil, err)
		return
	}
	if taskID == "" && kind == "finding" {
		taskID = artex.EvidenceTaskID(body)
	}
	state := "unmanaged"
	if s.runs != nil && taskID != "" {
		run, findErr := s.runs.FindByRemote(c.Request.Context(), taskID)
		if findErr == nil {
			records, archiveErr := artex.EvidenceRecords(body, kind, representation, parentID)
			for _, record := range records {
				if archiveErr != nil {
					break
				}
				_, archiveErr = s.runs.SaveEvidence(c.Request.Context(), run.UserID, run.ID, taskID, record)
			}
			if archiveErr != nil {
				state = "unavailable"
				// Do not log upstream bodies, credential fields, or database error values.
				logrus.WithFields(logrus.Fields{"component": "artex-evidence", "kind": kind, "run_id": run.ID}).Warn("Evidence archive incomplete")
			} else {
				state = "stored"
			}
		} else if !errors.Is(findErr, artex.ErrRunNotFound) {
			state = "unavailable"
		}
	} else if s.runs == nil {
		state = "disabled"
	}
	c.Header("X-PentAGI-Evidence-State", state)
	s.writeResult(c, body, nil)
}

func (s *ArtexService) ListSnapshots(c *gin.Context) {
	if !s.canView(c) {
		return
	}
	limit, err := artex.EvidenceLimit(c.Query("limit"))
	if err != nil {
		s.writeResult(c, nil, err)
		return
	}
	// This entire instance is admin-only. Ordinary users never reach the store.
	page, err := s.runs.SnapshotPage(c.Request.Context(), c.Param("runID"), c.Query("kind"), c.Query("before"), limit)
	if err != nil {
		s.writeResult(c, nil, err)
		return
	}
	response.Success(c, http.StatusOK, page)
}

// ListTaskSnapshots resolves the remote ARTEX task to its locally owned run
// record before exposing archived evidence. The UI only knows the remote task
// ID; it must never be allowed to guess a local run ID.
func (s *ArtexService) ListTaskSnapshots(c *gin.Context) {
	if !s.canView(c) {
		return
	}
	if s.runs == nil {
		s.writeResult(c, nil, artex.ErrStoreUnavailable)
		return
	}
	run, err := s.runs.FindByRemote(c.Request.Context(), c.Param("id"))
	if err != nil {
		s.writeResult(c, nil, err)
		return
	}
	limit, err := artex.EvidenceLimit(c.Query("limit"))
	if err != nil {
		s.writeResult(c, nil, err)
		return
	}
	page, err := s.runs.SnapshotPage(c.Request.Context(), run.ID, c.Query("kind"), c.Query("before"), limit)
	if err != nil {
		s.writeResult(c, nil, err)
		return
	}
	response.Success(c, http.StatusOK, page)
}

func (s *ArtexService) GetSnapshot(c *gin.Context) {
	if !s.canView(c) {
		return
	}
	item, err := s.runs.GetSnapshot(c.Request.Context(), c.Param("runID"), c.Param("snapshotID"))
	if err != nil {
		s.writeResult(c, nil, err)
		return
	}
	c.Header("Cache-Control", "no-store")
	response.Success(c, http.StatusOK, item)
}

func (s *ArtexService) GetTaskSnapshot(c *gin.Context) {
	if !s.canView(c) {
		return
	}
	if s.runs == nil {
		s.writeResult(c, nil, artex.ErrStoreUnavailable)
		return
	}
	run, err := s.runs.FindByRemote(c.Request.Context(), c.Param("id"))
	if err != nil {
		s.writeResult(c, nil, err)
		return
	}
	item, err := s.runs.GetSnapshot(c.Request.Context(), run.ID, c.Param("snapshotID"))
	if err != nil {
		s.writeResult(c, nil, err)
		return
	}
	c.Header("Cache-Control", "no-store")
	response.Success(c, http.StatusOK, item)
}
