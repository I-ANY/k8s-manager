package routers

import (
	"github.com/gin-gonic/gin"
	"k8soperation/internal/app/controllers/api/v1/user"
)

type UserRouter struct{}

func NewUserRouterV1() *UserRouter {
	return &UserRouter{}
}

func (u *UserRouter) Inject(router *gin.RouterGroup) {
	uc := user.NewUserController()
	router.POST("/user/delete", uc.Delete)
	router.POST("/user/update", uc.Update)
	router.GET("/user/list", uc.List)
}
