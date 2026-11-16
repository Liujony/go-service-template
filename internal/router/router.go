// Package router builds the HTTP handler and the cross-cutting middleware.
package router

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"go-service-template/internal/handler"
	"go-service-template/pkg/response"
)

// New returns the full HTTP handler: middleware, health, dashboard and v1 API.
func New(items *handler.ItemHandler, log *slog.Logger, version string) *gin.Engine {
	// Gin prints every route in debug mode; a library template should not.
	if os.Getenv("GIN_MODE") == "" && gin.Mode() == gin.DebugMode {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	engine.Use(requestID(), accessLog(log), recoverer(log))

	engine.GET("/healthz", func(c *gin.Context) {
		response.OK(c, http.StatusOK, gin.H{"status": "ok", "version": version})
	})
	engine.GET("/", handler.Home)

	v1 := engine.Group("/api/v1")
	{
		v1.POST("/items", items.Create)
		v1.GET("/items", items.List)
		v1.GET("/items/:id", items.Get)
		v1.PUT("/items/:id", items.Update)
		v1.DELETE("/items/:id", items.Delete)
	}
	return engine
}

// requestID gives every request a correlation ID, echoed in the response header
// and in every log line for that request.
func requestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			var buf [8]byte
			if _, err := rand.Read(buf[:]); err == nil {
				id = hex.EncodeToString(buf[:])
			} else {
				id = "unknown"
			}
		}
		c.Set("request_id", id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}

// accessLog writes one structured line per request.
func accessLog(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		c.Next()
		log.Info("request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration_ms", float64(time.Since(started).Microseconds())/1000.0,
			"request_id", c.GetString("request_id"),
			"bytes", c.Writer.Size(),
		)
	}
}

// recoverer turns a panic into a logged 500 instead of a dropped connection.
func recoverer(log *slog.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		log.Error("panic recovered",
			"path", c.Request.URL.Path,
			"request_id", c.GetString("request_id"),
			"panic", recovered,
		)
		response.Fail(c, http.StatusInternalServerError, "internal_error", "something went wrong")
	})
}
