package routers

import (
	"github.com/gin-gonic/gin"
	v1 "k8soperation/internal/app/controllers/api/v1/user"
)

type DebugSessionRouter struct {
}

func NewDebugSessionRouter() *DebugSessionRouter {
	return &DebugSessionRouter{}
}

func (d *DebugSessionRouter) Inject(router *gin.RouterGroup) {
	uc := v1.NewDebugController()
	router.GET("/debug/session", uc.DebuGSessionInfo)
}
