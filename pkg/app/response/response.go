// pkg/app/response/response.go
package response

import (
	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/errorcode"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Resp 是统一响应的包装器
type Response struct {
	ctx *gin.Context
}

// swagger:接口文档使用
type UserCreateRespDoc struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// New 创建响应对象：response.New(c)
func NewResponse(c *gin.Context) *Response {
	return &Response{ctx: c}
}

// Success 单对象成功响应：code=0, msg=OK
func (r *Response) Success(data interface{}) {
	if data == nil {
		data = gin.H{}
	}
	r.ctx.JSON(http.StatusOK, gin.H{
		"code":   0,
		"status": "OK",
		"data":   data,
	})
}

func (r *Response) SuccessList(items interface{}, result interface{}) {
	r.ctx.JSON(http.StatusOK, gin.H{
		"code":   0,
		"status": "OK",
		"list":   items,  // 列表数据
		"result": result, // 附加数据（比如 total、message、page/limit 等）
	})
}

func (r *Response) ToErrorResponse(err *errorcode.Error) {
	c := r.ctx
	if err == nil {
		err = errorcode.InvalidParams // 兜底
	}

	// 统一响应体
	payload := gin.H{
		"code": err.Code(),
		"msg":  err.Msg(),
	}
	if d := err.Details(); len(d) > 0 {
		payload["details"] = d
	}
	status := err.StatusCode()

	// —— 结构化错误日志 —— //
	if global.Logger != nil {
		fields := []zap.Field{
			zap.Int("status", status),
			zap.Int("code", err.Code()),
			zap.String("msg", err.Msg()),
			zap.Any("details", err.Details()),
		}
		if c != nil && c.Request != nil {
			fields = append(fields,
				zap.String("method", c.Request.Method),
				zap.String("path", c.FullPath()),              // 路由模板，如 /api/v1/k8s/kube_pod/container_log
				zap.String("uri", c.Request.URL.RequestURI()), // 实际 URI，含 query
				zap.String("query", c.Request.URL.RawQuery),
				zap.String("client_ip", c.ClientIP()),
				zap.String("ua", c.Request.UserAgent()),
				zap.String("request_id", c.GetString("request_id")), // 若你有请求ID中间件
			)
		}
		global.Logger.Error("API error", fields...)
	}

	// 写回并中止
	c.AbortWithStatusJSON(status, payload)
}
