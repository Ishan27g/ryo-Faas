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

// MethodWithOtel parses otel baggage and updates the span
func MethodWithOtel(w http.ResponseWriter, r *http.Request) {

	now := time.Now()

	ctx := r.Context()
	span := trace.SpanFromContext(ctx)

	bag := baggage.FromContext(ctx)
	fmt.Println("baggage extracted from context is ", bag.String())

	uk := attribute.Key("username")
	ik := attribute.Key("id")

	username := bag.Member("username").Value()
	id := bag.Member("id").Value()

	span.AddEvent("handling for user -", trace.WithAttributes(uk.String(username)))
	span.AddEvent("with id ", trace.WithAttributes(ik.String(username)))

	w.WriteHeader(http.StatusAccepted)
	_, _ = io.WriteString(w, "Cool!\n"+bag.String())

	span.SetAttributes(attribute.String("username", username))
	span.SetAttributes(attribute.String("id", id))
	span.SetAttributes(attribute.Key("time").String(time.Since(now).String()))
}
