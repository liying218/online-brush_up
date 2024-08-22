package middlewares

import (
	"gin_gorm/helper"
	"github.com/gin-gonic/gin"
	"net/http"
)

func AuthUserCheck() gin.HandlerFunc {
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
		if userClaim == nil {
			ctx.JSON(http.StatusOK, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "Unauthorized admin",
			})
			ctx.Abort()
			return
		}
		ctx.Set("user", userClaim)
		ctx.Next()
	}
}
