package service

import (
	"errors"
	"gin_gorm/define"
	"gin_gorm/helper"
	"gin_gorm/models"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strconv"
)

// GetProblemList
// @Tags 公共方法
// @Summary 问题列表
// @Param page query int false "请输入当前页，默认第一页"
// @Param size query int false "请输入每页数量，默认10条"
// @Param keyword query string false "请输入关键字"
// @Param category_identity query string false "请输入分类id"
// @Success 200 {string} json "{"code":200,"msg":"success","data":""}"
// @Router /problem-list [get]
func GetProblemList(ctx *gin.Context) {
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

	keyword := ctx.Query("keyword")
	categoryIdentity := ctx.Query("category_identity")
	data := make([]*models.ProblemBasic, 0)

	tx := models.GetProblemList(keyword, categoryIdentity)

	//获得问题数量
	var count int64
	err = tx.Count(&count).Error
	if err != nil {
		log.Println("获得问题表页数出错", err)
	}

	//分页查询
	err = tx.Limit(limit).Offset(offset).Find(&data).Error
	if err != nil {
		log.Println("获得问题表页数出错", err)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":  200,
		"msg":   "success",
		"count": count,
		"data":  data,
	})
}

// GetProblemDetail
// @Tags 公共方法
// @Summary 问题详情
// @Param identity query string false "problem identity"
// @Success 200 {string} json "{"code":200,"msg":"success","data":""}"
// @Router /problem-detail [get]
func GetProblemDetail(ctx *gin.Context) {
	identity := ctx.Query("identity")
	if identity == "" {
		ctx.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "identity不能为空",
		})
		return
	}

	problem := new(models.ProblemBasic)
	err := models.DB.Debug().Where("identity=?", identity).
		Preload("ProblemCategories").
		Preload("ProblemCategories.CategoryBasic").
		First(&problem).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "问题不存在",
			})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Get ProblemBasicDetail Error" + err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": problem,
	})
	return
}

// ProblemCreate
// @Tags 管理员私有方法
// @Summary 问题创建
// @Param token header string true "token"
// @Param title formData string true "title"
// @Param content formData string true "content"
// @Param max_runtime formData int false "max_runtime"
// @Param max_memory formData int false "max_memory"
// @Param category_ids formData []string false "category_ids" collectionFormat(multi)
// @Param test_cases formData []string false "test_cases" collectionFormat(multi)
// @Success 200 {string} json "{"code":200,"msg":"success","data":""}"
// @Router /admin/problem-create [post]
func ProblemCreate(ctx *gin.Context) {
	title := ctx.PostForm("title")
	content := ctx.PostForm("content")
	maxRuntime, _ := strconv.Atoi(ctx.PostForm("max_runtime"))
	maxMemory, _ := strconv.Atoi(ctx.PostForm("max_memory"))
	categoryIds := ctx.PostFormArray("category_ids")
	testCases := ctx.PostFormArray("test_case")
	if title == "" || content == "" {
		ctx.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "title和content不能为空",
		})
		return
	}

	identity := helper.GetUUID()
	data := &models.ProblemBasic{
		Identity:   identity,
		Title:      title,
		Content:    content,
		MaxRuntime: maxRuntime,
		MaxMemory:  maxMemory,
	}

	//处理分类
	var categoryBasic []*models.ProblemCategory
	for _, id := range categoryIds {
		categoryId, _ := strconv.Atoi(id)
		categoryBasic = append(categoryBasic, &models.ProblemCategory{
			ProblemID:  data.ID,
			CategoryID: uint(categoryId),
		})
	}
	data.ProblemCategories = categoryBasic

	//处理测试案例
	var testCasesTable []*models.TestCase
	for _, testCase := range testCases {
		//举个例子 {"input":"1 2\n", "output": "3\n"}
		caseMap := make(map[string]string)
		err := json.Unmarshal([]byte(testCase), &caseMap)
		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "测试案例格式错误" + err.Error(),
			})
			return
		}
		if _, ok := caseMap["input"]; !ok {
			ctx.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "测试案例格式错误，缺少input",
			})
			return
		}

		testCasesTable = append(testCasesTable, &models.TestCase{
			Identity:        helper.GetUUID(),
			ProblemIdentity: data.Identity,
			Input:           caseMap["input"],
			Output:          caseMap["output"],
		})
	}
	data.TestCases = testCasesTable

	err := models.DB.Create(&data).Error
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Create ProblemBasic Error" + err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": map[string]interface{}{
			"identity": identity,
			"list":     data,
		}})

}

