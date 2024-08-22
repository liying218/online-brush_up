package router

import (
	_ "gin_gorm/docs"
	"gin_gorm/middlewares"
	"gin_gorm/service"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Router() *gin.Engine {
	r := gin.Default()

	//swagger配置
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	//配置路由
	//问题
	r.GET("/problem-list", service.GetProblemList)
	r.GET("/problem-detail", service.GetProblemDetail)

	//用户
	r.GET("/user-detail", service.GetUserDetail)
	r.POST("/login", service.Login)
	r.POST("/send-code", service.SendCode)
	r.POST("/register", service.Register)
	//用户排行磅
	r.GET("/user-rank", service.GetUserRank)

	//问题提交表
	r.POST("/submit-list", service.GetSubmitList)

	//管理员私有方法
	admin := r.Group("/admin", middlewares.AuthAdminCheck())
	admin.POST("/problem-create", service.ProblemCreate)
	admin.PUT("/problem-modify", service.ProblemModify)
	admin.GET("/category-list", service.GetCategoryList)
	admin.POST("/category-create", service.CategoryCreate)
	admin.PUT("/category-modify", service.CategoryModify)
	admin.DELETE("/category-delete", service.CategoryDelete)

	//用户私有方法
	user := r.Group("/user", middlewares.AuthUserCheck())
	user.POST("/submit-code", service.SubmitCode)
	return r
}
