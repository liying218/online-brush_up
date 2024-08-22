package service

import (
	"gin_gorm/define"
	"gin_gorm/helper"
	"gin_gorm/models"
	"github.com/gin-gonic/gin"
	"log"
	"strconv"
)

// GetCategoryList
// @Tags 管理员私有方法
// @Summary 获取分类列表
// @Param page query int false "请输入当前页，默认第一页"
// @Param size query int false "请输入每页数量，默认20条"
// @Param name query string false "请输入类别名称"
// @Param token header string true "token"
// @Param identity query string false "identity"
// @Success 200 {string} json "{"code":200,"msg":"success","data":""}"
// @Router /admin/category-list [get]
func GetCategoryList(ctx *gin.Context) {
	size, err := strconv.Atoi(ctx.DefaultQuery("size", define.DefaultSize))
	if err != nil {
		log.Println("获得问题表页数出错", err)
		return
	}
	page, err := strconv.Atoi(ctx.DefaultQuery("page", define.DefaultPage))
	if err != nil {
		log.Println("获得问题表页数出错", err)
		return
	}

	offset := (page - 1) * size
	limit := size

	name := ctx.Query("name")

	var categoryList []*models.CategoryBasic
	var count int64
	models.DB.Model(new(models.CategoryBasic)).Where("name=", name).Count(&count)
	err = models.DB.Model(new(models.CategoryBasic)).Where("name=", name).Find(&categoryList).Offset(offset).Limit(limit).Error
	if err != nil {
		log.Println("获得问题表出错", err)
		return
	}
	ctx.JSON(200, gin.H{
		"code": 200,
		"msg":  "success",
		"data": map[string]interface{}{
			"list":  categoryList,
			"count": count,
		},
	})
}

// CategoryCreate
// @Tags 管理员私有方法
// @Summary 创建分类
// @Param token header string true "token"
// @Param name formData string true " name"
// @Param parent_id formData string false "parent_id"
// @Success 200 {string} json "{"code":200,"msg":"success","data":""}"
// @Router /admin/category-creat [post]
func CategoryCreate(ctx *gin.Context) {
	name := ctx.PostForm("name")
	parentId, _ := strconv.Atoi(ctx.PostForm("parent_id"))
	if name == "" {
		ctx.JSON(200, gin.H{
			"code": -1,
			"msg":  "必填参数不能为空",
		})
		return
	}

	categoryBasics := &models.CategoryBasic{
		Name:     name,
		ParentId: parentId,
		Identity: helper.GetUUID(),
	}
	err := models.DB.Create(categoryBasics).Error
	if err != nil {
		log.Println("creat category failed", err)
		ctx.JSON(200, gin.H{
			"code": -1,
			"msg":  "创建失败",
		})
		return
	}

	ctx.JSON(200, gin.H{
		"code": 200,
		"msg":  "创建成功",
		"data": ""})
}

// CategoryModify
// @Tags 管理员私有方法
// @Summary 分类修改
// @Param token header string true "token"
// @Param identity formData string true "identity"
// @Param name formData string true " name"
// @Param parent_id formData string false "parent_id"
// @Success 200 {string} json "{"code":200,"msg":"success","data":""}"
// @Router /admin/category-modify [put]
func CategoryModify(ctx *gin.Context) {
	identity := ctx.PostForm("identity")
	name := ctx.PostForm("name")
	parentId, _ := strconv.Atoi(ctx.PostForm("parent_id"))
	if name == "" || identity == "" {
		ctx.JSON(200, gin.H{
			"code": -1,
			"msg":  "必填参数不能为空",
		})
		return
	}

	categoryBasic := &models.CategoryBasic{
		Name:     name,
		ParentId: parentId,
	}
	err := models.DB.Where("identity=?", identity).Updates(categoryBasic).Error
	if err != nil {
		log.Println("creat category failed", err)
		ctx.JSON(200, gin.H{
			"code": -1,
			"msg":  "修改失败",
		})
		return
	}

	ctx.JSON(200, gin.H{
		"code": 200,
		"msg":  "修改成功",
		"data": "",
	})
}

// CategoryDelete
// @Tags 管理员私有方法
// @Summary 分类删除
// @Param token header string true "token"
// @Param identity query string true "identity"
// @Success 200 {string} json "{"code":200,"msg":"success","data":""}"
// @Router /admin/category-delete [delete]
func CategoryDelete(ctx *gin.Context) {
	identity := ctx.Query("identity")
	if identity == "" {
		ctx.JSON(200, gin.H{
			"code": -1,
			"msg":  "必填参数不能为空",
		})
		return
	}

	//判断问题表中是否有问题属于要删除的类型，如果有的话不能删除
	var count int64
	err := models.DB.Model(new(models.ProblemCategory)).Where("category_id=(SELECT id FROM category_basic WHERE identity =? )", identity).Count(&count).Error
	if err != nil {
		log.Println("delete category failed", err)
		ctx.JSON(200, gin.H{
			"code": -1,
			"msg":  "获取分类关联表失败",
		})
	}
	if count > 0 {
		ctx.JSON(200, gin.H{
			"code": -1,
			"msg":  "该分类下有题目，不能删除",
		})
	}

	err = models.DB.Where("identity=?", identity).Delete(new(models.CategoryBasic)).Error
	if err != nil {
		log.Println("delete category failed", err)
		ctx.JSON(200, gin.H{
			"code": -1,
			"msg":  "删除失败",
		})
		return
	}

	ctx.JSON(200, gin.H{
		"code": 200,
		"msg":  "删除成功",
	})
}
