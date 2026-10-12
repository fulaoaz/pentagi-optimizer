package config

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

const runtimeConfigFileName = "runtime-config.json"

var runtimeConfigMu sync.RWMutex

// LoadRuntimeConfigOverrides restores values saved from the administrator UI.
// Environment variables remain the first source of configuration; the override
// file is intended for installations where editing the container environment is
// not practical. Values are stored below DATA_DIR with restrictive permissions.
func LoadRuntimeConfigOverrides(cfg *Config) error {
	if cfg == nil {
		return nil
	}

	runtimeConfigMu.Lock()
	defer runtimeConfigMu.Unlock()

	path := runtimeConfigPath(cfg.DataDir)
	contents, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read runtime configuration overrides: %w", err)
	}

	var overrides map[string]string
	if err := json.Unmarshal(contents, &overrides); err != nil {
		return fmt.Errorf("parse runtime configuration overrides: %w", err)
	}

	for key, raw := range overrides {
		if err := setRuntimeConfigValue(cfg, key, raw); err != nil {
			return fmt.Errorf("apply runtime configuration %s: %w", key, err)
		}
		if raw == "" {
			_ = os.Unsetenv(key)
		} else {
			_ = os.Setenv(key, raw)
		}
	}

	return nil
}

// UpdateRuntimeConfig applies and persists administrator-edited values. It
// validates every update before changing the live config so a malformed batch
// cannot leave the process half-updated. Most settings are marked restart
// required because long-lived clients (database, Docker, providers) capture
// their values during startup.
func UpdateRuntimeConfig(cfg *Config, updates map[string]string) error {
	if cfg == nil {
		return fmt.Errorf("configuration is unavailable")
	}
	if len(updates) == 0 {
		return nil
	}

	runtimeConfigMu.Lock()
	defer runtimeConfigMu.Unlock()

	for key, raw := range updates {
		if err := validateRuntimeConfigValue(cfg, key, raw); err != nil {
			return err
		}
	}

	overrides, err := readRuntimeConfigOverrides(cfg.DataDir)
	if err != nil {
		return err
	}
	for key, raw := range updates {
		if err := setRuntimeConfigValue(cfg, key, raw); err != nil {
			return err
		}
		if raw == "" {
			delete(overrides, key)
			_ = os.Unsetenv(key)
		} else {
			overrides[key] = raw
			_ = os.Setenv(key, raw)
		}
	}

	return writeRuntimeConfigOverrides(cfg.DataDir, overrides)
}

func runtimeConfigPath(dataDir string) string {
	if strings.TrimSpace(dataDir) == "" {
		dataDir = "./data"
	}
	return filepath.Join(dataDir, runtimeConfigFileName)
}

func readRuntimeConfigOverrides(dataDir string) (map[string]string, error) {
	contents, err := os.ReadFile(runtimeConfigPath(dataDir))
	if os.IsNotExist(err) {
		return map[string]string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read runtime configuration overrides: %w", err)
	}
	var overrides map[string]string
	if err := json.Unmarshal(contents, &overrides); err != nil {
		return nil, fmt.Errorf("parse runtime configuration overrides: %w", err)
	}
	if overrides == nil {
		overrides = map[string]string{}
	}
	return overrides, nil
}

func writeRuntimeConfigOverrides(dataDir string, overrides map[string]string) error {
	if strings.TrimSpace(dataDir) == "" {
		dataDir = "./data"
	}
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return fmt.Errorf("create runtime configuration directory: %w", err)
	}
	contents, err := json.MarshalIndent(overrides, "", "  ")
	if err != nil {
		return fmt.Errorf("encode runtime configuration overrides: %w", err)
	}
	contents = append(contents, '\n')
	target := runtimeConfigPath(dataDir)
	temporary := target + ".tmp"
	if err := os.WriteFile(temporary, contents, 0o600); err != nil {
		return fmt.Errorf("write runtime configuration overrides: %w", err)
	}
	// Windows does not replace an existing destination with Rename. Remove the
	// old file only after the temporary file is complete so a failed write never
	// leaves a half-written JSON document.
	if err := os.Remove(target); err != nil && !os.IsNotExist(err) {
		_ = os.Remove(temporary)
		return fmt.Errorf("prepare runtime configuration overrides: %w", err)
	}
	if err := os.Rename(temporary, target); err != nil {
		_ = os.Remove(temporary)
		return fmt.Errorf("activate runtime configuration overrides: %w", err)
	}
	return nil
}

