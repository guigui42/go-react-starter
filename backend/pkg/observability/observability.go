// Package observability provides OpenTelemetry SDK initialization for the the API.
// It configures metrics (Prometheus + OTLP push), tracing (OTLP/gRPC), and logging
// exporters based on runtime configuration. When disabled, noop providers are installed
// so the rest of the application can call OTel APIs without nil checks.
package observability

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
	"go.opentelemetry.io/otel/trace"
)

// Config holds the configuration for OpenTelemetry providers.
type Config struct {
	ServiceName       string
	ServiceVersion    string
	Environment       string
	OTelEnabled       bool
	PrometheusEnabled bool
	TracingEnabled    bool
	LogsEnabled       bool              // Enable OTel log pipeline
	TraceEndpoint     string
	TraceProtocol     string            // "grpc" or "http"
	TraceInsecure     bool
	TraceSampleRate   float64
	OTLPHeaders       map[string]string // Extra headers for OTLP exporter (e.g., Azure auth)
}

// Init initialises OpenTelemetry providers based on cfg and returns a shutdown
// function that must be called on application exit to flush pending telemetry.
// If OTel is disabled, noop providers are registered so callers can use the OTel
// API without nil checks.
func Init(ctx context.Context, cfg Config) (shutdown func(context.Context) error, err error) {
	if !cfg.OTelEnabled {
		return func(context.Context) error { return nil }, nil
	}

	res, err := newResource(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create otel resource: %w", err)
	}

	var shutdownFns []func(context.Context) error

	if cfg.PrometheusEnabled {
		mp, err := newMeterProvider(ctx, res, cfg)
		if err != nil {
			return nil, fmt.Errorf("create meter provider: %w", err)
		}
		otel.SetMeterProvider(mp)
		shutdownFns = append(shutdownFns, mp.Shutdown)
	}

	if cfg.TracingEnabled {
		tp, err := newTracerProvider(ctx, res, cfg)
		if err != nil {
			return nil, fmt.Errorf("create tracer provider: %w", err)
		}
		otel.SetTracerProvider(tp)
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		))
		shutdownFns = append(shutdownFns, tp.Shutdown)
	}

	if cfg.LogsEnabled {
		lp, err := newLoggerProvider(ctx, res, cfg)
		if err != nil {
			return nil, fmt.Errorf("create logger provider: %w", err)
		}
		global.SetLoggerProvider(lp)
		shutdownFns = append(shutdownFns, lp.Shutdown)
	}

	return func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFns {
			err = errors.Join(err, fn(ctx))
		}
		return err
	}, nil
}

// PrometheusHandler returns the HTTP handler that serves Prometheus metrics.
// Mount this at the /metrics endpoint.
func PrometheusHandler() http.Handler {
	return promhttp.Handler()
}

// Tracer returns a named tracer from the global TracerProvider.
func Tracer(name string) trace.Tracer {
	return otel.Tracer(name)
}

// Meter returns a named meter from the global MeterProvider.
func Meter(name string) metric.Meter {
	return otel.Meter(name)
}

// RegisterDBPoolMetrics registers asynchronous gauge callbacks that expose
// database/sql connection pool statistics as OTel metrics.
func RegisterDBPoolMetrics(sqlDB *sql.DB) error {
	m := Meter("app.db")

	openConns, err := m.Int64ObservableGauge("db.pool.open_connections",
		metric.WithDescription("Number of open connections in the pool"),
	)
	if err != nil {
		return fmt.Errorf("create open_connections gauge: %w", err)
	}

	idleConns, err := m.Int64ObservableGauge("db.pool.idle_connections",
		metric.WithDescription("Number of idle connections in the pool"),
	)
	if err != nil {
		return fmt.Errorf("create idle_connections gauge: %w", err)
	}

	maxOpen, err := m.Int64ObservableGauge("db.pool.max_open_connections",
		metric.WithDescription("Maximum number of open connections allowed"),
	)
	if err != nil {
		return fmt.Errorf("create max_open_connections gauge: %w", err)
	}

	waitCount, err := m.Int64ObservableGauge("db.pool.wait_count",
		metric.WithDescription("Total number of connections waited for"),
	)
	if err != nil {
		return fmt.Errorf("create wait_count gauge: %w", err)
	}

	waitDuration, err := m.Int64ObservableGauge("db.pool.wait_duration_ms",
		metric.WithDescription("Total time blocked waiting for a new connection in milliseconds"),
	)
	if err != nil {
		return fmt.Errorf("create wait_duration_ms gauge: %w", err)
	}

	_, err = m.RegisterCallback(
		func(_ context.Context, o metric.Observer) error {
			stats := sqlDB.Stats()
			o.ObserveInt64(openConns, int64(stats.OpenConnections))
			o.ObserveInt64(idleConns, int64(stats.Idle))
			o.ObserveInt64(maxOpen, int64(stats.MaxOpenConnections))
			o.ObserveInt64(waitCount, stats.WaitCount)
			o.ObserveInt64(waitDuration, stats.WaitDuration.Milliseconds())
			return nil
		},
		openConns, idleConns, maxOpen, waitCount, waitDuration,
	)
	if err != nil {
		return fmt.Errorf("register db pool callback: %w", err)
	}

	return nil
}

