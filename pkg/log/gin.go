package log

import (
	"errors"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ipfs/go-log/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// https://github.com/gin-contrib/zap

var ginLogger = log.Logger("GIN")

type GinConfig struct {
	// SkipPaths is an url path array which logs are not written.
	// Optional.
	SkipPaths []string

	// Skip is a Skipper that indicates which logs should not be written.
	// Optional.
	Skip Skipper
}

// Skipper is a function to skip logs based on provided Context
type Skipper func(c *gin.Context) bool

func GinzapWithConfig(conf *GinConfig, activeReqs *int32) gin.HandlerFunc {
	var skip map[string]struct{}

	if length := len(conf.SkipPaths); length > 0 {
		skip = make(map[string]struct{}, length)

		for _, path := range conf.SkipPaths {
			skip[path] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		atomic.AddInt32(activeReqs, 1)
		defer atomic.AddInt32(activeReqs, -1)

		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log only when it is not being skipped
		if _, ok := skip[path]; ok || (conf.Skip != nil && conf.Skip(c)) {
			return
		}

		end := time.Now()
		latency := end.Sub(start)
		fields := []zapcore.Field{
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("protocol", c.Request.Proto),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.Duration("latency", latency),
			// zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
		}
		if len(c.Errors) > 0 {
			for _, e := range c.Errors.Errors() {
				ginLogger.SugaredLogger.Desugar().Error(e, fields...)
			}

			// Replace gin's default error response - "AbortWithStatus" with JSON format
			err := c.Errors.Last()
			var statusCode int
			var message string

			switch err.Type {
			case gin.ErrorTypePrivate:
				statusCode = http.StatusInternalServerError
				message = "Internal server error"
			case gin.ErrorTypePublic:
				statusCode = err.Meta.(int)
				message = err.Error()
			default:
				statusCode = http.StatusInternalServerError
				message = "Unknown error"
			}

			// default error code
			if statusCode == 0 {
				statusCode = http.StatusInternalServerError
			}

			c.AbortWithStatusJSON(
				statusCode,
				gin.H{
					"code":    statusCode,
					"message": message,
				},
			)
		} else {
			ginLogger.SugaredLogger.Desugar().Info(path, fields...)
		}
	}
}

func defaultHandleRecovery(c *gin.Context, _ any) {
	// c.AbortWithStatus(http.StatusInternalServerError)
	c.AbortWithStatusJSON(
		http.StatusInternalServerError,
		gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Recovery from panic",
		},
	)
}

// stack means whether output the stack info.
// The stack info is easy to find where the error occurs but the stack info is too large.
func GinzapRecovery(stack bool) gin.HandlerFunc {
	return CustomRecovery(stack, defaultHandleRecovery)
}

// stack means whether output the stack info.
// The stack info is easy to find where the error occurs but the stack info is too large.
func CustomRecovery(stack bool, recovery gin.RecoveryFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					var se *os.SyscallError
					if errors.As(ne, &se) {
						seStr := strings.ToLower(se.Error())
						if strings.Contains(seStr, "broken pipe") ||
							strings.Contains(seStr, "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					ginLogger.Errorw(c.Request.URL.Path,
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
					// If the connection is dead, we can't write a status to it.
					// c.Error(err.(error)) //nolint: errcheck
					c.AbortWithStatusJSON(
						http.StatusInternalServerError,
						gin.H{
							"code":    http.StatusInternalServerError,
							"message": "connection is broken",
						},
					)
					return
				}

				if stack {
					ginLogger.Errorw("[Recovery from panic]",
						zap.Time("time", time.Now()),
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
						zap.String("stack", string(debug.Stack())),
					)
				} else {
					ginLogger.Errorw("[Recovery from panic]",
						zap.Time("time", time.Now()),
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
				}
				recovery(c, err)
			}
		}()
		c.Next()
	}
}
