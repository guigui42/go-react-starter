// Package logger provides structured logging utilities for the the API using Zerolog.
// It includes request/response time tracking, request IDs, and consistent log formatting.
package logger

import (
	"context"
	"io"
	stdlog "log"
	"os"
	"strings"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/trace"
)

// Logger wraps zerolog.Logger to provide application-specific logging functionality
type Logger struct {
	zlog zerolog.Logger
}

// New creates a new Logger instance with the specified output and level
func New(output io.Writer, level zerolog.Level) *Logger {
	if output == nil {
		output = os.Stdout
	}

	zlog := zerolog.New(output).
		With().
		Timestamp().
		Logger().
		Level(level)

	return &Logger{
		zlog: zlog,
	}
}

// Default creates a logger that writes to stdout with INFO level
func Default() *Logger {
	return New(os.Stdout, zerolog.InfoLevel)
}

// AddHook attaches a zerolog.Hook to the logger.
func (l *Logger) AddHook(hook zerolog.Hook) {
	l.zlog = l.zlog.Hook(hook)
}

// Debug logs a debug message
func (l *Logger) Debug(message string) {
	l.zlog.Debug().Msg(message)
}

// Info logs an info message
func (l *Logger) Info(message string) {
	l.zlog.Info().Msg(message)
}

// Warn logs a warning message
func (l *Logger) Warn(message string) {
	l.zlog.Warn().Msg(message)
}

// Error logs an error message with optional error object
func (l *Logger) Error(message string, err error) {
	if err != nil {
		l.zlog.Error().Err(err).Msg(message)
	} else {
		l.zlog.Error().Msg(message)
	}
}

// WithRequest logs an HTTP request with metadata
func (l *Logger) WithRequest(requestID, method, path string, statusCode int, duration time.Duration) {
	l.zlog.Info().
		Str("request_id", requestID).
		Str("method", method).
		Str("path", path).
		Int("status_code", statusCode).
		Int64("duration_ms", duration.Milliseconds()).
		Msgf("%s %s", method, path)
}

// WithMetadata logs a message with additional metadata
func (l *Logger) WithMetadata(level zerolog.Level, message string, metadata map[string]interface{}) {
	event := l.zlog.WithLevel(level)
	for key, value := range metadata {
		event = event.Interface(key, value)
	}
	event.Msg(message)
}

// RequestLogger creates a logger with request context
func (l *Logger) RequestLogger(requestID string) *RequestLogger {
	return &RequestLogger{
		zlog: l.zlog.With().Str("request_id", requestID).Logger(),
	}
}

// RequestLoggerWithTrace creates a logger with request context and OTel trace correlation.
// If a valid trace context exists in ctx, trace_id and span_id are injected into every log entry.
func (l *Logger) RequestLoggerWithTrace(ctx context.Context, requestID string) *RequestLogger {
	logCtx := l.zlog.With().Str("request_id", requestID)

	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		logCtx = logCtx.Str("trace_id", spanCtx.TraceID().String())
	}
	if spanCtx.HasSpanID() {
		logCtx = logCtx.Str("span_id", spanCtx.SpanID().String())
	}

	return &RequestLogger{
		zlog: logCtx.Logger(),
	}
}

// RequestLogger is a logger with request context
type RequestLogger struct {
	zlog zerolog.Logger
}

// Debug logs a debug message with request ID
func (rl *RequestLogger) Debug(message string) {
	rl.zlog.Debug().Msg(message)
}

// Info logs an info message with request ID
func (rl *RequestLogger) Info(message string) {
	rl.zlog.Info().Msg(message)
}

// Warn logs a warning message with request ID
func (rl *RequestLogger) Warn(message string) {
	rl.zlog.Warn().Msg(message)
}

// Error logs an error message with request ID
func (rl *RequestLogger) Error(message string, err error) {
	if err != nil {
		rl.zlog.Error().Err(err).Msg(message)
	} else {
		rl.zlog.Error().Msg(message)
	}
}

// WithMetadata logs a message with additional metadata and request ID
func (rl *RequestLogger) WithMetadata(level zerolog.Level, message string, metadata map[string]interface{}) {
	event := rl.zlog.WithLevel(level)
	for key, value := range metadata {
		event = event.Interface(key, value)
	}
	event.Msg(message)
}

// ZLogger returns the underlying zerolog.Logger for advanced use cases
// that require direct access to zerolog's structured logging API.
func (l *Logger) ZLogger() zerolog.Logger {
	return l.zlog
}

// dualWriter writes raw JSON to the ring buffer (for admin API) and
// console-formatted output to stdout. zerolog always produces JSON internally;
// ConsoleWriter reads that JSON and reformats it for human consumption.
type dualWriter struct {
	console zerolog.ConsoleWriter
	buffer  *RingBuffer
}

func (d *dualWriter) Write(p []byte) (n int, err error) {
	d.buffer.Write(p)
	return d.console.Write(p)
}

// stdlibWriter adapts zerolog for use as Go stdlib log output.
// Stdlib log.Printf calls are captured and emitted as zerolog Info events.
type stdlibWriter struct {
	zlog zerolog.Logger
}

func (w stdlibWriter) Write(p []byte) (n int, err error) {
	msg := strings.TrimRight(string(p), "\n")
	w.zlog.Info().Msg(msg)
	return len(p), nil
}

// NewWithBuffer creates a Logger that writes to both stdout and a ring buffer.
// When consoleMode is true, stdout gets human-readable output (for development).
// When consoleMode is false, stdout gets JSON output (for production log aggregation).
// The ring buffer always receives JSON for the admin API.
// Stdlib log.Printf is redirected through zerolog for unified output.
func NewWithBuffer(level zerolog.Level, consoleMode bool) (*Logger, *RingBuffer) {
	buffer := NewRingBufferFromEnv()

	var output io.Writer
	if consoleMode {
		isTTY := isatty.IsTerminal(os.Stdout.Fd())
		console := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "2006/01/02 15:04:05",
			NoColor:    !isTTY,
		}
		if !isTTY {
			// When piped/redirected (e.g. Azure Log Stream), the platform
			// already prepends its own timestamp — omit ours to avoid duplication.
			console.PartsExclude = []string{zerolog.TimestampFieldName}
		}
		output = &dualWriter{console: console, buffer: buffer}
	} else {
		output = io.MultiWriter(os.Stdout, buffer)
	}

	l := New(output, level)

	// Redirect Go stdlib log through zerolog for unified format
	stdlog.SetFlags(0)
	stdlog.SetOutput(stdlibWriter{zlog: l.zlog})

	// Set zerolog's global logger so that packages importing
	// "github.com/rs/zerolog/log" also go through the console writer.
	zlog.Logger = l.zlog

	return l, buffer
}
