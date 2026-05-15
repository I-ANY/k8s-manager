package services

import (
	"k8soperation/internal/app/models"
	"k8soperation/internal/app/requests"
)

// UserLogin 用户登录
func (s *Services) UserLogin(param *requests.AuthLoginRequest) (*models.User, error) {
	return s.dao.UserGetByName(param.Username)
}
