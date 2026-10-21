package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := c.GetHeader("X-Request-ID")
		if reqID == "" {
			reqID = uuid.NewString()
		}
		c.Set(RequestIDKey, reqID)
		c.Header("X-Request-ID", reqID)
		c.Next()
	}
}

func SlogLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		reqID, _ := c.Get(RequestIDKey)

		attrs := []any{
			slog.String("request_id", reqID.(string)),
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.String("query", query),
			slog.Int("status", status),
			slog.Duration("latency", latency),
			slog.String("ip", c.ClientIP()),
		}

		if status >= 500 {
			slog.Error("http request error", attrs...)
		} else if status >= 400 {
			slog.Warn("http client error", attrs...)
		} else {
			slog.Info("http request", attrs...)
		}
	}
}
