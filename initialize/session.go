package initialize

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"k8soperation/pkg/app"
)

func SetupSession(a *app.App) error {
	if a.CacheSetting.Username == "" {
		return fmt.Errorf("redis username is empty")
	}

	store, err := redis.NewStore(
		a.CacheSetting.MaxConnect,
		a.CacheSetting.Network,
		a.CacheSetting.Address,
		a.CacheSetting.Username,
		a.CacheSetting.Password,
		[]byte(a.CacheSetting.Secret),
	)
	if err != nil {
		return fmt.Errorf("new redis session store failed: %w", err)
	}

	secure := a.ServerSetting.RunMode == "release"
	sameSite := http.SameSiteLaxMode

	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   7 * 24 * 3600,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})

	a.SessionStore = store
	return nil
}
