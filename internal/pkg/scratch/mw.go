package scratch

import (
	"github.com/Yakwilik/MRGAbackend/internal/logger"
	"github.com/google/uuid"
	"net/http"
	"time"
)

var allowedHosts = []string{}
var now = time.Now()

func CorsMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("ApiVersion", now.String())
		if r.Header.Get("Origin") == "" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		}
		//if allowedOrigin(r.Header.Get("Origin")) {
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, ResponseType")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		//}
		if r.Method == "OPTIONS" {
			return
		}
		h.ServeHTTP(w, r)
	})
}

type ResponseWriter struct {
	ResponseWriter http.ResponseWriter
	statusCode     int
}

func (rw *ResponseWriter) Header() http.Header {
	return rw.ResponseWriter.Header()
}

func (rw *ResponseWriter) Write(bytes []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	return rw.ResponseWriter.Write(bytes)
}

func (rw *ResponseWriter) Flush() {
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		statusCode:     0,
	}

}

func (rw *ResponseWriter) WriteHeader(code int) {
	if rw.statusCode != 0 {
		return
	}
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
			r.Header.Set("X-Request-ID", requestID) // Устанавливаем для внутреннего использования
		}
		w.Header().Set("X-Request-ID", requestID) // Отправляем обратно клиенту
		ctx := logger.WithRequestID(r.Context(), requestID)

		rw := NewResponseWriter(w)
		next.ServeHTTP(rw, r.WithContext(ctx))
		duration := time.Since(start)

		logger.Info(ctx, "request_info",
			"method", r.Method,
			"path", r.URL.Path,
			"status_code", rw.statusCode,
			"duration", duration,
			"duration_string", duration.String(),
			"duration_milliseconds", duration.Milliseconds(),
			"ip", r.RemoteAddr,
			"timestamp", time.Now().Format(time.RFC3339))
	})
}
