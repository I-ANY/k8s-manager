package middlewares

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"k8soperation/pkg/logger"
)

type responseBodyWriter struct {
	body *bytes.Buffer
	gin.ResponseWriter
}

func (w *responseBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *responseBodyWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func RequestMetadata() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()
		}
		c.Set("request_id", requestID)
		c.Set("trace_id", requestID)
		c.Set("ip", c.ClientIP())
		c.Writer.Header().Set("X-Request-ID", requestID)
		c.Next()
	}
}

func Logger(l *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		w := &responseBodyWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = w

		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		start := time.Now()
		c.Next()
		cost := time.Since(start)
		status := c.Writer.Status()

		logFields := []zap.Field{
			zap.String("request_id", c.GetString("request_id")),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.Int("status", status),
			zap.Int64("latency_ms", cost.Milliseconds()),
		}

		if c.Request.Method == http.MethodPost ||
			c.Request.Method == http.MethodPut ||
			c.Request.Method == http.MethodDelete {
			logFields = append(logFields,
				zap.String("requests-body", string(requestBody)),
				zap.String("response-body", w.body.String()),
			)
		}

		switch {
		case status >= http.StatusInternalServerError:
			l.Error("HTTP Error "+strconv.Itoa(status), logFields...)
		case status >= http.StatusBadRequest:
			l.Warn("HTTP Warning "+strconv.Itoa(status), logFields...)
		default:
			l.Debug("HTTP Access Log", logFields...)
		}
	}
}
