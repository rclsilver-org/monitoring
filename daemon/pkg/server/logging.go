package server

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const (
	RequestID = "request-id"

	HeaderRequestID = "X-Request-ID"
)

type statusRecorder struct {
	http.ResponseWriter
	Status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.Status = status
	r.ResponseWriter.WriteHeader(status)
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New()
		start := time.Now()

		ctx := r.Context()
		ctx = context.WithValue(ctx, RequestID, id.String())
		r = r.WithContext(ctx)

		h := w.Header()
		h.Set(HeaderRequestID, id.String())

		recorder := &statusRecorder{
			ResponseWriter: w,
			Status:         200,
		}

		next.ServeHTTP(recorder, r)

		elapsed := time.Since(start)

		logrus.WithContext(r.Context()).WithFields(logrus.Fields{
			"duration":       elapsed,
			"status_code":    recorder.Status,
			"request_id":     id.String(),
			"method":         r.Method,
			"uri":            r.RequestURI,
			"remote_address": getRemoteAddress(r),
		}).Infof("%s %s", r.Method, r.RequestURI)
	})
}