func validateRuntimeConfigValue(cfg *Config, key, raw string) error {
	_, err := runtimeConfigField(cfg, key)
	if err != nil {
		return err
	}
	_, err = parseRuntimeConfigValue(key, raw)
	return err
}

func setRuntimeConfigValue(cfg *Config, key, raw string) error {
	field, err := runtimeConfigField(cfg, key)
	if err != nil {
		return err
	}
	value, err := parseRuntimeConfigValue(key, raw)
	if err != nil {
		return err
	}
	field.Set(value)
	return nil
}

func runtimeConfigField(cfg *Config, key string) (reflect.Value, error) {
	if cfg == nil {
		return reflect.Value{}, fmt.Errorf("configuration is unavailable")
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return reflect.Value{}, fmt.Errorf("configuration key is required")
	}
	value := reflect.ValueOf(cfg).Elem()
	typ := value.Type()
	for i := 0; i < typ.NumField(); i++ {
		fieldInfo := typ.Field(i)
		envKey := strings.Split(fieldInfo.Tag.Get("env"), ",")[0]
		if envKey == key {
			field := value.Field(i)
			if !field.CanSet() {
				return reflect.Value{}, fmt.Errorf("configuration key %s is not editable", key)
			}
			return field, nil
		}
	}
	return reflect.Value{}, fmt.Errorf("unknown configuration key %s", key)
}

func parseRuntimeConfigValue(key, raw string) (reflect.Value, error) {
	// The type is resolved from the Config field so the parser cannot accept a
	// value that would later be silently truncated or coerced.
	var fieldType reflect.Type
	var probe Config
	field, err := runtimeConfigField(&probe, key)
	if err != nil {
		return reflect.Value{}, err
	}
	fieldType = field.Type()
	parsed, err := parseRuntimeConfigValueForType(key, raw, fieldType)
	if err != nil {
		return reflect.Value{}, err
	}
	return parsed, nil
}

func parseRuntimeConfigValueForType(key, raw string, typ reflect.Type) (reflect.Value, error) {
	trimmed := strings.TrimSpace(raw)
	if typ == reflect.TypeOf((*url.URL)(nil)) {
		if trimmed == "" {
			return reflect.Zero(typ), nil
		}
		parsed, err := url.Parse(trimmed)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return reflect.Value{}, fmt.Errorf("%s must be a valid URL", key)
		}
		result := reflect.New(typ.Elem())
		result.Elem().Set(reflect.ValueOf(*parsed))
		return result, nil
	}

	switch typ.Kind() {
	case reflect.String:
		return reflect.ValueOf(raw).Convert(typ), nil
	case reflect.Bool:
		value, err := strconv.ParseBool(trimmed)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("%s must be true or false", key)
		}
		return reflect.ValueOf(value).Convert(typ), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value, err := strconv.ParseInt(trimmed, 10, typ.Bits())
		if err != nil {
			return reflect.Value{}, fmt.Errorf("%s must be an integer", key)
		}
		result := reflect.New(typ).Elem()
		result.SetInt(value)
		return result, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value, err := strconv.ParseUint(trimmed, 10, typ.Bits())
		if err != nil {
			return reflect.Value{}, fmt.Errorf("%s must be an unsigned integer", key)
		}
		result := reflect.New(typ).Elem()
		result.SetUint(value)
		return result, nil
	case reflect.Slice:
		if typ.Elem().Kind() != reflect.String {
			return reflect.Value{}, fmt.Errorf("%s has an unsupported list type", key)
		}
		if trimmed == "" {
			return reflect.MakeSlice(typ, 0, 0), nil
		}
		parts := strings.Split(raw, ",")
		result := reflect.MakeSlice(typ, 0, len(parts))
		for _, part := range parts {
			result = reflect.Append(result, reflect.ValueOf(strings.TrimSpace(part)).Convert(typ.Elem()))
		}
		return result, nil
	default:
		return reflect.Value{}, fmt.Errorf("%s has an unsupported configuration type", key)
	}
}
