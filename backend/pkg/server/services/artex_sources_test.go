package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
	"pentagi/pkg/integrations/artex"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
)

type modelSourceDB struct {
	database.Querier
	rows         []database.Provider
	requestedUID int64
}

func (d *modelSourceDB) GetUserProviders(ctx context.Context, uid int64) ([]database.Provider, error) {
	d.requestedUID = uid
	return d.rows, nil
}
func (d *modelSourceDB) GetUserProvider(ctx context.Context, request database.GetUserProviderParams) (database.Provider, error) {
	d.requestedUID = request.UserID
	for _, row := range d.rows {
		if row.ID == request.ID && row.UserID == request.UserID {
			return row, nil
		}
	}
	return database.Provider{}, sql.ErrNoRows
}

func TestArtexModelSourceOwnsAndDecryptsStoredConnection(t *testing.T) {
	// Fixed encrypted fixture generated from a synthetic key and fixture salt;
	// no application/provider credentials are read by this test.
	raw := []byte(`{"connection":{"base_url":"https://model.invalid/v1/","model_prefix":"vendor","api_key_encrypted":"v1:AwMDAwMDAwMDAwMDUKebmjzuYjqClue4mJKFcbBL74--s3I3bJhL-O3eWSVEsU-R"},"primary_agent":{"model":"model-a"}}`)
	db := &modelSourceDB{rows: []database.Provider{{ID: 7, UserID: 1, Name: "Owned", Type: database.ProviderTypeCustom, Config: raw}, {ID: 8, UserID: 2, Name: "Other user", Type: database.ProviderTypeCustom, Config: raw}}}
	source := &artexModelSource{db: db, cfg: &config.Config{CookieSigningSalt: "fixture-salt"}, defaults: provider.ProvidersConfig{}}
	choices, err := source.List(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(choices) != 1 || choices[0].Ref != "saved:7" || !choices[0].Compatible || choices[0].DefaultModel != "vendor/model-a" {
		t.Fatal("catalogue ownership/default model mismatch")
	}
	encoded, _ := json.Marshal(choices)
	if strings.Contains(string(encoded), "fixture-provider-key") || strings.Contains(string(encoded), "api_key_encrypted") {
		t.Fatal("catalogue leaked key material")
	}
	resolved, err := source.Resolve(context.Background(), 1, "saved:7", "model-b")
	if err != nil {
		t.Fatal(err)
	}
	if db.requestedUID != 1 || resolved.APIKey != "fixture-provider-key" || resolved.Model != "vendor/model-b" || resolved.BaseURL != "https://model.invalid/v1" {
		t.Fatal("stored connection was not resolved correctly")
	}
	if _, err := source.Resolve(context.Background(), 2, "saved:7", "model-b"); !errors.Is(err, artex.ErrSourceUnavailable) {
		t.Fatal("foreign provider resolved")
	}
	source.cfg.CookieSigningSalt = "incorrect-salt"
	if _, err := source.Resolve(context.Background(), 1, "saved:7", "model-b"); !errors.Is(err, artex.ErrSourceUnavailable) {
		t.Fatal("invalid ciphertext accepted")
	}
}

func TestArtexModelSourceExposesOnlyCompatibleProtocol(t *testing.T) {
	db := &modelSourceDB{rows: []database.Provider{{ID: 9, UserID: 1, Name: "Gemini", Type: database.ProviderTypeGemini, Config: []byte("{}")}}}
	source := &artexModelSource{db: db, cfg: &config.Config{OpenAIKey: "synthetic-environment-key", OpenAIServerURL: "https://model.invalid/v1"}, defaults: provider.ProvidersConfig{provider.ProviderOpenAI: &pconfig.ProviderConfig{PrimaryAgent: &pconfig.AgentConfig{Model: "env-model"}}}}
	choices, err := source.List(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(choices) != 2 || choices[0].Ref != "default:openai" || !choices[0].Compatible || choices[1].Compatible {
		t.Fatal("protocol compatibility catalogue is incorrect")
	}
	if _, err := source.Resolve(context.Background(), 1, "saved:9", "gemini"); !errors.Is(err, artex.ErrUnsupportedProvider) {
		t.Fatal("unsupported protocol was silently converted")
	}
	resolved, err := source.Resolve(context.Background(), 1, "default:openai", "")
	if err != nil || resolved.Model != "env-model" {
		t.Fatal("environment model not reusable")
	}
}
