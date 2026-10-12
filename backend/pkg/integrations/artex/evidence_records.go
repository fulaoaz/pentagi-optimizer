package artex

import (
	"encoding/json"
	"strconv"
)

// EvidenceRecords extracts stable source identities, never page numbers. The
// summary and detail representations are versioned separately because ARTEX
// intentionally omits full reports from finding lists.
func EvidenceRecords(body json.RawMessage, kind, representation, parentID string) ([]EvidenceRecord, error) {
	if !validEvidenceKind(kind) || (representation != "list" && representation != "detail") {
		return nil, ErrInvalidRequest
	}
	var records []json.RawMessage
	if representation == "detail" {
		records = []json.RawMessage{body}
	} else {
		var envelope map[string]json.RawMessage
		if json.Unmarshal(body, &envelope) != nil || envelope == nil {
			return nil, ErrResponse
		}
		key := "items"
		switch kind {
		case "asset":
			key = "assets"
		case "retest":
			key = "retests"
		}
		raw, ok := envelope[key]
		if !ok {
			return nil, ErrResponse
		}
		if json.Unmarshal(raw, &records) != nil {
			return nil, ErrResponse
		}
	}
	result := make([]EvidenceRecord, 0, len(records))
	for _, raw := range records {
		var item map[string]json.RawMessage
		if json.Unmarshal(raw, &item) != nil || item == nil {
			return nil, ErrResponse
		}
		key := "id"
		if kind == "activity" {
			key = "seq"
		}
		id := evidenceID(item[key])
		if id == "" {
			return nil, ErrResponse
		}
		if kind == "retest" && evidenceID(item["finding_id"]) != parentID {
			return nil, ErrResponse
		}
		result = append(result, EvidenceRecord{ID: id, ParentID: parentID, Kind: kind, Representation: representation, Content: raw})
	}
	return result, nil
}

func evidenceID(raw json.RawMessage) string {
	var value string
	if json.Unmarshal(raw, &value) != nil {
		value = string(raw)
	}
	if !validID(value) {
		return ""
	}
	return value
}

func EvidenceTaskID(raw json.RawMessage) string {
	var item map[string]json.RawMessage
	if json.Unmarshal(raw, &item) != nil {
		return ""
	}
	return evidenceID(item["task_id"])
}

func EvidenceLimit(raw string) (int, error) {
	if raw == "" {
		return 50, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 || n > 100 {
		return 0, ErrInvalidRequest
	}
	return n, nil
}
