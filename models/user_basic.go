package models

import "gorm.io/gorm"

type UserBasic struct {
	gorm.Model
	Identity  string `gorm:"column:identity; type:varchar(36);" json:"identity"` //用户的唯一标识
	Name      string `gorm:"column:name; type:varchar(100);" json:"name"`
	Email     string `gorm:"column:email; type:varchar(100);" json:"email"`
	Phone     string `gorm:"column:phone; type:varchar(20);" json:"phone"`
	Password  string `gorm:"column:password; type:varchar(32);" json:"password"`
	PassNum   int64  `gorm:"column:pass_num; type:int(11);" json:"pass_num"`     //完成问题个数
	SubmitNum int64  `gorm:"column:submit_num; type:int(11);" json:"submit_num"` //提交次数
	IsAdmin   int    `gorm:"column:is_admin; type:int(1);" json:"is_admin"`      //是否为管理员 0:不是 1:是
}

// TableName sets the insert table name for this struct type
func (table *UserBasic) TableName() string {
	return "user_basic"
}