// ProblemModify
// @Tags 管理员私有方法
// @Summary 问题修改
// @Param token header string true "token"
// @Param identity formData string true "identity"
// @Param title formData string true "title"
// @Param content formData string true " content"
// @Param category_ids formData []string false "category_ids" collectionFormat(multi)
// @Param test_cases formData []string false "test_cases" collectionFormat(multi)
// @Success 200 {string} json "{"code":200,"msg":"success","data":""}"
// @Router /admin/problem-modify [put]
func ProblemModify(ctx *gin.Context) {
	identity := ctx.PostForm("identity")
	title := ctx.PostForm("title")
	content := ctx.PostForm("content")
	categoryIds := ctx.PostFormArray("category_ids")
	testCases := ctx.PostFormArray("test_cases")

	if identity == "" || title == "" || content == "" {
		ctx.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "请求参数不能为空",
		})
		return
	}

	//问题基础信息保存
	data := &models.ProblemBasic{
		Identity: identity,
		Title:    title,
		Content:  content,
	}

	models.DB.Transaction(func(tx *gorm.DB) error {
		//修改问题表信息
		err := models.DB.Model(&models.ProblemBasic{}).Where("identity = ?", identity).Updates(&data).Error
		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "修改问题表失败 " + err.Error(),
			})
			return nil
		}

		//删除prolem-category表信息中关联的问题（删除关联）
		err = models.DB.Model(new(models.ProblemCategory)).Where("problem_id = ?", data.ID).Delete(&models.ProblemCategory{}).Error
		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "删除问题分类关联表失败 " + err.Error(),
			})
			return nil
		}

		//添加新的问题prolem-category表信息（添加新的关联）
		var problemCategorytTable []*models.ProblemCategory
		for _, categoryId := range categoryIds {
			categoryId, _ := strconv.Atoi(categoryId)
			categories := &models.ProblemCategory{
				ProblemID:  data.ID,
				CategoryID: uint(categoryId),
			}
			problemCategorytTable = append(problemCategorytTable, categories)
		}
		err = models.DB.Model(&models.ProblemCategory{}).Create(&problemCategorytTable).Error
		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "添加问题分类关联表失败 " + err.Error(),
			})
			return nil
		}

		//删除problem-testcase表信息中关联的问题
		err = models.DB.Model(new(models.TestCase)).Where("problem_id = ?", data.ID).Delete(&models.TestCase{}).Error
		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "删除问题测试用例关联表失败 " + err.Error(),
			})
			return nil
		}

		//添加新的问题problem-testcase表信息
		var TestcaseTable []*models.TestCase
		for _, testCase := range testCases {
			//举个例子 {"input":"1 2\n", "output": "3\n"}
			caseMap := make(map[string]string)
			err := json.Unmarshal([]byte(testCase), &caseMap)
			if err != nil {
				ctx.JSON(http.StatusOK, gin.H{
					"code": -1,
					"msg":  "测试案例格式错误" + err.Error(),
				})
				return nil
			}
			if _, ok := caseMap["input"]; !ok {
				ctx.JSON(http.StatusOK, gin.H{
					"code": -1,
					"msg":  "测试案例格式错误，缺少input",
				})
				return nil
			}

			testCase := &models.TestCase{
				Identity:        helper.GetUUID(),
				ProblemIdentity: data.Identity,
				Input:           caseMap["input"],
				Output:          caseMap["output"],
			}
			TestcaseTable = append(TestcaseTable, testCase)
		}
		err = models.DB.Model(&models.TestCase{}).Create(&TestcaseTable).Error
		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "添加问题测试用例关联表失败 " + err.Error(),
			})
			return nil
		}
		return nil
	})

	ctx.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "修改问题表成功",
	})
}
