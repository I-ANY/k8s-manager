package middlewares

import (
	"github.com/gin-gonic/gin"
	"k8soperation/internal/app/models"
	"k8soperation/internal/errorcode"
	"k8soperation/pkg/app"
	"k8soperation/pkg/app/response"
	jwt2 "k8soperation/pkg/jwt"
)

func AuthJWT(a *app.App) gin.HandlerFunc {
	m := jwt2.NewManager(
		a.AppSetting.JWTSigningKey,
		a.AppSetting.JWTMaxRefreshTime,
		a.AppSetting.JWTExpireTime,
		a.AppSetting.AppName,
	)

	return func(ctx *gin.Context) {
		rsp := response.NewResponse(ctx)

		tokenStr, err := jwt2.GetTokenFromHeader(ctx)
		if err != nil {
			rsp.ToErrorResponse(errorcode.UnauthorizedTokenError)
			ctx.Abort()
			return
		}

		claims, err := m.ParseToken(tokenStr)
		if err != nil {
			switch err {
			case errorcode.ErrTokenExpired:
				rsp.ToErrorResponse(errorcode.UnauthorizedTokenError)
			default:
				rsp.ToErrorResponse(errorcode.UnauthorizedTokenError)
			}
			ctx.Abort()
			return
		}

		u := models.NewUser().GetUserByID(a.DB, claims.UserID)
		if u.ID == 0 {
			rsp.ToErrorResponse(errorcode.UnauthorizedTokenError)
			ctx.Abort()
			return
		}

		ctx.Set("current_user_id", u.GetStringID())
		ctx.Set("current_user_name", u.Username)
		ctx.Set("current_user", u)

		ctx.Next()
	}
}
