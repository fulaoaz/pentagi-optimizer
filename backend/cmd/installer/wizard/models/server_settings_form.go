package models

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"pentagi/cmd/installer/loader"
	"pentagi/cmd/installer/wizard/controller"
	"pentagi/cmd/installer/wizard/locale"
	"pentagi/cmd/installer/wizard/logger"
	"pentagi/cmd/installer/wizard/styles"
	"pentagi/cmd/installer/wizard/window"
	"pentagi/pkg/config"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/vxcontrol/cloud/sdk"
)

// schemaNameRegex keeps DATABASE_EXTENSIONS_SCHEMA within PostgreSQL's
// unquoted-identifier rules and 63-byte limit, so a typo surfaces here rather
// than as a failed CREATE EXTENSION on first boot.
var schemaNameRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]{0,62}$`)

const mcpMaxRequestBytesLimit = 64 * 1024 * 1024
const mcpToolRateLimitLimit = 100000

var mcpToolNameRegex = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_.-]{0,127}$`)

// ServerSettingsFormModel represents the PentAGI server settings configuration form
type ServerSettingsFormModel struct {
	*BaseScreen
}

// NewServerSettingsFormModel creates a new server settings form model
func NewServerSettingsFormModel(c controller.Controller, s styles.Styles, w window.Window) *ServerSettingsFormModel {
	m := &ServerSettingsFormModel{}

	// create base screen with this model as handler (no list handler needed)
	m.BaseScreen = NewBaseScreen(c, s, w, m, nil)

	return m
}