// newResource builds the OTel resource that describes this service.
// Merges with resource.Default() to include OS, process, and SDK attributes.
func newResource(ctx context.Context, cfg Config) (*resource.Resource, error) {
	return resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			semconv.DeploymentEnvironmentName(cfg.Environment),
		),
	)
}

// newMeterProvider creates a MeterProvider with a Prometheus exporter (pull-based
// for /metrics endpoint) and, when tracing is enabled, an OTLP periodic exporter
// that pushes metrics to the OTel Collector for forwarding to Azure Monitor.
func newMeterProvider(ctx context.Context, res *resource.Resource, cfg Config) (*sdkmetric.MeterProvider, error) {
	promExporter, err := prometheus.New()
	if err != nil {
		return nil, fmt.Errorf("create prometheus exporter: %w", err)
	}

	opts := []sdkmetric.Option{
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(promExporter),
	}

	// When an OTLP endpoint is configured, also push metrics periodically
	// so they reach Azure Monitor (or any OTLP-compatible backend).
	if cfg.TracingEnabled && cfg.TraceEndpoint != "" {
		otlpExporter, err := newOTLPMetricExporter(ctx, cfg)
		if err != nil {
			return nil, fmt.Errorf("create otlp metric exporter: %w", err)
		}
		opts = append(opts, sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(otlpExporter, sdkmetric.WithInterval(30*time.Second)),
		))
	}

	return sdkmetric.NewMeterProvider(opts...), nil
}

// newTracerProvider creates a TracerProvider that exports spans via OTLP.
// Supports both gRPC (for Tempo/local) and HTTP (for Azure Monitor/cloud) protocols.
func newTracerProvider(ctx context.Context, res *resource.Resource, cfg Config) (*sdktrace.TracerProvider, error) {
	var exporter sdktrace.SpanExporter
	var err error

	switch cfg.TraceProtocol {
	case "http":
		exporter, err = newOTLPHTTPExporter(ctx, cfg)
	default: // "grpc"
		exporter, err = newOTLPGRPCExporter(ctx, cfg)
	}
	if err != nil {
		return nil, err
	}

	sampler := sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.TraceSampleRate))

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithSampler(sampler),
	)
	return tp, nil
}

// newOTLPGRPCExporter creates an OTLP/gRPC span exporter (for Tempo, Jaeger, etc.).
func newOTLPGRPCExporter(ctx context.Context, cfg Config) (sdktrace.SpanExporter, error) {
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(cfg.TraceEndpoint),
	}
	if cfg.TraceInsecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}
	if len(cfg.OTLPHeaders) > 0 {
		opts = append(opts, otlptracegrpc.WithHeaders(cfg.OTLPHeaders))
	}

	exporter, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create otlp/grpc trace exporter: %w", err)
	}
	return exporter, nil
}

// newOTLPHTTPExporter creates an OTLP/HTTP span exporter (for Azure Monitor, cloud backends).
func newOTLPHTTPExporter(ctx context.Context, cfg Config) (sdktrace.SpanExporter, error) {
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(cfg.TraceEndpoint),
	}
	if cfg.TraceInsecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}
	if len(cfg.OTLPHeaders) > 0 {
		opts = append(opts, otlptracehttp.WithHeaders(cfg.OTLPHeaders))
	}

	exporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create otlp/http trace exporter: %w", err)
	}
	return exporter, nil
}

// newLoggerProvider creates a LoggerProvider that exports logs via OTLP/gRPC
// to the same endpoint as traces (OTel Collector or Tempo).
func newLoggerProvider(ctx context.Context, res *resource.Resource, cfg Config) (*sdklog.LoggerProvider, error) {
	opts := []otlploggrpc.Option{
		otlploggrpc.WithEndpoint(cfg.TraceEndpoint),
	}
	if cfg.TraceInsecure {
		opts = append(opts, otlploggrpc.WithInsecure())
	}
	if len(cfg.OTLPHeaders) > 0 {
		opts = append(opts, otlploggrpc.WithHeaders(cfg.OTLPHeaders))
	}

	exporter, err := otlploggrpc.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create otlp/grpc log exporter: %w", err)
	}

	lp := sdklog.NewLoggerProvider(
		sdklog.WithResource(res),
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
	)
	return lp, nil
}

// newOTLPMetricExporter creates an OTLP/gRPC metric exporter that pushes metrics
// to the OTel Collector (which forwards them to Azure Monitor).
func newOTLPMetricExporter(ctx context.Context, cfg Config) (sdkmetric.Exporter, error) {
	opts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint(cfg.TraceEndpoint),
	}
	if cfg.TraceInsecure {
		opts = append(opts, otlpmetricgrpc.WithInsecure())
	}
	if len(cfg.OTLPHeaders) > 0 {
		opts = append(opts, otlpmetricgrpc.WithHeaders(cfg.OTLPHeaders))
	}

	exporter, err := otlpmetricgrpc.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create otlp/grpc metric exporter: %w", err)
	}
	return exporter, nil
}
