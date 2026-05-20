package middlewares

import (
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8soperation/pkg/logger"
)

func Recovery(l *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				httpRequest, _ := httputil.DumpRequest(c.Request, true)

				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						errStr := strings.ToLower(se.Error())
						if strings.Contains(errStr, "broken pipe") || strings.Contains(errStr, "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				if brokenPipe {
					l.Error(c.Request.URL.Path,
						zap.Time("time", time.Now()),
						zap.Any("error", err),
						zap.String("requests", string(httpRequest)),
					)
					c.Error(err.(error))
					c.Abort()
					return
				}

				l.Error("recovery from panic",
					zap.Time("time", time.Now()),
					zap.Any("error", err),
					zap.String("requests", string(httpRequest)),
					zap.Stack("stacktrace"),
				)

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"message": "服务器内部错误，请稍后再试",
				})
			}
		}()
		c.Next()
	}
}