// BuildForm constructs the fields for server settings
func (m *ServerSettingsFormModel) BuildForm() tea.Cmd {
	config := m.GetController().GetServerSettingsConfig()
	fields := []FormField{}

	fields = append(fields, m.createTextField("pentagi_license_key",
		locale.ServerSettingsLicenseKey,
		locale.ServerSettingsLicenseKeyDesc,
		config.LicenseKey,
		true,
	))

	fields = append(fields, m.createTextField("pentagi_tenant_id",
		locale.ServerSettingsTenantID,
		locale.ServerSettingsTenantIDDesc,
		config.TenantID,
		false,
	))

	fields = append(fields, m.createTextField("pentagi_pprof_addr",
		locale.ServerSettingsPprofAddr,
		locale.ServerSettingsPprofAddrDesc,
		config.PprofAddr,
		false,
	))

	// host and port
	fields = append(fields, m.createTextField("pentagi_server_host",
		locale.ServerSettingsHost,
		locale.ServerSettingsHostDesc,
		config.ListenIP,
		false,
	))

	fields = append(fields, m.createTextField("pentagi_server_port",
		locale.ServerSettingsPort,
		locale.ServerSettingsPortDesc,
		config.ListenPort,
		false,
	))

	// public url
	fields = append(fields, m.createTextField("pentagi_public_url",
		locale.ServerSettingsPublicURL,
		locale.ServerSettingsPublicURLDesc,
		config.PublicURL,
		false,
	))

	// cors origins
	fields = append(fields, m.createTextField("pentagi_cors_origins",
		locale.ServerSettingsCORSOrigins,
		locale.ServerSettingsCORSOriginsDesc,
		config.CorsOrigins,
		false,
	))

	fields = append(fields, m.createBooleanField("mcp_enabled",
		locale.ServerSettingsMCPEnabled,
		locale.ServerSettingsMCPEnabledDesc,
		config.MCPEnabled,
	))

	fields = append(fields, m.createTextField("mcp_server_name",
		locale.ServerSettingsMCPServerName,
		locale.ServerSettingsMCPServerNameDesc,
		config.MCPServerName,
		false,
	))

	fields = append(fields, m.createTextField("mcp_server_version",
		locale.ServerSettingsMCPServerVersion,
		locale.ServerSettingsMCPServerVersionDesc,
		config.MCPServerVersion,
		false,
	))

	fields = append(fields, m.createTextField("mcp_api_key",
		locale.ServerSettingsMCPAPIKey,
		locale.ServerSettingsMCPAPIKeyDesc,
		config.MCPAPIKey,
		true,
	))

	fields = append(fields, m.createTextField("mcp_write_api_key",
		locale.ServerSettingsMCPWriteAPIKey,
		locale.ServerSettingsMCPWriteAPIKeyDesc,
		config.MCPWriteAPIKey,
		true,
	))

	fields = append(fields, m.createBooleanField("mcp_allow_anonymous",
		locale.ServerSettingsMCPAllowAnonymous,
		locale.ServerSettingsMCPAllowAnonymousDesc,
		config.MCPAllowAnonymous,
	))

	fields = append(fields, m.createTextField("mcp_allowed_origins",
		locale.ServerSettingsMCPAllowedOrigins,
		locale.ServerSettingsMCPAllowedOriginsDesc,
		config.MCPAllowedOrigins,
		false,
	))

	fields = append(fields, m.createTextField("mcp_allowed_tools",
		locale.ServerSettingsMCPAllowedTools,
		locale.ServerSettingsMCPAllowedToolsDesc,
		config.MCPAllowedTools,
		false,
	))

	fields = append(fields, m.createTextField("mcp_max_request_bytes",
		locale.ServerSettingsMCPMaxRequestBytes,
		locale.ServerSettingsMCPMaxRequestBytesDesc,
		config.MCPMaxRequestBytes,
		false,
	))

	fields = append(fields, m.createTextField("mcp_read_tool_rate_limit",
		locale.ServerSettingsMCPReadToolRateLimit,
		locale.ServerSettingsMCPReadToolRateLimitDesc,
		config.MCPReadToolRateLimit,
		false,
	))

	fields = append(fields, m.createTextField("mcp_write_tool_rate_limit",
		locale.ServerSettingsMCPWriteToolRateLimit,
		locale.ServerSettingsMCPWriteToolRateLimitDesc,
		config.MCPWriteToolRateLimit,
		false,
	))

	fields = append(fields, m.createTextField("mcp_approval_mode",
		locale.ServerSettingsMCPApprovalMode,
		locale.ServerSettingsMCPApprovalModeDesc,
		config.MCPApprovalMode,
		false,
	))

	fields = append(fields, m.createBooleanField("mcp_enable_write_tools",
		locale.ServerSettingsMCPEnableWriteTools,
		locale.ServerSettingsMCPEnableWriteToolsDesc,
		config.MCPEnableWriteTools,
	))

	// proxy: url, username, password
	fields = append(fields, m.createTextField("proxy_url",
		locale.ServerSettingsProxyURL,
		locale.ServerSettingsProxyURLDesc,
		config.ProxyURL,
		false,
	))
	fields = append(fields, m.createRawField("proxy_username",
		locale.ServerSettingsProxyUsername,
		locale.ServerSettingsProxyUsernameDesc,
		config.ProxyUsername,
		true,
	))
	fields = append(fields, m.createRawField("proxy_password",
		locale.ServerSettingsProxyPassword,
		locale.ServerSettingsProxyPasswordDesc,
		config.ProxyPassword,
		true,
	))

	// http client timeout
	fields = append(fields, m.createTextField("http_client_timeout",
		locale.ServerSettingsHTTPClientTimeout,
		locale.ServerSettingsHTTPClientTimeoutDesc,
		config.HTTPClientTimeout,
		false,
	))
	fields = append(fields, m.createTextField("terminal_tool_timeout",
		locale.ServerSettingsTerminalToolTimeout,
		locale.ServerSettingsTerminalToolTimeoutDesc,
		config.TerminalToolTimeout,
		false,
	))

	// external ssl settings
	fields = append(fields, m.createTextField("external_ssl_ca_path",
		locale.ServerSettingsExternalSSLCAPath,
		locale.ServerSettingsExternalSSLCAPathDesc,
		config.ExternalSSLCAPath,
		false,
	))
	fields = append(fields, m.createTextField("external_ssl_insecure",
		locale.ServerSettingsExternalSSLInsecure,
		locale.ServerSettingsExternalSSLInsecureDesc,
		config.ExternalSSLInsecure,
		false,
	))

	// ssl dir
	fields = append(fields, m.createTextField("pentagi_ssl_dir",
		locale.ServerSettingsSSLDir,
		locale.ServerSettingsSSLDirDesc,
		config.SSLDir,
		false,
	))

	// data dir
	fields = append(fields, m.createTextField("pentagi_data_dir",
		locale.ServerSettingsDataDir,
		locale.ServerSettingsDataDirDesc,
		config.DataDir,
		false,
	))

	// cookie signing salt (masked)
	fields = append(fields, m.createTextField("pentagi_cookie_signing_salt",
		locale.ServerSettingsCookieSigningSalt,
		locale.ServerSettingsCookieSigningSaltDesc,
		config.CookieSigningSalt,
		true,
	))

	fields = append(fields, m.createTextField("database_extensions_schema",
		locale.ServerSettingsDatabaseExtensionsSchema,
		locale.ServerSettingsDatabaseExtensionsSchemaDesc,
		config.DatabaseExtensionsSchema,
		false,
	))

	fields = append(fields, m.createTextField("database_search_path_via_options",
		locale.ServerSettingsDatabaseSearchPathViaOptions,
		locale.ServerSettingsDatabaseSearchPathViaOptionsDesc,
		config.DatabaseSearchPathViaOpt,
		false,
	))

	m.SetFormFields(fields)
	return nil
}

func (m *ServerSettingsFormModel) createTextField(key, title, description string, envVar loader.EnvVar, masked bool) FormField {
	// reuse generic text input builder
	input := NewTextInput(m.GetStyles(), m.GetWindow(), envVar)

	return FormField{
		Key:         key,
		Title:       title,
		Description: description,
		Required:    false,
		Masked:      masked,
		Input:       input,
		Value:       input.Value(),
	}
}

func (m *ServerSettingsFormModel) createBooleanField(key, title, description string, envVar loader.EnvVar) FormField {
	input := NewBooleanInput(m.GetStyles(), m.GetWindow(), envVar)
	return FormField{
		Key:         key,
		Title:       title,
		Description: description,
		Required:    false,
		Masked:      false,
		Input:       input,
		Value:       input.Value(),
		Suggestions: input.AvailableSuggestions(),
	}
}

// createRawField is used for non-env raw values (like usernames/passwords parsed from URLs)
func (m *ServerSettingsFormModel) createRawField(key, title, description, value string, masked bool) FormField {
	input := NewTextInput(m.GetStyles(), m.GetWindow(), loader.EnvVar{Value: value})
	return FormField{
		Key:         key,
		Title:       title,
		Description: description,
		Required:    false,
		Masked:      masked,
		Input:       input,
		Value:       input.Value(),
	}
}

