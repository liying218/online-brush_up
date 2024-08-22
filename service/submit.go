package service

import (
	"bufio"
	"bytes"
	"errors"
	"gin_gorm/define"
	"gin_gorm/helper"
	"gin_gorm/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"io"
	"log"
	"os/exec"
	"runtime"
	"strconv"
	"sync"
	"time"
)

// GetSubmitList 获取题目列表
// @Tags 公共方法
// @Summary 提交表详情
// @Param user_identity query string false "user_identity"
// @Param problem_identity query string false "problem_identity"
// @Param status query int false "status"
// @Success 200 {string} json "{"code":200,"msg":"success","data":""}"
// @Router /submit-list [get]
func GetSubmitList(ctx *gin.Context) {
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

	userIdentity := ctx.Query("user_identity")
	problemIdentity := ctx.Query("problem_identity")
	status, _ := strconv.Atoi(ctx.Query("status"))

	var count int64
	var submitTable []*models.SubmitBasic

	err = models.GetSubmitList(userIdentity, problemIdentity, status).Count(&count).Offset(offset).Limit(limit).Find(&submitTable).Error
	if err != nil {
		ctx.JSON(200, gin.H{
			"code": -1,
			"msg":  "获取提交问题表失败",
		})
	}

	ctx.JSON(200, gin.H{
		"code": 200,
		"msg":  "success",
		"data": map[string]interface{}{
			"list":  submitTable,
			"count": count,
		},
	})

}

// SubmitCode
// @Tags 用户私有方法
// @Summary 提交代码
// @Param token header string true "token"
// @Param problem_identity formData string true "problem_identity"
// @Param code body string true "code"
// @Success 200 {string} json "{"code":200,"msg":"success","data":""}"
// @Router /admin/submit-code [post]
func SubmitCode(ctx *gin.Context) {
	problemIdentity := ctx.PostForm("problem_identity")
	code, err := io.ReadAll(bufio.NewReader(ctx.Request.Body))
	if err != nil {
		ctx.JSON(200, gin.H{
			"code": -1,
			"msg":  "读取代码失败" + err.Error(),
		})
		return
	}
	//代码保存
	path, err := helper.CodeSave(code)
	if err != nil {
		ctx.JSON(200, gin.H{
			"code": -1,
			"msg":  "代码保存失败",
		})
		return
	}

	u, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(200, gin.H{
			"code": -1,
			"msg":  "用户不存在",
		})
	}
	userClaims := u.(*helper.UserClaims)

	submitCode := &models.SubmitBasic{
		Identity:        helper.GetUUID(),
		ProblemIdentity: problemIdentity,
		UserIdentity:    userClaims.Identity,
		Path:            path,
	}

	//代码判断
	problemBasic := new(models.ProblemBasic)
	err = models.DB.Where("identity = ?", problemIdentity).Preload("TestCases").First(&problemBasic).Error
	if err != nil {
		ctx.JSON(200, gin.H{
			"code": -1,
			"msg":  "获取问题信息失败" + err.Error(),
		})
		return
	}
	WA := make(chan int)  //答案错误的channel
	OOM := make(chan int) //答案超内存的channel
	CE := make(chan int)  //编译错误的Channel
	passCount := 0        //通过的个数
	var lock sync.Mutex
	var msg string //提示信息

	for _, testCase := range problemBasic.TestCases {
		testCase := testCase
		go func() {
			//执行测试
			cmd := exec.Command("go", "run", path)
			var out, stderr bytes.Buffer
			cmd.Stderr = &stderr
			cmd.Stdout = &out
			stdinPipe, err := cmd.StdinPipe()
			if err != nil {
				log.Println("stdinPipe err", err.Error())
				return
			}
			io.WriteString(stdinPipe, testCase.Input)

			var bm runtime.MemStats
			runtime.ReadMemStats(&bm)

			err = cmd.Run()
			if err != nil {
				log.Println("cmd.Run err", err.Error())
				if err.Error() == "exit status 2" {
					msg = stderr.String()
					CE <- 1
					return
				}
			}
			var em runtime.MemStats
			runtime.ReadMemStats(&em)
			if em.Alloc-bm.Alloc > 100000000 {
				OOM <- 1
				return
			}
			//答案错误
			if testCase.Output != out.String() {
				msg = "答案错误"
				WA <- 1
				return
			}
			//运行超内存
			if (em.Alloc/1024)-(bm.Alloc/1024) > uint64(problemBasic.MaxRuntime) {
				msg = "运行超内存"
				OOM <- 1
				return
			}
			//运行通过
			lock.Lock()
			passCount++
			lock.Unlock()
		}()
	}
	//判断是否通过
	select {
	//0：待判断 1：答案正确 2：答案错误 3：运行超时 4：运行超内存 5:编译错误
	case <-WA:
		submitCode.Status = 2
	case <-OOM:
		submitCode.Status = 4
	case <-CE:
		submitCode.Status = 5
	case <-time.After(time.Millisecond * time.Duration(problemBasic.MaxRuntime)):
		if passCount == len(problemBasic.TestCases) {
			submitCode.Status = 1
		} else {
			submitCode.Status = 3
			msg = "运行超时"
		}
	}

	err = models.DB.Transaction(func(tx *gorm.DB) error {
		err = tx.Create(&submitCode).Error
		if err != nil {
			return errors.New("创建提交记录失败" + err.Error())
		}
		m := make(map[string]interface{})
		m["submit_count"] = gorm.Expr("submit_count + ?", 1)
		if submitCode.Status == 1 {
			m["pass_num"] = gorm.Expr("pass_num + ?", 1)
		}
		//更新user-basic
		err := tx.Model(new(models.UserBasic)).Where("identity = ?", userClaims.Identity).Updates(m).Error
		if err != nil {
			return errors.New("更新用户提交记录失败" + err.Error())
		}
		//更新problem-basic
		err = tx.Model(new(models.ProblemBasic)).Where("identity = ?", problemIdentity).Updates(m).Error
		if err != nil {
			return errors.New("更新问题提交记录失败" + err.Error())
		}
		return nil
	})
	if err != nil {
		ctx.JSON(200, gin.H{
			"code": -1,
			"msg":  "Transaction failed" + err.Error(),
		})

		ctx.JSON(200, gin.H{
			"code": 200,
			"data": map[string]interface{}{
				"msg":    msg,
				"status": submitCode.Status,
			},
		})

	}
}
