package middleware

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"

	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
)

// OTelRouteLabeler is a Chi middleware that injects the matched route pattern
// into otelhttp's labeler so that the http.route attribute appears on HTTP
// server metrics exported by otelhttp.NewHandler.
// It also renames the active span to "METHOD /route/pattern" so that
// Application Insights (and other backends) show per-route endpoint names
// instead of the generic server name.
func OTelRouteLabeler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)

		// After the request is handled, Chi has resolved the route pattern.
		rctx := chi.RouteContext(r.Context())
		if rctx != nil {
			route := rctx.RoutePattern()
			if route != "" {
				if labeler, ok := otelhttp.LabelerFromContext(r.Context()); ok {
					labeler.Add(semconv.HTTPRoute(route))
				}
				// Rename the span so trace backends (Application Insights, Tempo)
				// show individual routes instead of the generic otelhttp server name.
				span := trace.SpanFromContext(r.Context())
				span.SetName(fmt.Sprintf("%s %s", r.Method, route))
			}
		}
	})
}
