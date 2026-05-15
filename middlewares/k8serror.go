package middlewares

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func K8sError() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 没有错误或已经写过响应 -> 不处理
		if len(c.Errors) == 0 || c.Writer.Written() {
			return
		}

		ge := c.Errors.Last()
		if ge == nil || ge.Err == nil {
			return
		}
		err := ge.Err

		// TODO: 在这里用 apierrors.IsNotFound/IsForbidden 等做分类映射
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"kube_message_error": err.Error(),
		})
	}
}
