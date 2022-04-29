package tracing

import (
	"context"
	"strings"

	"github.com/Ishan27g/go-utils/noop/noop"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
)

// NoopSpanFromGin returns a span with noop-ctx if url has a query => `?noop=true`
// otherwise it returns the request's span (if any)
func NoopSpanFromGin(c *gin.Context) (trace.Span, context.Context) {
	isNoop := strings.EqualFold(c.Query("noop"), "true")
	ctx := noop.NewCtxWithNoop(c.Request.Context(), isNoop)
	return trace.SpanFromContext(ctx), ctx
}
