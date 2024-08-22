package helper

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jordan-wright/email"
	uuid "github.com/satori/go.uuid"
	"math/big"
	"net/smtp"
	"os"
	"strings"
)

type UserClaims struct {
	jwt.RegisteredClaims
	Name     string `json:"name"`
	Identity string `json:"identity"`
	IsAdmin  int    `json:"is_admin"`
}

// GetMd5 获取md5
func GetMd5(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

var myKey = []byte("gin-gorm-key")

// GenerateToken 生成token
func GenerateToken(name, identity string, idAdmin int) (string, error) {
	userClaim := &UserClaims{
		Name:             name,
		Identity:         identity,
		IsAdmin:          idAdmin,
		RegisteredClaims: jwt.RegisteredClaims{},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, userClaim)
	tokenString, err := token.SignedString(myKey)
	if err != nil {
		return "", err
	}
	return tokenString, err
}

// 解析token
func ParseToken(tokenString string) (*UserClaims, error) {
	userClaims := new(UserClaims)
	claims, err := jwt.ParseWithClaims(tokenString, userClaims, func(token *jwt.Token) (interface{}, error) {
		return myKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !claims.Valid {
		return nil, fmt.Errorf("token is invalid")
	}
	return userClaims, nil
}

// SendEmailCode
func SendCode(toUserEmail, code string) error {
	e := email.NewEmail()
	e.From = "<1161546480@qq.com>"
	e.To = []string{toUserEmail}
	e.Subject = "发送邮箱验证码"
	e.HTML = []byte("您的验证码：<b>" + code + "</b>")
	//返回EOF时，关闭SSL重试
	return e.SendWithTLS("smtp.qq.com:465",
		smtp.PlainAuth("", "1161546480@qq.com", "vueidyskurpihhhi", "smtp.qq.com"),
		&tls.Config{InsecureSkipVerify: true, ServerName: "smtp.qq.com"})

}

// GetUUID
func GetUUID() string {
	return uuid.NewV4().String()
}

// 生成验证码
func GenerateRandCode() string {
	const charset = "0123456789" // 只包含数字
	var sb strings.Builder
	for i := 0; i < 6; i++ {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			panic(err)
		}
		sb.WriteByte(charset[randomIndex.Int64()])
	}
	return sb.String()
}

func CodeSave(code []byte) (string, error) {
	dirName := "code-user/" + GetUUID()
	path := dirName + "/main.go"
	err := os.Mkdir(dirName, 0777)
	if err != nil {
		return "", err
	}
	file, err := os.Create(path)
	if err != nil {
		return "", err
	}
	file.Write(code)
	defer file.Close()
	return path, nil
}
