package routers

import (
	"github.com/gin-gonic/gin"
	v1 "k8soperation/internal/app/controllers/api/v1/user"
)

type RegistryUserRouter struct{}

func NewRegistryUserRouter() *RegistryUserRouter {
	return &RegistryUserRouter{}
}

func (rr *RegistryUserRouter) Inject(router *gin.RouterGroup) {
	uc := v1.NewUserController()
	router.POST("/user/create", uc.Create)
}
