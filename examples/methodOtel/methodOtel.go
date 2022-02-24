package method1

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/trace"
)

// MethodWithOtel
func MethodWithOtel(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	uk := attribute.Key("username")

	ctx := r.Context()
	span := trace.SpanFromContext(ctx)

	bag := baggage.FromContext(ctx)
	fmt.Println("bag is ", bag.String())

	username := bag.Member("username").Value()
	span.AddEvent("handling this...", trace.WithAttributes(uk.String(username)))

	w.WriteHeader(http.StatusAccepted)
	_, _ = io.WriteString(w, "Hello, world!\n"+bag.String())

	span.SetAttributes(attribute.String("username", username))
	span.SetAttributes(attribute.Key("time").String(time.Since(now).String()))
}
