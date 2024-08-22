package models

import (
	"gorm.io/gorm"
)

type ProblemBasic struct {
	gorm.Model
	Identity          string             `gorm:"column:identity; type:varchar(36);" json:"identity"` //问题的唯一标识
	ProblemCategories []*ProblemCategory `gorm:"foreignKey:problem_id;references:id"`
	Title             string             `gorm:"column:title; type:varchar(255);" json:"title"`    //问题标题
	Content           string             `gorm:"column:content; type:text;" json:"content"`        //问题正文
	MaxMemory         int                `gorm:"column:max_memory; type:int;" json:"max_memory"`   //最大内存限制
	MaxRuntime        int                `gorm:"column:max_runtime; type:int;" json:"max_runtime"` //最大运行时间限制
	TestCases         []*TestCase        `gorm:"foreignKey:problem_identity;references:identity"`
	PassNum           int                `gorm:"column:pass_num; type:int;" json:"pass_num"`     //通过问题次数
	SubmitNum         int                `gorm:"column:submit_num; type:int;" json:"submit_num"` //提交问题次数
}

func (table *ProblemBasic) TableName() string {
	return "problem_basic"
}

func GetProblemList(keyword, categoryIdentity string) *gorm.DB {
	tx := DB.Model(new(ProblemBasic)).Preload("ProblemCategories").Preload("ProblemCategories.CategoryBasic").
		Where("title like ? OR content like ?", "%"+keyword+"%", "%"+keyword+"%")
	if categoryIdentity != "" {
		tx.Joins("RIGHT JOIN problem_category pc ON pc.problem_id = problem_basic.id").
			Where("pc.category_id =(SELECT cb.id FROM category_basic cb WHERE cb.identity = ? )", categoryIdentity)
	}
	return tx
}
