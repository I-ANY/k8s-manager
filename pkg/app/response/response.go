package response

import (
	"go.uber.org/zap"
	"net/http"

	"github.com/gin-gonic/gin"
	"k8soperation/internal/errorcode"
	appctx "k8soperation/pkg/app"
)

type Response struct {
	ctx *gin.Context
}

type UserCreateRespDoc struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func NewResponse(c *gin.Context) *Response {
	return &Response{ctx: c}
}

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
		"list":   items,
		"result": result,
	})
}

func (r *Response) ToErrorResponse(err *errorcode.Error) {
	c := r.ctx
	if err == nil {
		err = errorcode.InvalidParams
	}

	payload := gin.H{
		"code": err.Code(),
		"msg":  err.Msg(),
	}
	if d := err.Details(); len(d) > 0 {
		payload["details"] = d
	}
	status := err.StatusCode()

	if a := appctx.FromContext(c); a != nil && a.Logger != nil {
		fields := []zap.Field{
			zap.Int("status", status),
			zap.Int("code", err.Code()),
			zap.String("msg", err.Msg()),
			zap.Any("details", err.Details()),
		}
		if c.Request != nil {
			fields = append(fields,
				zap.String("method", c.Request.Method),
				zap.String("path", c.FullPath()),
				zap.String("uri", c.Request.URL.RequestURI()),
				zap.String("query", c.Request.URL.RawQuery),
				zap.String("client_ip", c.ClientIP()),
				zap.String("ua", c.Request.UserAgent()),
				zap.String("request_id", c.GetString("request_id")),
			)
		}
		a.Logger.Error("API error", fields...)
	}

	c.AbortWithStatusJSON(status, payload)
}
