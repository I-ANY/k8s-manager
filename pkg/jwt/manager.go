package jwt

import (
	"time"

	jwtpkg "github.com/golang-jwt/jwt"
	"k8soperation/pkg/utils"
)

type Manager struct {
	SignKey    []byte
	MaxRefresh time.Duration
	AppName    string
	ExpireTime time.Duration
}

func NewManager(signKey string, maxRefreshMinutes int, expireMinutes int, appName string) *Manager {
	return &Manager{
		SignKey:    []byte(signKey),
		MaxRefresh: time.Duration(maxRefreshMinutes) * time.Minute,
		AppName:    appName,
		ExpireTime: time.Duration(expireMinutes) * time.Minute,
	}
}

func (m *Manager) IssueToken(userID, userName string) (string, error) {
	now := utils.TimenowInTimezone().Unix()
	exp := m.expireAtTime()

	claims := Claims{
		UserID:   userID,
		UserName: userName,
		StandardClaims: jwtpkg.StandardClaims{
			NotBefore: now,
			IssuedAt:  now,
			ExpiresAt: exp,
			Issuer:    m.AppName,
		},
	}
	return m.createToken(claims)
}

func (m *Manager) expireAtTime() int64 {
	return utils.TimenowInTimezone().Add(m.ExpireTime).Unix()
}

func (m *Manager) createToken(claims Claims) (string, error) {
	t := jwtpkg.NewWithClaims(jwtpkg.SigningMethodHS256, claims)
	return t.SignedString(m.SignKey)
}
