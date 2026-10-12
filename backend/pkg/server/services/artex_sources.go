package services

import (
	"context"
	"encoding/json"
	"sort"
	"strconv"
	"strings"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
	"pentagi/pkg/integrations/artex"
	"pentagi/pkg/providers"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
)

// Source rows are loaded by user ID before decrypting anything. The catalogue
// never decrypts or returns stored keys and never constructs an upstream client.
type artexModelSource struct {
	db       database.Querier
	cfg      *config.Config
	defaults provider.ProvidersConfig
}

func (s *artexModelSource) List(ctx context.Context, uid int64) ([]artex.ProviderChoice, error) {
	if uid <= 0 {
		return nil, artex.ErrSourceUnavailable
	}
	rows, err := s.db.GetUserProviders(ctx, uid)
	if err != nil {
		return nil, artex.ErrSourceUnavailable
	}
	result := []artex.ProviderChoice{}
	for _, kind := range []provider.ProviderType{provider.ProviderOpenAI, provider.ProviderCustom} {
		pcfg := s.defaults[kind]
		if pcfg == nil {
			continue
		}
		_, key, _, _ := s.connection(kind, pcfg, false)
		if key == "" {
			continue
		}
		result = append(result, s.choice("default:"+string(kind), string(kind), kind, pcfg))
	}
	for _, row := range rows {
		if row.UserID != uid {
			continue
		}
		var pcfg pconfig.ProviderConfig
		if json.Unmarshal(row.Config, &pcfg) != nil {
			result = append(result, artex.ProviderChoice{Ref: "saved:" + strconv.FormatInt(row.ID, 10), Name: row.Name, Type: string(row.Type), Models: []string{}, Reason: "invalid_configuration"})
			continue
		}
		result = append(result, s.choice("saved:"+strconv.FormatInt(row.ID, 10), row.Name, provider.ProviderType(row.Type), &pcfg))
	}
	return result, nil
}

func (s *artexModelSource) choice(ref, name string, kind provider.ProviderType, pcfg *pconfig.ProviderConfig) artex.ProviderChoice {
	result := artex.ProviderChoice{Ref: ref, Name: name, Type: string(kind), Models: []string{}}
	base, key, prefix, err := s.connection(kind, pcfg, false)
	if err != nil {
		result.Reason = "unsupported_protocol"
	} else if key == "" {
		result.Reason = "missing_credentials"
	} else {
		spec := artex.ProfileSpec{Format: "openai", BaseURL: base, Proxy: s.cfg.ProxyURL, Model: "capability-check", PoolExclude: true}
		if spec.Validate() != nil {
			result.Reason = "invalid_connection"
		} else {
			result.Compatible = true
		}
	}
	fallback := s.defaults[kind]
	seen := map[string]bool{}
	for _, candidate := range []*pconfig.ProviderConfig{pcfg, fallback} {
		if candidate == nil {
			continue
		}
		for _, agent := range []*pconfig.AgentConfig{candidate.PrimaryAgent, candidate.Simple, candidate.Assistant, candidate.Pentester, candidate.Coder} {
			if agent == nil || agent.Model == "" {
				continue
			}
			model := provider.ApplyModelPrefix(agent.Model, prefix)
			if result.DefaultModel == "" {
				result.DefaultModel = model
			}
			if !seen[model] {
				result.Models = append(result.Models, model)
				seen[model] = true
			}
		}
	}
	if result.DefaultModel == "" && kind == provider.ProviderCustom {
		result.DefaultModel = provider.ApplyModelPrefix(s.cfg.LLMServerModel, prefix)
		if result.DefaultModel != "" {
			result.Models = append(result.Models, result.DefaultModel)
		}
	}
	sort.Strings(result.Models)
	return result
}

func (s *artexModelSource) Resolve(ctx context.Context, uid int64, ref, model string) (artex.ResolvedProvider, error) {
	if uid <= 0 {
		return artex.ResolvedProvider{}, artex.ErrSourceUnavailable
	}
	var kind provider.ProviderType
	var pcfg *pconfig.ProviderConfig
	switch {
	case strings.HasPrefix(ref, "saved:"):
		raw := strings.TrimPrefix(ref, "saved:")
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || id <= 0 || strconv.FormatInt(id, 10) != raw {
			return artex.ResolvedProvider{}, artex.ErrSourceUnavailable
		}
		row, err := s.db.GetUserProvider(ctx, database.GetUserProviderParams{ID: id, UserID: uid})
		if err != nil || row.UserID != uid {
			return artex.ResolvedProvider{}, artex.ErrSourceUnavailable
		}
		kind = provider.ProviderType(row.Type)
		pcfg = new(pconfig.ProviderConfig)
		if json.Unmarshal(row.Config, pcfg) != nil {
			return artex.ResolvedProvider{}, artex.ErrSourceUnavailable
		}
	case ref == "default:openai" || ref == "default:custom":
		kind = provider.ProviderType(strings.TrimPrefix(ref, "default:"))
		pcfg = s.defaults[kind]
		if pcfg == nil {
			return artex.ResolvedProvider{}, artex.ErrSourceUnavailable
		}
	default:
		return artex.ResolvedProvider{}, artex.ErrSourceUnavailable
	}
	base, key, prefix, err := s.connection(kind, pcfg, true)
	if err != nil {
		return artex.ResolvedProvider{}, err
	}
	if key == "" {
		return artex.ResolvedProvider{}, artex.ErrSourceUnavailable
	}
	model = strings.TrimSpace(model)
	if model == "" {
		model = s.choice(ref, "", kind, pcfg).DefaultModel
	}
	return artex.ResolvedProvider{BaseURL: base, APIKey: key, Proxy: s.cfg.ProxyURL, Model: provider.ApplyModelPrefix(model, prefix)}, nil
}

func (s *artexModelSource) connection(kind provider.ProviderType, pcfg *pconfig.ProviderConfig, decrypt bool) (string, string, string, error) {
	switch kind {
	case provider.ProviderOpenAI:
		base := strings.TrimRight(strings.TrimSpace(s.cfg.OpenAIServerURL), "/")
		if base == "" {
			base = "https://api.openai.com/v1"
		}
		return base, s.cfg.OpenAIKey, "", nil
	case provider.ProviderCustom:
		if pcfg != nil && pcfg.Connection != nil && strings.TrimSpace(pcfg.Connection.BaseURL) != "" {
			connection := pcfg.Connection
			key := ""
			if decrypt {
				resolved, err := providers.ResolveStoredProviderConnection(s.cfg, pcfg, &pconfig.ConnectionConfig{BaseURL: connection.BaseURL})
				if err != nil {
					return "", "", "", artex.ErrSourceUnavailable
				}
				key = resolved.APIKey
			} else if connection.EncryptedAPIKey != "" || connection.APIKey != "" {
				key = "configured"
			}
			return strings.TrimRight(strings.TrimSpace(connection.BaseURL), "/"), key, connection.ModelPrefix, nil
		}
		return strings.TrimRight(strings.TrimSpace(s.cfg.LLMServerURL), "/"), s.cfg.LLMServerKey, s.cfg.LLMServerProvider, nil
	default:
		return "", "", "", artex.ErrUnsupportedProvider
	}
}
