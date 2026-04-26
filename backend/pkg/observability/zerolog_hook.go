package observability

import (
	"context"

	"github.com/rs/zerolog"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/trace"
)

// ZerologOTelHook is a zerolog.Hook that forwards log entries to the OTel
// LoggerProvider. Each log record includes the message, severity, and any
// active trace/span context for correlation.
type ZerologOTelHook struct {
	logger otellog.Logger
}

// NewZerologOTelHook creates a hook that bridges zerolog to OTel logs.
// The instrumentationName identifies the source in telemetry backends.
func NewZerologOTelHook(instrumentationName string) ZerologOTelHook {
	return ZerologOTelHook{
		logger: global.GetLoggerProvider().Logger(instrumentationName),
	}
}

// Run implements zerolog.Hook. It emits an OTel log record for each zerolog event.
func (h ZerologOTelHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	if level == zerolog.NoLevel || level == zerolog.Disabled {
		return
	}

	ctx := e.GetCtx()
	if ctx == nil {
		ctx = context.Background()
	}

	var record otellog.Record
	record.SetBody(otellog.StringValue(msg))
	record.SetSeverity(zerologToOTelSeverity(level))
	record.SetSeverityText(level.String())

	h.logger.Emit(ctx, record)
}

// zerologToOTelSeverity maps zerolog levels to OTel log severity levels.
func zerologToOTelSeverity(level zerolog.Level) otellog.Severity {
	switch level {
	case zerolog.TraceLevel:
		return otellog.SeverityTrace
	case zerolog.DebugLevel:
		return otellog.SeverityDebug
	case zerolog.InfoLevel:
		return otellog.SeverityInfo
	case zerolog.WarnLevel:
		return otellog.SeverityWarn
	case zerolog.ErrorLevel:
		return otellog.SeverityError
	case zerolog.FatalLevel:
		return otellog.SeverityFatal
	case zerolog.PanicLevel:
		return otellog.SeverityFatal2
	default:
		return otellog.SeverityInfo
	}
}

// SpanContextFromZerologEvent extracts trace context for correlation.
// This is used internally — OTel's Emit automatically picks up span context
// from the context passed to it.
func SpanContextFromContext(ctx context.Context) (traceID, spanID string) {
	sc := trace.SpanContextFromContext(ctx)
	if sc.HasTraceID() {
		traceID = sc.TraceID().String()
	}
	if sc.HasSpanID() {
		spanID = sc.SpanID().String()
	}
	return
}
