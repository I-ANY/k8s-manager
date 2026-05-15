package requests

import (
	"github.com/gin-gonic/gin"
	"github.com/thedevsaddam/govalidator"
	"k8soperation/pkg/valid"
)

type AuthLoginRequest struct {
	Username string `json:"username" form:"username" valid:"username"`
	Password string `json:"password" form:"password" valid:"password"`
}

func NewAuthLoginRequest() *AuthLoginRequest {
	return &AuthLoginRequest{}
}

func ValidAuthLoginRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"username": []string{"required"},
		"password": []string{"required", "min:6"},
	}
	messages := govalidator.MapData{
		"username": []string{
			"required: 用户名为必填字段,字段为 username",
		},
		"password": []string{
			"required: 密码为必填字段,字段为 password",
			"min:密码长度需大于 6",
		},
	}

	// 校验入参
	return valid.ValidateOptions(data, rules, messages)
}