func (m *ServerSettingsFormModel) GetFormTitle() string {
	return locale.ServerSettingsFormTitle
}

func (m *ServerSettingsFormModel) GetFormDescription() string {
	return locale.ServerSettingsFormDescription
}

func (m *ServerSettingsFormModel) GetFormName() string {
	return locale.ServerSettingsFormName
}

func (m *ServerSettingsFormModel) GetFormSummary() string {
	return ""
}

func (m *ServerSettingsFormModel) GetFormOverview() string {
	var sections []string

	sections = append(sections, m.GetStyles().Subtitle.Render(locale.ServerSettingsFormTitle))
	sections = append(sections, "")
	sections = append(sections, m.GetStyles().Paragraph.Bold(true).Render(locale.ServerSettingsFormDescription))
	sections = append(sections, "")
	sections = append(sections, m.GetStyles().Paragraph.Render(locale.ServerSettingsFormOverview))

	return strings.Join(sections, "\n")
}

func (m *ServerSettingsFormModel) GetCurrentConfiguration() string {
	var sections []string
	cfg := m.GetController().GetServerSettingsConfig()

	sections = append(sections, m.GetStyles().Subtitle.Render(m.GetFormName()))

	getMaskedValue := func(value string) string {
		maskedValue := strings.Repeat("*", len(value))
		if len(value) > 15 {
			maskedValue = maskedValue[:15] + "..."
		}
		return maskedValue
	}

	licenseStatus := locale.StatusNotConfigured
	if licenseKey := cfg.LicenseKey.Value; licenseKey != "" {
		licenseStatus = locale.StatusConfigured
	}
	licenseStatus = m.GetStyles().Muted.Render(licenseStatus)
	sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsLicenseKeyHint, licenseStatus))

	if tenantID := cfg.TenantID.Value; tenantID != "" {
		tenantID = m.GetStyles().Info.Render(tenantID)
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsTenantIDHint, tenantID))
	} else {
		tenantID = locale.StatusNotConfigured
		tenantID = m.GetStyles().Muted.Render(tenantID)
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsTenantIDHint, tenantID))
	}

	// both only take effect with a tenant configured, so keep them out of the
	// overview of a single-instance deployment
	if cfg.TenantID.Value != "" {
		if schema := cfg.DatabaseExtensionsSchema.Value; schema != "" {
			schema = m.GetStyles().Info.Render(schema)
			sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsDatabaseExtensionsSchemaHint, schema))
		} else if schema := cfg.DatabaseExtensionsSchema.Default; schema != "" {
			schema = m.GetStyles().Muted.Render(schema)
			sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsDatabaseExtensionsSchemaHint, schema))
		}

		if viaOptions := cfg.DatabaseSearchPathViaOpt.Value; viaOptions == "true" {
			viaOptions = m.GetStyles().Info.Render(locale.StatusEnabled)
			sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsDatabaseSearchPathViaOptionsHint, viaOptions))
		} else {
			viaOptions = m.GetStyles().Muted.Render(locale.StatusDisabled)
			sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsDatabaseSearchPathViaOptionsHint, viaOptions))
		}
	}

	if pprofAddr := cfg.PprofAddr.Value; pprofAddr != "" {
		pprofAddr = m.GetStyles().Info.Render(pprofAddr)
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsPprofAddrHint, pprofAddr))
	} else {
		pprofAddr = locale.StatusNotConfigured
		pprofAddr = m.GetStyles().Muted.Render(pprofAddr)
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsPprofAddrHint, pprofAddr))
	}

	if listenIP := cfg.ListenIP.Value; listenIP != "" {
		listenIP = m.GetStyles().Info.Render(listenIP)
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsHostHint, listenIP))
	} else if listenIP := cfg.ListenIP.Default; listenIP != "" {
		listenIP = m.GetStyles().Muted.Render(listenIP)
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsHostHint, listenIP))
	}

	if listenPort := cfg.ListenPort.Value; listenPort != "" {
		listenPort = m.GetStyles().Info.Render(listenPort)
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsPortHint, listenPort))
	} else if listenPort := cfg.ListenPort.Default; listenPort != "" {
		listenPort = m.GetStyles().Muted.Render(listenPort)
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsPortHint, listenPort))
	}

	if publicURL := cfg.PublicURL.Value; publicURL != "" {
		publicURL = m.GetStyles().Info.Render(publicURL)
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsPublicURLHint, publicURL))
	} else if publicURL := cfg.PublicURL.Default; publicURL != "" {
		publicURL = m.GetStyles().Muted.Render(publicURL)
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsPublicURLHint, publicURL))
	}

	if cors := cfg.CorsOrigins.Value; cors != "" {
		cors = m.GetStyles().Info.Render(cors)
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsCORSOriginsHint, cors))
	} else if cors := cfg.CorsOrigins.Default; cors != "" {
		cors = m.GetStyles().Muted.Render(cors)
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsCORSOriginsHint, cors))
	}

	mcpEnabled := cfg.MCPEnabled.Value
	if mcpEnabled == "" {
		mcpEnabled = cfg.MCPEnabled.Default
	}
	mcpEnabledStatus := locale.StatusDisabled
	mcpEnabledStyle := m.GetStyles().Muted
	if mcpEnabled == "true" {
		mcpEnabledStatus = locale.StatusEnabled
		mcpEnabledStyle = m.GetStyles().Success
	}
	sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsMCPEnabledHint, mcpEnabledStyle.Render(mcpEnabledStatus)))

	mcpServerName := cfg.MCPServerName.Value
	if mcpServerName == "" {
		mcpServerName = cfg.MCPServerName.Default
	}
	if mcpServerName != "" {
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsMCPServerNameHint, m.GetStyles().Info.Render(mcpServerName)))
	}

	mcpServerVersion := cfg.MCPServerVersion.Value
	if mcpServerVersion == "" {
		mcpServerVersion = cfg.MCPServerVersion.Default
	}
	if mcpServerVersion != "" {
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsMCPServerVersionHint, m.GetStyles().Info.Render(mcpServerVersion)))
	}

	appendMCPConfigured := func(title, value string) {
		status := locale.StatusNotConfigured
		style := m.GetStyles().Muted
		if strings.TrimSpace(value) != "" {
			status = locale.StatusConfigured
			style = m.GetStyles().Info
		}
		sections = append(sections, fmt.Sprintf("• %s: %s", title, style.Render(status)))
	}
	appendMCPConfigured(locale.ServerSettingsMCPAPIKeyHint, cfg.MCPAPIKey.Value)
	appendMCPConfigured(locale.ServerSettingsMCPWriteAPIKeyHint, cfg.MCPWriteAPIKey.Value)

	mcpAnonymous := cfg.MCPAllowAnonymous.Value
	if mcpAnonymous == "" {
		mcpAnonymous = cfg.MCPAllowAnonymous.Default
	}
	mcpAnonymousStatus := locale.StatusDisabled
	mcpAnonymousStyle := m.GetStyles().Muted
	if mcpAnonymous == "true" {
		mcpAnonymousStatus = locale.StatusEnabled
		mcpAnonymousStyle = m.GetStyles().Warning
	}
	sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsMCPAllowAnonymousHint, mcpAnonymousStyle.Render(mcpAnonymousStatus)))
	appendMCPConfigured(locale.ServerSettingsMCPAllowedOriginsHint, cfg.MCPAllowedOrigins.Value)
	appendMCPConfigured(locale.ServerSettingsMCPAllowedToolsHint, cfg.MCPAllowedTools.Value)

	mcpMaxRequestBytes := cfg.MCPMaxRequestBytes.Value
	if mcpMaxRequestBytes == "" {
		mcpMaxRequestBytes = cfg.MCPMaxRequestBytes.Default
	}
	if mcpMaxRequestBytes != "" {
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsMCPMaxRequestBytesHint, m.GetStyles().Info.Render(mcpMaxRequestBytes+" B")))
	}

	mcpReadRate := cfg.MCPReadToolRateLimit.Value
	if mcpReadRate == "" {
		mcpReadRate = cfg.MCPReadToolRateLimit.Default
	}
	if mcpReadRate != "" {
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsMCPReadToolRateLimitHint, m.GetStyles().Info.Render(mcpReadRate+locale.ServerSettingsMCPRateLimitUnit)))
	}

	mcpWriteRate := cfg.MCPWriteToolRateLimit.Value
	if mcpWriteRate == "" {
		mcpWriteRate = cfg.MCPWriteToolRateLimit.Default
	}
	if mcpWriteRate != "" {
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsMCPWriteToolRateLimitHint, m.GetStyles().Warning.Render(mcpWriteRate+locale.ServerSettingsMCPRateLimitUnit)))
	}

	mcpApprovalMode := cfg.MCPApprovalMode.Value
	if mcpApprovalMode == "" {
		mcpApprovalMode = cfg.MCPApprovalMode.Default
	}
	if mcpApprovalMode != "" {
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsMCPApprovalModeHint, m.GetStyles().Info.Render(mcpApprovalMode)))
	}

	mcpWriteTools := cfg.MCPEnableWriteTools.Value
	if mcpWriteTools == "" {
		mcpWriteTools = cfg.MCPEnableWriteTools.Default
	}
	mcpWriteToolsStatus := locale.StatusDisabled
	mcpWriteToolsStyle := m.GetStyles().Muted
	if mcpWriteTools == "true" {
		mcpWriteToolsStatus = locale.StatusEnabled
		mcpWriteToolsStyle = m.GetStyles().Warning
	}
	sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsMCPEnableWriteToolsHint, mcpWriteToolsStyle.Render(mcpWriteToolsStatus)))

	if proxyURL := cfg.ProxyURL.Value; proxyURL != "" {
		proxyURL = m.GetStyles().Info.Render(proxyURL)
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsProxyURLHint, proxyURL))
	} else {
		proxyURL = locale.StatusNotConfigured
		proxyURL = m.GetStyles().Muted.Render(proxyURL)
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsProxyURLHint, proxyURL))
	}

	if proxyUsername := getMaskedValue(cfg.ProxyUsername); proxyUsername != "" {
		proxyUsername = m.GetStyles().Muted.Render(proxyUsername)
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsProxyUsernameHint, proxyUsername))
	}

	if proxyPassword := getMaskedValue(cfg.ProxyPassword); proxyPassword != "" {
		proxyPassword = m.GetStyles().Muted.Render(proxyPassword)
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsProxyPasswordHint, proxyPassword))
	}

	if httpTimeout := cfg.HTTPClientTimeout.Value; httpTimeout != "" {
		httpTimeout = m.GetStyles().Info.Render(httpTimeout + "s")
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsHTTPClientTimeoutHint, httpTimeout))
	} else if httpTimeout := cfg.HTTPClientTimeout.Default; httpTimeout != "" {
		httpTimeout = m.GetStyles().Muted.Render(httpTimeout + "s")
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsHTTPClientTimeoutHint, httpTimeout))
	}

	if terminalTimeout := cfg.TerminalToolTimeout.Value; terminalTimeout != "" {
		terminalTimeout = m.GetStyles().Info.Render(terminalTimeout + "s")
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsTerminalToolTimeoutHint, terminalTimeout))
	} else if terminalTimeout := cfg.TerminalToolTimeout.Default; terminalTimeout != "" {
		terminalTimeout = m.GetStyles().Muted.Render(terminalTimeout + "s")
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsTerminalToolTimeoutHint, terminalTimeout))
	}

	if externalSSLCAPath := cfg.ExternalSSLCAPath.Value; externalSSLCAPath != "" {
		externalSSLCAPath = m.GetStyles().Info.Render(externalSSLCAPath)
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsExternalSSLCAPathHint, externalSSLCAPath))
	} else {
		externalSSLCAPath = locale.StatusNotConfigured
		externalSSLCAPath = m.GetStyles().Muted.Render(externalSSLCAPath)
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsExternalSSLCAPathHint, externalSSLCAPath))
	}

	if externalSSLInsecure := cfg.ExternalSSLInsecure.Value; externalSSLInsecure == "true" {
		externalSSLInsecure = m.GetStyles().Warning.Render(locale.StatusEnabledInsecure)
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsExternalSSLInsecureHint, externalSSLInsecure))
	} else if externalSSLInsecure := cfg.ExternalSSLInsecure.Default; externalSSLInsecure == "false" || externalSSLInsecure == "" {
		externalSSLInsecure = m.GetStyles().Muted.Render(locale.StatusDisabled)
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsExternalSSLInsecureHint, externalSSLInsecure))
	}

	if sslDir := cfg.SSLDir.Value; sslDir != "" {
		sslDir = m.GetStyles().Info.Render(sslDir)
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsSSLDirHint, sslDir))
	} else if sslDir := cfg.SSLDir.Default; sslDir != "" {
		sslDir = m.GetStyles().Muted.Render(sslDir)
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsSSLDirHint, sslDir))
	}

	if dataDir := cfg.DataDir.Value; dataDir != "" {
		dataDir = m.GetStyles().Info.Render(dataDir)
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsDataDirHint, dataDir))
	} else if dataDir := cfg.DataDir.Default; dataDir != "" {
		dataDir = m.GetStyles().Muted.Render(dataDir)
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ServerSettingsDataDirHint, dataDir))
	}

	return strings.Join(sections, "\n")
}

