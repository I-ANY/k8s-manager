package valid

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/thedevsaddam/govalidator"
	"go.uber.org/zap"
	"io"
	"k8soperation/global"
	"net/http"
	"strings"
)

type ValidationErrorResponse struct {
	Message string              `json:"message"`
	Errors  map[string][]string `json:"errors"`
}

// ---------------------------------------------
// 通用验证入口
// ---------------------------------------------
// ValidatorFunc 验证函数类型签名
// 任何函数只要符合 (数据, gin.Context) -> 错误 map 的形式，就可以作为校验处理函数
// 返回值 map[string][]string ：key 是字段名，value 是该字段的所有错误提示
type ValidatorFunc func(interface{}, *gin.Context) map[string][]string

// Validate 控制器里调用的通用入口
// 使用示例：
//
//	var req requests.UserSaveRequest
//	if !app.Validate(c, &req, requests.UserSave) {
//	    return // 如果校验失败，这里直接 return，中断后续逻辑
//	}

func Validate(c *gin.Context, obj interface{}, handler ValidatorFunc) bool {
	devMode := global.ServerSetting.RunMode != "release"
	const maxLogBytes = 4096

	// 打印一下 obj 的实际类型
	global.Logger.Info("DEBUG TYPE",
		zap.String("path", c.Request.URL.Path),
		zap.String("obj-type", fmt.Sprintf("%T", obj)),
	)

	global.Logger.Info("DEBUG URL",
		zap.String("raw", c.Request.URL.RawQuery),
	)

	// 1) 非 GET 才考虑打印原始体，且仅在 dev
	if devMode && c.Request.Method != http.MethodGet {
		raw, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(raw)) // 还原
		if len(raw) > maxLogBytes {
			raw = raw[:maxLogBytes]
		}
		global.Logger.Info("REQ RAW BODY",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("content_type", c.GetHeader("Content-Type")),
			zap.ByteString("body", raw),
		)
	}

	// 2) 绑定
	var err error
	if c.Request.Method == http.MethodGet {
		err = c.ShouldBindQuery(obj)
	} else {
		ct := strings.ToLower(c.GetHeader("Content-Type"))
		switch {
		case strings.Contains(ct, "application/json"):
			err = c.ShouldBindBodyWith(obj, binding.JSON)
		case strings.Contains(ct, "application/x-www-form-urlencoded"),
			strings.Contains(ct, "multipart/form-data"):
			err = c.ShouldBind(obj)
		case strings.Contains(ct, "text/plain"): // 可选扩展: 接收 text/plain
			body, _ := io.ReadAll(c.Request.Body)
			// 这里要你自己定义如何赋值，例如 obj.(*MyStruct).Content = string(body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body)) // 还原
		default:
			if err = c.ShouldBindBodyWith(obj, binding.JSON); err != nil {
				err = c.ShouldBind(obj)
			}
		}
	}
	if err != nil {
		global.Logger.Warn("BIND ERROR",
			zap.String("path", c.Request.URL.Path),
			zap.Error(err),
		)
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{
			"message": "请求解析错误",
			"error":   err.Error(),
		})
		return false
	}

	// 3) 绑定后的对象只在 dev 打印，且脱敏/限长
	if devMode {
		b, _ := json.Marshal(obj)
		if len(b) > maxLogBytes {
			b = b[:maxLogBytes]
		}
		global.Logger.Info("BOUND PARAM",
			zap.String("path", c.Request.URL.Path),
			zap.ByteString("param", b),
		)
	}

	if handler == nil {
		// 没传 handler 直接返回 true，避免 panic
		global.Logger.Warn("VALIDATE SKIPPED (nil handler)",
			zap.String("path", c.Request.URL.Path),
		)
		return true
	}

	// 4) 业务校验
	if errs := handler(obj, c); len(errs) > 0 {
		// 打印具体字段和错误信息
		global.Logger.Warn("VALIDATION ERROR",
			zap.String("path", c.Request.URL.Path),
			zap.Int("count", len(errs)),
			zap.Any("errors", errs), // 打印具体错误列表
		)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "参数校验失败",
			"errors":  errs,
		})
		return false
	}

	return true
}

// ---------------------------------------------
// 直接调用 govalidator 库的封装
// ---------------------------------------------
// ValidateOptions 底层调用 govalidator 库，直接传入 struct、rules、messages
// 返回值：map[string][]string（字段对应的所有错误提示）
func ValidateOptions(data interface{}, rules govalidator.MapData, messages govalidator.MapData) map[string][]string {
	opts := govalidator.Options{
		Data:          data,     // 要校验的数据（struct 指针）
		Rules:         rules,    // 校验规则
		TagIdentifier: "valid",  // struct tag 使用的标识符，例如 `valid:"required"`
		Messages:      messages, // 自定义错误提示
	}
	return govalidator.New(opts).ValidateStruct()
}

// ---------------------------------------------
// 自定义扩展：校验两次输入的密码是否一致
// ---------------------------------------------
func ValidatePasswordConfirm(password, PasswordConfirm string, errs map[string][]string) map[string][]string {
	if password != PasswordConfirm {
		errs["PasswordConfirm"] = append(errs["PasswordConfirm"], "两次输入的密码不一致")
	}
	return errs
}
