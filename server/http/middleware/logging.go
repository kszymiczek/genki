package middleware

import (
	"net/http"
	"time"

	"github.com/lukasjarosch/genki/logger"
)

func Logging(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Infof("incoming request to %s %s", r.Method, r.URL)
		defer func(started time.Time) {
			logger.Infof("served request to %s %s (took %s)", r.Method, r.URL, time.Since(started).String())
		}(time.Now())
		handler.ServeHTTP(w, r)
	})
}