func (m *ServerSettingsFormModel) IsConfigured() bool {
	cfg := m.GetController().GetServerSettingsConfig()
	return cfg.ListenIP.Value != "" && cfg.ListenPort.Value != ""
}

func (m *ServerSettingsFormModel) GetHelpContent() string {
	var sections []string

	sections = append(sections, m.GetStyles().Subtitle.Render(locale.ServerSettingsFormTitle))
	sections = append(sections, "")
	sections = append(sections, m.GetStyles().Paragraph.Bold(true).Render(locale.ServerSettingsFormDescription))
	sections = append(sections, "")
	sections = append(sections, m.GetStyles().Paragraph.Render(locale.ServerSettingsGeneralHelp))
	sections = append(sections, "")

	fieldIndex := m.GetFocusedIndex()
	fields := m.GetFormFields()

	if fieldIndex >= 0 && fieldIndex < len(fields) {
		field := fields[fieldIndex]
		switch field.Key {
		case "pentagi_license_key":
			sections = append(sections, locale.ServerSettingsLicenseKeyHelp)
		case "pentagi_tenant_id":
			sections = append(sections, locale.ServerSettingsTenantIDHelp)
		case "pentagi_pprof_addr":
			sections = append(sections, locale.ServerSettingsPprofAddrHelp)
		case "pentagi_server_host":
			sections = append(sections, locale.ServerSettingsHostHelp)
		case "pentagi_server_port":
			sections = append(sections, locale.ServerSettingsPortHelp)
		case "pentagi_public_url":
			sections = append(sections, locale.ServerSettingsPublicURLHelp)
		case "pentagi_cors_origins":
			sections = append(sections, locale.ServerSettingsCORSOriginsHelp)
		case "mcp_enabled":
			sections = append(sections, locale.ServerSettingsMCPEnabledHelp)
		case "mcp_server_name":
			sections = append(sections, locale.ServerSettingsMCPServerNameHelp)
		case "mcp_server_version":
			sections = append(sections, locale.ServerSettingsMCPServerVersionHelp)
		case "mcp_api_key":
			sections = append(sections, locale.ServerSettingsMCPAPIKeyHelp)
		case "mcp_write_api_key":
			sections = append(sections, locale.ServerSettingsMCPWriteAPIKeyHelp)
		case "mcp_allow_anonymous":
			sections = append(sections, locale.ServerSettingsMCPAllowAnonymousHelp)
		case "mcp_allowed_origins":
			sections = append(sections, locale.ServerSettingsMCPAllowedOriginsHelp)
		case "mcp_allowed_tools":
			sections = append(sections, locale.ServerSettingsMCPAllowedToolsHelp)
		case "mcp_max_request_bytes":
			sections = append(sections, locale.ServerSettingsMCPMaxRequestBytesHelp)
		case "mcp_read_tool_rate_limit":
			sections = append(sections, locale.ServerSettingsMCPReadToolRateLimitHelp)
		case "mcp_write_tool_rate_limit":
			sections = append(sections, locale.ServerSettingsMCPWriteToolRateLimitHelp)
		case "mcp_approval_mode":
			sections = append(sections, locale.ServerSettingsMCPApprovalModeHelp)
		case "mcp_enable_write_tools":
			sections = append(sections, locale.ServerSettingsMCPEnableWriteToolsHelp)
		case "proxy_url":
			sections = append(sections, locale.ServerSettingsProxyURLHelp)
		case "http_client_timeout":
			sections = append(sections, locale.ServerSettingsHTTPClientTimeoutHelp)
		case "terminal_tool_timeout":
			sections = append(sections, locale.ServerSettingsTerminalToolTimeoutHelp)
		case "external_ssl_ca_path":
			sections = append(sections, locale.ServerSettingsExternalSSLCAPathHelp)
		case "external_ssl_insecure":
			sections = append(sections, locale.ServerSettingsExternalSSLInsecureHelp)
		case "pentagi_ssl_dir":
			sections = append(sections, locale.ServerSettingsSSLDirHelp)
		case "pentagi_data_dir":
			sections = append(sections, locale.ServerSettingsDataDirHelp)
		case "pentagi_cookie_signing_salt":
			sections = append(sections, locale.ServerSettingsCookieSigningSaltHelp)
		case "database_extensions_schema":
			sections = append(sections, locale.ServerSettingsDatabaseExtensionsSchemaHelp)
		case "database_search_path_via_options":
			sections = append(sections, locale.ServerSettingsDatabaseSearchPathViaOptionsHelp)
		default:
			sections = append(sections, locale.ServerSettingsFormOverview)
		}
	}

	return strings.Join(sections, "\n")
}

