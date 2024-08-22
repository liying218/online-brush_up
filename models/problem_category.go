package models

import "gorm.io/gorm"

type ProblemCategory struct {
	gorm.Model
	ProblemID     uint           `gorm:"column:problem_id; type: uint" json:"problem_id"`
	CategoryID    uint           `gorm:"column:category_id; type: uint" json:"category_id"`
	CategoryBasic *CategoryBasic `gorm:"foreignKey:id; references:category_id"` //关联分类表的基础信息
}

func (table *ProblemCategory) TableName() string {
	return "problem_category"
}
