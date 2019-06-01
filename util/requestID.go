package util

import (
	"context"
	"net/http"

	"github.com/gofrs/uuid"
)

type key int

const requestIDKey key = iota

// RequestIDFromContext extracts a Request ID from a context.
func RequestIDFromContext(ctx context.Context) string {
	reqID, ok := ctx.Value(requestIDKey).(string)
	if !ok {
		reqID = ""
	}
	return reqID
}

// AssignRequestID checks the X-Request-ID Header if an ID is present and adds it to the
// request context. If not present, will create a Request ID and add it to the
// request context and headers.
func AssignRequestID(inner http.Handler, logger *Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		ctx := r.Context()
		if reqID == "" {
			reqID = uuid.Must(uuid.NewV4()).String()
			ctx = context.WithValue(ctx, requestIDKey, reqID)
			logger.Debugw("assigned new requestID",
				"requestID", reqID,
			)
			r.Header.Add("X-Request-ID", reqID)
		} else {
			logger.Debugw("found existing requestID",
				"requestID", reqID,
			)
		}
		w.Header().Set("X-Request-ID", reqID)
		inner.ServeHTTP(w, r.WithContext(ctx))
	})
}
