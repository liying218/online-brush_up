package test

import (
	"fmt"
	"gin_gorm/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
)

func TestGormTest(t *testing.T) {
	dsn := "root:123456@tcp(127.0.0.1:3306)/gin_gorm_oj?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	data := make([]*models.ProblemBasic, 0)
	err = db.Find(&data).Error
	if err != nil {
		t.Fatal(err)
	}

	for _, v := range data {
		fmt.Println(v.ID)
		fmt.Println(v.Title)
		fmt.Println(v.Content)
	}
}
