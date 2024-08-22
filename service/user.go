package service

import (
	"errors"
	"fmt"
	"gin_gorm/define"
	"gin_gorm/helper"
	"gin_gorm/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"strconv"
	"time"
)

// GetUserDetail
// @Tags 公共方法
// @Summary 用户详情
// @Param identity query string false "user identity"
// @Success 200 {string} json "{"code":200,"msg":"success","data":""}"
// @Router /user-detail [get]
func GetUserDetail(ctx *gin.Context) {
	identity := ctx.Query("identity")
	if identity == "" {
		ctx.JSON(200, gin.H{
			"code": -1,
			"msg":  "输入用户identity",
		})
		return
	}

	userTable := new(models.UserBasic)
	err := models.DB.Omit("password").Where("identity = ?", identity).First(&userTable).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(200, gin.H{
				"code": -1,
				"msg":  "用户不存在",
			})
			return
		}
		ctx.JSON(200, gin.H{
			"code": -1,
			"msg":  "数据库查询错误" + err.Error(),
		})
		return
	}

	ctx.JSON(200, gin.H{
		"code": 200,
		"msg":  "success",
		"data": userTable,
	})
}

// Login
// @Tags 公共方法
// @Summary 用户登录
// @Param name formData string false "name"
// @Param password formData string false "password"
// @Success 200 {string} json "{"code":200,"msg":"success","data":""}"
// @Router /login [post]
func Login(ctx *gin.Context) {
	username := ctx.PostForm("name")
	password := ctx.PostForm("password")
	if username == "" {
		ctx.JSON(200, gin.H{
			"code": -1,
			"msg":  "name信息为空",
		})
		return
	}
	if password == "" {
		ctx.JSON(200, gin.H{
			"code": -1,
			"msg":  "password信息为空",
		})
		return
	}

	password = helper.GetMd5(password)
	fmt.Println("password:", password)

	userBasicData := new(models.UserBasic)
	err := models.DB.Where("name = ? AND password = ?", username, password).First(&userBasicData).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(200, gin.H{
				"code": -1,
				"msg":  "用户不存在",
			})
			return
		}
		ctx.JSON(200, gin.H{
			"code": -1,
			"msg":  "数据库查询错误" + err.Error(),
		})
	}

	//token生成
	token, err := helper.GenerateToken(userBasicData.Identity, userBasicData.Name, userBasicData.IsAdmin)
	if err != nil {
		ctx.JSON(200, gin.H{
			"code": -1,
			"msg":  "token生成失败" + err.Error(),
		})
		return
	}

	ctx.JSON(200, gin.H{
		"code": 200,
		"msg":  "success",
		"data": map[string]interface{}{
			"token": token,
		},
	})
}

// SendCode
// @Tags 公共方法
// @Summary 发送验证码
// @Param email formData string false "email"
// @Success 200 {string} json "{"code":200,"msg":"success","data":""}"
// @Router /send-code [post]
func SendCode(ctx *gin.Context) {
	email := ctx.PostForm("email")
	if email == "" {
		ctx.JSON(200, gin.H{
			"code": -1,
			"msg":  "email为空",
		})
		return
	}

	code := helper.GenerateRandCode()
	models.RDB.Set(ctx, email, code, time.Minute*5)

	fmt.Println(models.RDB.Get(ctx, email))
	err := helper.SendCode(email, code)
	if err != nil {
		ctx.JSON(200, gin.H{
			"code": -1,
			"msg":  "验证码发送失败" + err.Error(),
		})
		return
	}

	ctx.JSON(200, gin.H{
		"code": 200,
		"msg":  "success",
	})
}

// Register
// @Tags 公共方法
// @Summary 用户注册
// @Param email formData string true "email"
// @Param code formData string true "code"
// @Param name formData string true "name"
// @Param password formData string true "password"
// @Param phone formData string false "phone"
// @Success 200 {string} json "{"code":200,"msg":"success","data":""}"
// @Router /register [post]
func Register(ctx *gin.Context) {
	email := ctx.PostForm("email")
	code := ctx.PostForm("code")
	name := ctx.PostForm("name")
	password := ctx.PostForm("password")
	phone := ctx.PostForm("phone")
	if email == "" || code == "" || name == "" {
		ctx.JSON(200, gin.H{
			"code": -1,
			"msg":  "参数不能为空",
		})
		return
	}

	// 验证验证码是否正确
	syscode, err := models.RDB.Get(ctx, email).Result()
	if err != nil {
		log.Printf("验证码错误%v\n", err)
		ctx.JSON(200, gin.H{
			"code": -1,
			"msg":  "验证码错误,请重新获取验证码",
		})
		return
	}
	if syscode != code {
		ctx.JSON(200, gin.H{
			"code": -1,
			"msg":  "验证码错误",
		})
	}

	//写入数据前判断用户是否存在
	var cnt int64
	err = models.DB.Model(new(models.UserBasic)).Where("email = ?", email).Count(&cnt).Error
	if err != nil {
		ctx.JSON(200, gin.H{
			"code": -1,
			"msg":  "Get User Error" + err.Error(),
		})
		return
	}
	if cnt > 0 {
		ctx.JSON(200, gin.H{
			"code": -1,
			"msg":  "用户已存在",
		})
		return
	}

	// 注册用户,写入数据库
	userData := &models.UserBasic{
		Identity: helper.GetUUID(),
		Name:     name,
		Email:    email,
		Phone:    phone,
		Password: helper.GetMd5(password),
	}
	err = models.DB.Create(&userData).Error
	if err != nil {
		ctx.JSON(200, gin.H{
			"code": -1,
			"msg":  "注册失败" + err.Error(),
		})
		return
	}

	//生成token
	token, err := helper.GenerateToken(userData.Identity, userData.Name, userData.IsAdmin)
	if err != nil {
		ctx.JSON(200, gin.H{
			"code": -1,
			"msg":  "生成token失败" + err.Error(),
		})
	}

	ctx.JSON(200, gin.H{
		"code": 200,
		"msg":  "success",
		"data": map[string]interface{}{
			"token": token,
		},
	})
}

// GetUserRank
// @Tags 公共方法
// @Summary 用户排名
// @Param page query string false "page"
// @Param size query string false "size"
// @Success 200 {string} json "{"code":200,"msg":"success","data":""}"
// @Router /user-rank [get]
func GetUserRank(ctx *gin.Context) {
	size, err := strconv.Atoi(ctx.DefaultQuery("size", define.DefaultSize))
	if err != nil {
		log.Println("获得提交问题表页数出错", err)
		return
	}
	page, err := strconv.Atoi(ctx.DefaultQuery("page", define.DefaultPage))
	if err != nil {
		log.Println("获得提交问题表页数出错", err)
		return
	}

	offset := (page - 1) * size
	limit := size

	var userList []*models.UserBasic
	var count int64
	err = models.DB.Model(new(models.UserBasic)).Count(&count).Order("finish_problem_num DESC, submit_num ASC").Offset(offset).Limit(limit).Find(&userList).Error
	if err != nil {
		ctx.JSON(200, gin.H{
			"code": -1,
			"msg":  "Get User Rank Error" + err.Error(),
		})
		return
	}

	ctx.JSON(200, gin.H{
		"code": 200,
		"msg":  "success",
		"data": map[string]interface{}{
			"count": count,
			"list":  userList,
		},
	})

}
