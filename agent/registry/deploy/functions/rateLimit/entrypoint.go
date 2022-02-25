package rateLimit

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
)

type response struct {
	Id            string `json:"id"`
	Allowed       bool   `json:"allowed"`
	RequestNumber int    `json:"requestNumber"`
	RequestLimit  int    `json:"requestLimit"`
	Interval      string `json:"interval"`
}

// RateLimit exposes a redis based cache over http
// ?key=someId
func RateLimit(w http.ResponseWriter, r *http.Request) {

	key := r.URL.Query()["key"]
	allowed, requestNumber, id, span := cache.Allow(r.Context(), key[0])
	// defer span.End()

	span.SetAttributes(attribute.String("id", id))
	// if too many requests, add to span
	if !allowed {
		span.SetAttributes(attribute.Key(id).Bool(allowed))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	var rsp = response{
		Id:            id,
		Allowed:       allowed,
		RequestNumber: requestNumber,
		RequestLimit:  RequestLimit,
		Interval:      Interval.String(),
	}
	span.AddEvent(fmt.Sprintf("%v", rsp))

	json.NewEncoder(w).Encode(&rsp)
}
