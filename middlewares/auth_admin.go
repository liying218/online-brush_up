package middlewares

import (
	"gin_gorm/helper"
	"github.com/gin-gonic/gin"
	"net/http"
)

// AuthAdminCheck
// 验证用户是不是admin(验证中间件)
func AuthAdminCheck() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.GetHeader("token")
		userClaim, err := helper.ParseToken(token)
		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "Unauthorized authorization",
			})
			ctx.Abort()
			return
		}
		if userClaim == nil || userClaim.IsAdmin != 1 {
			ctx.JSON(http.StatusOK, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "Unauthorized admin",
			})
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
