package models

import (
	"gorm.io/gorm"
)

type SubmitBasic struct {
	Identity        string        `gorm:"column:identity; type:varchar(36);" json:"identity"` //提交表唯一标识
	ProblemIdentity string        `gorm:"column:problem_identity; type:varchar(36);" json:"problem_identity"`
	ProblemBasic    *ProblemBasic `gorm:"foreignKey:identity; references:problem_identity" ` //关联问题基础表
	UserBasic       *UserBasic    `gorm:"foreignKey:identity; references:user_identity" `    //关联用户基础表
	UserIdentity    string        `gorm:"column:user_identity; type:varchar(36);" json:"user_identity"`
	Path            string        `gorm:"column:path; type:varchar(255);" json:"path"`
	Status          int           `gorm:"column:status; type:int;" json:"status"` //0：待判断 1：答案正确 2：答案错误 3：运行超时 4：运行超内存
}

func (table *SubmitBasic) TableName() string {
	return "submit_basic"
}

func GetSubmitList(userIdentity, problemIdentity string, status int) *gorm.DB {
	tx := DB.Model(new(SubmitBasic)).Preload("ProblemBasic", func(db *gorm.DB) *gorm.DB {
		return db.Omit("content")
	}).Preload("UserBasic")
	if userIdentity != "" {
		tx.Where("submit_basic.user_identity = ?", userIdentity)
	}
	if problemIdentity != "" {
		tx.Where("submit_basic.problem_identity = ?", problemIdentity)
	}
	if status != -1 {
		tx.Where("submit_basic.status = ?", status)
	}
	return tx
}
