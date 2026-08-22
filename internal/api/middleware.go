package api

import (
	"log"
	"net/http"
	"time"
)

// statusRecorder 记录响应码供访问日志使用。
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// LogMiddleware 记录请求日志。
func LogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		log.Printf("%s %s -> %d (%s)", r.Method, r.URL.Path, rec.status, time.Since(start))
	})
}

// RecoverMiddleware 兜底异常转成 500。
func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("recovered in %s %s: %v", r.Method, r.URL.Path, rec)
				writeJSON(w, http.StatusInternalServerError, errorBody{Error: "internal error"})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// Chain 按顺序包装中间件。
func Chain(h http.Handler, mids ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mids) - 1; i >= 0; i-- {
		h = mids[i](h)
	}
	return h
}