func (m *ServerSettingsFormModel) HandleSave() error {
	cfg := m.GetController().GetServerSettingsConfig()
	fields := m.GetFormFields()

	newCfg := &controller.ServerSettingsConfig{
		TenantID:                 cfg.TenantID,
		LicenseKey:               cfg.LicenseKey,
		PprofAddr:                cfg.PprofAddr,
		ListenIP:                 cfg.ListenIP,
		ListenPort:               cfg.ListenPort,
		CorsOrigins:              cfg.CorsOrigins,
		CookieSigningSalt:        cfg.CookieSigningSalt,
		ProxyURL:                 cfg.ProxyURL,
		HTTPClientTimeout:        cfg.HTTPClientTimeout,
		TerminalToolTimeout:      cfg.TerminalToolTimeout,
		ExternalSSLCAPath:        cfg.ExternalSSLCAPath,
		ExternalSSLInsecure:      cfg.ExternalSSLInsecure,
		SSLDir:                   cfg.SSLDir,
		DataDir:                  cfg.DataDir,
		PublicURL:                cfg.PublicURL,
		DatabaseExtensionsSchema: cfg.DatabaseExtensionsSchema,
		DatabaseSearchPathViaOpt: cfg.DatabaseSearchPathViaOpt,
		MCPEnabled:               cfg.MCPEnabled,
		MCPServerName:            cfg.MCPServerName,
		MCPServerVersion:         cfg.MCPServerVersion,
		MCPAPIKey:                cfg.MCPAPIKey,
		MCPWriteAPIKey:           cfg.MCPWriteAPIKey,
		MCPAllowAnonymous:        cfg.MCPAllowAnonymous,
		MCPAllowedOrigins:        cfg.MCPAllowedOrigins,
		MCPAllowedTools:          cfg.MCPAllowedTools,
		MCPMaxRequestBytes:       cfg.MCPMaxRequestBytes,
		MCPEnableWriteTools:      cfg.MCPEnableWriteTools,
		MCPReadToolRateLimit:     cfg.MCPReadToolRateLimit,
		MCPWriteToolRateLimit:    cfg.MCPWriteToolRateLimit,
		MCPApprovalMode:          cfg.MCPApprovalMode,
	}

	for _, field := range fields {
		value := strings.TrimSpace(field.Input.Value())

		switch field.Key {
		case "pentagi_license_key":
			if value != "" {
				if info, err := sdk.IntrospectLicenseKey(value); err != nil {
					return fmt.Errorf("invalid license key: %v", err)
				} else if !info.IsValid() {
					return fmt.Errorf("invalid license key")
				}
			}
			newCfg.LicenseKey.Value = value
		case "pentagi_tenant_id":
			if err := (&config.Config{TenantID: value}).ValidateTenantID(); err != nil {
				return err
			}
			newCfg.TenantID.Value = value
		case "pentagi_pprof_addr":
			if value != "" {
				if _, _, err := net.SplitHostPort(value); err != nil {
					return fmt.Errorf("invalid pprof address: must be host:port (e.g., :7777 or 127.0.0.1:7778)")
				}
			}
			newCfg.PprofAddr.Value = value
		case "pentagi_server_host":
			newCfg.ListenIP.Value = value
		case "pentagi_server_port":
			if value != "" {
				if _, err := strconv.Atoi(value); err != nil {
					return fmt.Errorf("invalid port: %s", value)
				}
			}
			newCfg.ListenPort.Value = value
		case "pentagi_public_url":
			newCfg.PublicURL.Value = value
		case "pentagi_cors_origins":
			newCfg.CorsOrigins.Value = value
		case "mcp_enabled":
			if value == "" {
				value = newCfg.MCPEnabled.Default
			}
			if err := validateMCPBoolean(value, locale.ServerSettingsMCPEnabled); err != nil {
				return err
			}
			newCfg.MCPEnabled.Value = value
		case "mcp_server_name":
			if value == "" {
				value = newCfg.MCPServerName.Default
			}
			if err := validateMCPText(value, locale.ServerSettingsMCPServerName, 128); err != nil {
				return err
			}
			newCfg.MCPServerName.Value = value
		case "mcp_server_version":
			if value == "" {
				value = newCfg.MCPServerVersion.Default
			}
			if err := validateMCPText(value, locale.ServerSettingsMCPServerVersion, 128); err != nil {
				return err
			}
			newCfg.MCPServerVersion.Value = value
		case "mcp_api_key":
			if err := validateMCPText(value, locale.ServerSettingsMCPAPIKey, 4096); err != nil {
				return err
			}
			newCfg.MCPAPIKey.Value = value
		case "mcp_write_api_key":
			if err := validateMCPText(value, locale.ServerSettingsMCPWriteAPIKey, 4096); err != nil {
				return err
			}
			newCfg.MCPWriteAPIKey.Value = value
		case "mcp_allow_anonymous":
			if value == "" {
				value = newCfg.MCPAllowAnonymous.Default
			}
			if err := validateMCPBoolean(value, locale.ServerSettingsMCPAllowAnonymous); err != nil {
				return err
			}
			newCfg.MCPAllowAnonymous.Value = value
		case "mcp_allowed_origins":
			normalized, err := normalizeMCPOriginList(value)
			if err != nil {
				return err
			}
			newCfg.MCPAllowedOrigins.Value = normalized
		case "mcp_allowed_tools":
			normalized, err := normalizeMCPToolList(value)
			if err != nil {
				return err
			}
			newCfg.MCPAllowedTools.Value = normalized
		case "mcp_max_request_bytes":
			if value == "" {
				value = newCfg.MCPMaxRequestBytes.Default
			}
			maxBytes, err := strconv.Atoi(value)
			if err != nil || maxBytes < 0 || maxBytes > mcpMaxRequestBytesLimit {
				return fmt.Errorf("%s 必须是 0 至 %d 之间的整数", locale.ServerSettingsMCPMaxRequestBytes, mcpMaxRequestBytesLimit)
			}
			newCfg.MCPMaxRequestBytes.Value = strconv.Itoa(maxBytes)
		case "mcp_read_tool_rate_limit":
			if value == "" {
				value = newCfg.MCPReadToolRateLimit.Default
			}
			rateLimit, err := parseMCPRateLimit(value)
			if err != nil {
				return fmt.Errorf("%s 必须是 0 至 %d 之间的整数", locale.ServerSettingsMCPReadToolRateLimit, mcpToolRateLimitLimit)
			}
			newCfg.MCPReadToolRateLimit.Value = strconv.Itoa(rateLimit)
		case "mcp_write_tool_rate_limit":
			if value == "" {
				value = newCfg.MCPWriteToolRateLimit.Default
			}
			rateLimit, err := parseMCPRateLimit(value)
			if err != nil {
				return fmt.Errorf("%s 必须是 0 至 %d 之间的整数", locale.ServerSettingsMCPWriteToolRateLimit, mcpToolRateLimitLimit)
			}
			newCfg.MCPWriteToolRateLimit.Value = strconv.Itoa(rateLimit)
		case "mcp_approval_mode":
			if value == "" {
				value = newCfg.MCPApprovalMode.Default
			}
			approvalMode, err := parseMCPApprovalMode(value)
			if err != nil {
				return fmt.Errorf("%s 必须是 scope、destructive 或 write", locale.ServerSettingsMCPApprovalMode)
			}
			newCfg.MCPApprovalMode.Value = approvalMode
		case "mcp_enable_write_tools":
			if value == "" {
				value = newCfg.MCPEnableWriteTools.Default
			}
			if err := validateMCPBoolean(value, locale.ServerSettingsMCPEnableWriteTools); err != nil {
				return err
			}
			newCfg.MCPEnableWriteTools.Value = value
		case "proxy_url":
			newCfg.ProxyURL.Value = value
		case "proxy_username":
			newCfg.ProxyUsername = value
		case "proxy_password":
			newCfg.ProxyPassword = value
		case "http_client_timeout":
			if value != "" {
				if timeout, err := strconv.Atoi(value); err != nil {
					return fmt.Errorf("invalid HTTP client timeout: must be a number")
				} else if timeout < 0 {
					return fmt.Errorf("invalid HTTP client timeout: must be >= 0")
				}
			}
			newCfg.HTTPClientTimeout.Value = value
		case "terminal_tool_timeout":
			if value != "" {
				if timeout, err := strconv.Atoi(value); err != nil {
					return fmt.Errorf("invalid terminal tool timeout: must be a number")
				} else if timeout < 0 {
					return fmt.Errorf("invalid terminal tool timeout: must be >= 0")
				}
			}
			newCfg.TerminalToolTimeout.Value = value
		case "external_ssl_ca_path":
			newCfg.ExternalSSLCAPath.Value = value
		case "external_ssl_insecure":
			if value != "" && value != "true" && value != "false" {
				return fmt.Errorf("invalid value for skip SSL verification: must be 'true' or 'false'")
			}
			newCfg.ExternalSSLInsecure.Value = value
		case "pentagi_ssl_dir":
			newCfg.SSLDir.Value = value
		case "pentagi_data_dir":
			newCfg.DataDir.Value = value
		case "pentagi_cookie_signing_salt":
			newCfg.CookieSigningSalt.Value = value
		case "database_extensions_schema":
			if value != "" && !schemaNameRegex.MatchString(value) {
				return fmt.Errorf("invalid extensions schema: must match %s", schemaNameRegex.String())
			}
			newCfg.DatabaseExtensionsSchema.Value = value
		case "database_search_path_via_options":
			if value != "" && value != "true" && value != "false" {
				return fmt.Errorf("invalid value for search path via options: must be 'true' or 'false'")
			}
			newCfg.DatabaseSearchPathViaOpt.Value = value
		}
	}

	if newCfg.MCPEnableWriteTools.Value == "true" && strings.TrimSpace(newCfg.MCPAPIKey.Value) == "" {
		return fmt.Errorf("启用 %s 前必须配置 %s", locale.ServerSettingsMCPEnableWriteTools, locale.ServerSettingsMCPAPIKey)
	}

	if err := m.GetController().UpdateServerSettingsConfig(newCfg); err != nil {
		logger.Errorf("[ServerSettingsFormModel] SAVE: error updating server settings: %v", err)
		return err
	}

	logger.Log("[ServerSettingsFormModel] SAVE: success")
	return nil
}

