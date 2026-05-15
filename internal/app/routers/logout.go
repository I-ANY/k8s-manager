package routers

import (
	"github.com/gin-gonic/gin"
	v1 "k8soperation/internal/app/controllers/api/v1/user"
)

type AuthLogoutRouter struct {
}

func NewAuthLogoutRouter() *AuthLogoutRouter {
	return &AuthLogoutRouter{}
}

func (a *AuthLogoutRouter) Inject(r *gin.RouterGroup) {
	uc := v1.NewAuthControllerLogout()
	r.POST("/auth/logout", uc.Logout)
}
