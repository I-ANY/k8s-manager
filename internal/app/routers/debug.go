package routers

import "github.com/gin-gonic/gin"

type DebugRouter struct{}

func NewDebugRouter() *DebugRouter {
	return &DebugRouter{}
}

func (r *DebugRouter) Inject(router *gin.RouterGroup) {
	router.GET("/panic-nil", func(c *gin.Context) {
		var p *int
		_ = *p // 故意 panic
	})
}