func validateMCPBoolean(value, fieldName string) error {
	if value != "true" && value != "false" {
		return fmt.Errorf("%s 必须为 true 或 false", fieldName)
	}
	return nil
}

func validateMCPText(value, fieldName string, maxBytes int) error {
	if strings.ContainsAny(value, "\r\n") {
		return fmt.Errorf("%s 不能包含换行符", fieldName)
	}
	if len(value) > maxBytes {
		return fmt.Errorf("%s 长度不能超过 %d 字节", fieldName, maxBytes)
	}
	return nil
}

func parseMCPRateLimit(value string) (int, error) {
	rateLimit, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || rateLimit < 0 || rateLimit > mcpToolRateLimitLimit {
		return 0, fmt.Errorf("rate limit out of range")
	}
	return rateLimit, nil
}

func parseMCPApprovalMode(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "scope", "destructive", "write":
		return value, nil
	default:
		return "", fmt.Errorf("unknown approval mode")
	}
}

func normalizeMCPOriginList(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}

	result := make([]string, 0)
	seen := make(map[string]struct{})
	hasWildcard := false
	for _, rawOrigin := range strings.Split(value, ",") {
		origin := strings.TrimSpace(rawOrigin)
		if origin == "" {
			continue
		}
		if origin == "*" {
			if len(result) > 0 {
				return "", fmt.Errorf("MCP 来源白名单不能把 * 与具体来源混用")
			}
			hasWildcard = true
			continue
		}
		if hasWildcard {
			return "", fmt.Errorf("MCP 来源白名单不能把 * 与具体来源混用")
		}

		parsed, err := url.ParseRequestURI(origin)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || (parsed.Path != "" && parsed.Path != "/") {
			return "", fmt.Errorf("MCP 来源 %q 必须是没有路径和参数的 http/https 来源", origin)
		}
		if !strings.EqualFold(parsed.Scheme, "http") && !strings.EqualFold(parsed.Scheme, "https") {
			return "", fmt.Errorf("MCP 来源 %q 只支持 http 或 https", origin)
		}

		normalized := strings.ToLower(parsed.Scheme) + "://" + strings.ToLower(parsed.Host)
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	if hasWildcard {
		return "*", nil
	}
	return strings.Join(result, ","), nil
}

