package initialize

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"k8soperation/global"
)

// SetupSession 初始化 Gin 的 Redis Session Store，并设置 Cookie 选项
func SetupSession() error {
	if global.CacheSetting.Username == "" {
		global.Logger.Error("redis 用户名不能为空")
		return fmt.Errorf("redis username is empty")
	}

	store, err := redis.NewStore(
		global.CacheSetting.MaxConnect,     // 连接池大小
		global.CacheSetting.Network,        // "tcp" / "unix"
		global.CacheSetting.Address,        // "host:port" 或 socket 路径
		global.CacheSetting.Username,       // redis 用户名
		global.CacheSetting.Password,       // 仅密码，username 无效
		[]byte(global.CacheSetting.Secret), // 用于签名/加密的密钥
	)
	if err != nil {
		return fmt.Errorf("new redis session store failed: %w", err)
	}

	// === 统一设置 Cookie 行为 ===
	// 生产（release）建议 Secure=true；本地 http 调试要设为 false，否则浏览器不会保存 Cookie
	secure := global.ServerSetting.RunMode == "release"

	// SameSite 默认用 Lax（同站点能带 Cookie）；如果你有“跨域前端调用”，需要改成 None，并保持 Secure=true（浏览器要求）
	sameSite := http.SameSiteLaxMode
	// 如需跨域携带 Cookie，放开下一行并确保 HTTPS：
	// sameSite = http.SameSiteNoneMode

	store.Options(sessions.Options{
		Path:     "/",           // 整站有效
		MaxAge:   7 * 24 * 3600, // 7 天（秒）；如需“关浏览器即失效”，设为 0
		HttpOnly: true,          // 前端 JS 不可读，防 XSS
		Secure:   secure,        // 生产 HTTPS 下应为 true
		SameSite: sameSite,      // 跨域场景需 None（且 Secure=true）
		// Domain: "example.com",     // 如需跨子域共享再配置
	})

	global.SessionStore = store
	return nil
}
