package routers

import (
	"github.com/gin-gonic/gin"
	v1 "k8soperation/internal/app/controllers/api/v1/user"
)

type AuthRouter struct{}

func NewAuthRouter() *AuthRouter {
	return &AuthRouter{}
}

func (a *AuthRouter) Inject(r *gin.RouterGroup) {
	ac := v1.NewAuthController()
	r.POST("/auth/login", ac.Login)
}