func normalizeMCPToolList(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}

	result := make([]string, 0)
	seen := make(map[string]struct{})
	for _, rawName := range strings.Split(value, ",") {
		name := strings.TrimSpace(rawName)
		if name == "" {
			continue
		}
		if !mcpToolNameRegex.MatchString(name) {
			return "", fmt.Errorf("MCP 工具名 %q 只能包含字母、数字、点、下划线和连字符，且必须以字母开头", name)
		}
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}
		result = append(result, name)
	}
	return strings.Join(result, ","), nil
}

func (m *ServerSettingsFormModel) HandleReset() {
	m.GetController().ResetServerSettingsConfig()
	m.BuildForm()
}

func (m *ServerSettingsFormModel) OnFieldChanged(fieldIndex int, oldValue, newValue string) {
	// no-op for now
}

func (m *ServerSettingsFormModel) GetFormFields() []FormField {
	return m.BaseScreen.fields
}

func (m *ServerSettingsFormModel) SetFormFields(fields []FormField) {
	m.BaseScreen.fields = fields
}

// Update handles screen-specific input, then delegates to base screen
func (m *ServerSettingsFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if cmd := m.HandleFieldInput(msg); cmd != nil {
			return m, cmd
		}
	}

	cmd := m.BaseScreen.Update(msg)
	return m, cmd
}

// compile-time interface validation
var _ BaseScreenModel = (*ServerSettingsFormModel)(nil)
var _ BaseScreenHandler = (*ServerSettingsFormModel)(nil)
