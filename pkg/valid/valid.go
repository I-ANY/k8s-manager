package valid

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	appctx "k8soperation/pkg/app"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/thedevsaddam/govalidator"
	"go.uber.org/zap"
)

type ValidationErrorResponse struct {
	Message string              `json:"message"`
	Errors  map[string][]string `json:"errors"`
}

type ValidatorFunc func(interface{}, *gin.Context) map[string][]string

func Validate(c *gin.Context, obj interface{}, handler ValidatorFunc) bool {
	a := appctx.FromContext(c)
	devMode := a != nil && a.ServerSetting != nil && a.ServerSetting.RunMode != "release"
	const maxLogBytes = 4096

	l := a.Logger

	l.Info("DEBUG TYPE",
		zap.String("path", c.Request.URL.Path),
		zap.String("obj-types", fmt.Sprintf("%T", obj)),
	)

	l.Info("DEBUG URL",
		zap.String("raw", c.Request.URL.RawQuery),
	)

	if devMode && c.Request.Method != http.MethodGet {
		raw, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(raw))
		if len(raw) > maxLogBytes {
			raw = raw[:maxLogBytes]
		}
		l.Info("REQ RAW BODY",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("content_type", c.GetHeader("Content-Type")),
			zap.ByteString("body", raw),
		)
	}

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
		case strings.Contains(ct, "text/plain"):
			body, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		default:
			if err = c.ShouldBindBodyWith(obj, binding.JSON); err != nil {
				err = c.ShouldBind(obj)
			}
		}
	}
	if err != nil {
		l.Warn("BIND ERROR",
			zap.String("path", c.Request.URL.Path),
			zap.Error(err),
		)
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{
			"message": "请求解析错误",
			"error":   err.Error(),
		})
		return false
	}

	if devMode {
		b, _ := json.Marshal(obj)
		if len(b) > maxLogBytes {
			b = b[:maxLogBytes]
		}
		l.Info("BOUND PARAM",
			zap.String("path", c.Request.URL.Path),
			zap.ByteString("param", b),
		)
	}

	if handler == nil {
		l.Warn("VALIDATE SKIPPED (nil handler)",
			zap.String("path", c.Request.URL.Path),
		)
		return true
	}

	if errs := handler(obj, c); len(errs) > 0 {
		l.Warn("VALIDATION ERROR",
			zap.String("path", c.Request.URL.Path),
			zap.Int("count", len(errs)),
			zap.Any("errors", errs),
		)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "参数校验失败",
			"errors":  errs,
		})
		return false
	}

	return true
}

func ValidateOptions(data interface{}, rules govalidator.MapData, messages govalidator.MapData) map[string][]string {
	opts := govalidator.Options{
		Data:          data,
		Rules:         rules,
		TagIdentifier: "valid",
		Messages:      messages,
	}
	return govalidator.New(opts).ValidateStruct()
}

func ValidatePasswordConfirm(password, PasswordConfirm string, errs map[string][]string) map[string][]string {
	if password != PasswordConfirm {
		errs["PasswordConfirm"] = append(errs["PasswordConfirm"], "两次输入的密码不一致")
	}
	return errs
}
