package providers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"strings"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
)

const providerConnectionCipherVersion = "v1:"

func sealProviderConnection(cfg *config.Config, providerConfig *pconfig.ProviderConfig) error {
	if providerConfig == nil || providerConfig.Connection == nil {
		return nil
	}
	connection := providerConfig.Connection
	normalizedBaseURL, err := normalizeProviderBaseURL(connection.BaseURL)
	if err != nil {
		return err
	}
	connection.BaseURL = normalizedBaseURL
	connection.ModelPrefix = strings.Trim(strings.TrimSpace(connection.ModelPrefix), "/")
	if connection.APIKey == "" {
		return nil
	}
	encoded, err := encryptProviderSecret(cfg, connection.APIKey)
	if err != nil {
		return err
	}
	connection.EncryptedAPIKey = encoded
	connection.APIKey = ""
	return nil
}

func normalizeProviderBaseURL(raw string) (string, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(raw), "/")
	if baseURL == "" {
		return "", nil
	}

	u, err := url.ParseRequestURI(baseURL)
	if err != nil || u.Scheme == "" || u.Host == "" || (u.Scheme != "https" && u.Scheme != "http") {
		return "", fmt.Errorf("connection base URL must be an http or https URL")
	}

	return baseURL, nil
}

// ResolveStoredProviderConnection fills an omitted API key from a user-owned
// provider configuration. The caller must load the stored configuration after
// enforcing ownership. A stored key is only reused when the requested endpoint
// matches the saved endpoint, so a credential cannot be sent to a different URL.
func ResolveStoredProviderConnection(
	cfg *config.Config,
	stored *pconfig.ProviderConfig,
	requested *pconfig.ConnectionConfig,
) (*pconfig.ConnectionConfig, error) {
	if requested == nil {
		return nil, nil
	}

	resolved := *requested
	normalizedBaseURL, err := normalizeProviderBaseURL(resolved.BaseURL)
	if err != nil {
		return nil, err
	}
	resolved.BaseURL = normalizedBaseURL

	if strings.TrimSpace(resolved.APIKey) != "" || stored == nil || stored.Connection == nil {
		return &resolved, nil
	}

	storedBaseURL, err := normalizeProviderBaseURL(stored.Connection.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("stored provider connection is invalid: %w", err)
	}
	if storedBaseURL == "" || storedBaseURL != normalizedBaseURL {
		return nil, fmt.Errorf("provider base URL does not match the saved connection")
	}

	if stored.Connection.APIKey != "" {
		resolved.APIKey = stored.Connection.APIKey
		return &resolved, nil
	}
	if stored.Connection.EncryptedAPIKey == "" {
		return &resolved, nil
	}

	secret, err := decryptProviderSecret(cfg, stored.Connection.EncryptedAPIKey)
	if err != nil {
		return nil, fmt.Errorf("decrypt stored provider connection key: %w", err)
	}
	resolved.APIKey = secret

	return &resolved, nil
}

func hydrateProviderConnection(cfg *config.Config, providerConfig *pconfig.ProviderConfig) error {
	if providerConfig == nil || providerConfig.Connection == nil || providerConfig.Connection.EncryptedAPIKey == "" {
		return nil
	}
	secret, err := decryptProviderSecret(cfg, providerConfig.Connection.EncryptedAPIKey)
	if err != nil {
		return fmt.Errorf("decrypt provider connection key: %w", err)
	}
	providerConfig.Connection.APIKey = secret
	return nil
}

func mergeProviderConnection(existing, next *pconfig.ProviderConfig) {
	if next == nil {
		return
	}
	if next.Connection == nil {
		if existing != nil {
			next.Connection = existing.Connection
		}
		return
	}
	if next.Connection.APIKey == "" && next.Connection.EncryptedAPIKey == "" && existing != nil && existing.Connection != nil {
		next.Connection.EncryptedAPIKey = existing.Connection.EncryptedAPIKey
	}
}

func customConnectionConfigured(providerConfig *pconfig.ProviderConfig) bool {
	return providerConfig != nil && providerConfig.Connection != nil && strings.TrimSpace(providerConfig.Connection.BaseURL) != ""
}

func runtimeConfigForProvider(cfg *config.Config, providerType provider.ProviderType, providerConfig *pconfig.ProviderConfig) (*config.Config, error) {
	if providerType != provider.ProviderCustom || !customConnectionConfigured(providerConfig) {
		return cfg, nil
	}
	if err := hydrateProviderConnection(cfg, providerConfig); err != nil {
		return nil, err
	}
	runtimeCfg := *cfg
	runtimeCfg.LLMServerURL = providerConfig.Connection.BaseURL
	runtimeCfg.LLMServerKey = providerConfig.Connection.APIKey
	runtimeCfg.LLMServerProvider = providerConfig.Connection.ModelPrefix
	runtimeCfg.LLMServerModel = providerConnectionDefaultModel(providerConfig)
	return &runtimeCfg, nil
}

func providerConnectionDefaultModel(providerConfig *pconfig.ProviderConfig) string {
	if providerConfig == nil {
		return ""
	}
	for _, agent := range []*pconfig.AgentConfig{providerConfig.Simple, providerConfig.PrimaryAgent, providerConfig.Assistant} {
		if agent != nil && agent.Model != "" {
			return agent.Model
		}
	}
	return ""
}

func providerConnectionCipher(cfg *config.Config) (cipher.AEAD, error) {
	if cfg == nil || cfg.AuthSalt() == "" {
		return nil, fmt.Errorf("COOKIE_SIGNING_SALT is required to protect provider credentials")
	}
	key := sha256.Sum256([]byte("pentagi/provider-connection/v1|" + cfg.AuthSalt()))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func encryptProviderSecret(cfg *config.Config, secret string) (string, error) {
	aead, err := providerConnectionCipher(cfg)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := aead.Seal(nil, nonce, []byte(secret), nil)
	return providerConnectionCipherVersion + base64.RawURLEncoding.EncodeToString(append(nonce, ciphertext...)), nil
}

func decryptProviderSecret(cfg *config.Config, encoded string) (string, error) {
	if !strings.HasPrefix(encoded, providerConnectionCipherVersion) {
		return "", fmt.Errorf("unsupported provider credential format")
	}
	aead, err := providerConnectionCipher(cfg)
	if err != nil {
		return "", err
	}
	payload, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(encoded, providerConnectionCipherVersion))
	if err != nil {
		return "", err
	}
	if len(payload) < aead.NonceSize() {
		return "", fmt.Errorf("invalid encrypted provider credential")
	}
	plaintext, err := aead.Open(nil, payload[:aead.NonceSize()], payload[aead.NonceSize():], nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
