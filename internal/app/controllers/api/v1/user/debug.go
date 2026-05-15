package user

import (
	"encoding/json"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"k8soperation/internal/app/models"
	"k8soperation/pkg/utils"
)

type DebugController struct{}

func NewDebugController() *DebugController {
	return &DebugController{}
}

// DebuGSessionInfo godoc
// @Summary 获取调试 Session 信息
// @Description 读取当前请求 Cookie 中的 SessionID，并返回 Redis 里保存的 LoginSessionInfo
// @Tags 调试工具
// @Produce json
// @Param user query string false "用户名（可选，用于计算 MD5 作为 session key），默认 admin"
// @Success 200 {object} map[string]interface{} "返回会话信息或 empty"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /debug/session [get]
func (d *DebugController) DebuGSessionInfo(ctx *gin.Context) {
	s := sessions.Default(ctx)

	// 用固定 key，便于调试（比如 admin 用户）
	md5Key := utils.EncodeMD5("admin") // ⚠️ 如果你有多个用户，可以通过 query 参数传进来
	raw := s.Get(md5Key)
	if raw == nil {
		ctx.JSON(200, gin.H{"session": "empty"})
		return
	}

	// 反序列化 JSON
	var info models.LoginSessionInfo
	if err := json.Unmarshal([]byte(raw.(string)), &info); err != nil {
		ctx.JSON(500, gin.H{"error": "failed to parse session data"})
		return
	}

	ctx.JSON(200, gin.H{
		"username":   info.Username,
		"token":      info.Token,
		"login_time": info.LoginTime,
	})
}
