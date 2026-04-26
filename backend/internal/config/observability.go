package config

import (
	"strconv"
	"strings"
)

// ObservabilityConfig holds OpenTelemetry and monitoring configuration.
type ObservabilityConfig struct {
	OTelEnabled        bool    // Master switch for all OTel instrumentation
	ServiceName        string  // Service name for traces (default: "go-react-starter-api")
	ServiceVersion     string  // Application version for traces
	Environment        string  // Deployment environment (e.g., "production", "development")
	PrometheusEnabled  bool    // Enable /metrics endpoint
	MetricsAuthToken   string  // Bearer token required to scrape /metrics in production (METRICS_AUTH_TOKEN)
	TracingEnabled     bool    // Enable distributed tracing
	LogsEnabled        bool    // Enable OTel log pipeline (sends logs to collector/Azure)
	TraceEndpoint      string  // OTLP endpoint (default: "localhost:4317")
	TraceProtocol      string  // "grpc" (default, for Tempo) or "http" (for Azure/cloud)
	TraceInsecure      bool    // Use insecure connection (default: true for dev)
	TraceSampleRate    float64 // Sampling rate 0.0-1.0 (default: 1.0)
	OTLPHeaders        map[string]string // Extra headers for OTLP exporter (e.g., auth tokens)
}

// LoadObservability reads observability configuration from environment variables.
// If APPLICATIONINSIGHTS_CONNECTION_STRING is set and no explicit OTLP endpoint is
// configured, the Azure Monitor OTLP ingestion endpoint and auth header are
// auto-configured from the connection string.
func LoadObservability() ObservabilityConfig {
	sampleRate := 1.0
	if parsed, err := strconv.ParseFloat(getEnv("OTEL_TRACES_SAMPLER_ARG", "1.0"), 64); err == nil {
		sampleRate = parsed
	}

	cfg := ObservabilityConfig{
		OTelEnabled:       parseBool(getEnv("OTEL_ENABLED", "false")),
		ServiceName:       getEnv("OTEL_SERVICE_NAME", "go-react-starter-api"),
		ServiceVersion:    getEnv("APP_VERSION", "dev"),
		Environment:       getEnv("ENVIRONMENT", "development"),
		PrometheusEnabled: parseBool(getEnv("PROMETHEUS_ENABLED", "false")),
		MetricsAuthToken:  getEnv("METRICS_AUTH_TOKEN", ""),
		TracingEnabled:    parseBool(getEnv("OTEL_TRACING_ENABLED", "false")),
		LogsEnabled:      parseBool(getEnv("OTEL_LOGS_ENABLED", "false")),
		TraceEndpoint:     getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
		TraceProtocol:     getEnv("OTEL_EXPORTER_OTLP_PROTOCOL", "grpc"),
		TraceInsecure:     parseBool(getEnv("OTEL_EXPORTER_OTLP_INSECURE", "true")),
		TraceSampleRate:   sampleRate,
		OTLPHeaders:       parseOTLPHeaders(getEnv("OTEL_EXPORTER_OTLP_HEADERS", "")),
	}

	// Auto-configure Azure Monitor from connection string when no explicit endpoint is set.
	aiConnStr := getEnv("APPLICATIONINSIGHTS_CONNECTION_STRING", "")
	if aiConnStr != "" && cfg.TraceEndpoint == "" {
		cfg.applyAzureMonitorConfig(aiConnStr)
	}

	// Default endpoint for local dev if nothing was configured.
	if cfg.TraceEndpoint == "" {
		cfg.TraceEndpoint = "localhost:4317"
	}

	return cfg
}

// applyAzureMonitorConfig detects that Azure Monitor is configured and sets the
// OTLP endpoint to the local OTel Collector sidecar (embedded in the same container
// via supervisord). The collector handles the Azure-proprietary ingestion protocol.
//
// Azure Application Insights does NOT support direct OTLP ingestion (/v1/traces
// returns 404). The OTel Collector with the azuremonitor exporter is required.
func (c *ObservabilityConfig) applyAzureMonitorConfig(connStr string) {
	parts := parseConnectionString(connStr)
	if parts["InstrumentationKey"] == "" {
		return
	}

	// Point to the embedded OTel Collector sidecar running on localhost:4317.
	// The collector uses APPLICATIONINSIGHTS_CONNECTION_STRING from its own env.
	c.TraceEndpoint = "localhost:4317"
	c.TraceProtocol = "grpc"
	c.TraceInsecure = true
}

// parseConnectionString splits a semicolon-delimited key=value connection string.
func parseConnectionString(connStr string) map[string]string {
	result := make(map[string]string)
	for _, pair := range strings.Split(connStr, ";") {
		k, v, ok := strings.Cut(strings.TrimSpace(pair), "=")
		if ok && k != "" {
			result[k] = v
		}
	}
	return result
}

// parseOTLPHeaders parses a comma-separated list of key=value pairs.
// Example: "x-api-key=abc123,Authorization=Bearer token"
func parseOTLPHeaders(s string) map[string]string {
	if s == "" {
		return nil
	}
	headers := make(map[string]string)
	for _, pair := range strings.Split(s, ",") {
		k, v, ok := strings.Cut(strings.TrimSpace(pair), "=")
		if ok && k != "" {
			headers[k] = v
		}
	}
	return headers
}

// LogSection returns observability configuration values for structured logging.
func (c ObservabilityConfig) LogSection() ConfigSection {
	metricsAuthConfigured := c.MetricsAuthToken != ""
	return ConfigSection{
		Name: "Observability",
		Values: map[string]interface{}{
			"otel_enabled":               c.OTelEnabled,
			"service_name":               c.ServiceName,
			"service_version":            c.ServiceVersion,
			"environment":                c.Environment,
			"prometheus_enabled":         c.PrometheusEnabled,
			"metrics_auth_token_set":     metricsAuthConfigured,
			"tracing_enabled":            c.TracingEnabled,
			"logs_enabled":               c.LogsEnabled,
			"trace_endpoint":             c.TraceEndpoint,
			"trace_protocol":             c.TraceProtocol,
			"trace_insecure":             c.TraceInsecure,
			"trace_sample_rate":          c.TraceSampleRate,
			"otlp_headers_set":          len(c.OTLPHeaders) > 0,
			"azure_monitor":             getEnv("APPLICATIONINSIGHTS_CONNECTION_STRING", "") != "",
		},
	}
}
