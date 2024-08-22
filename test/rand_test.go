package test

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"testing"
)

func TestRand(t *testing.T) {
	const charset = "0123456789" // 只包含数字
	var sb strings.Builder
	for i := 0; i < 6; i++ {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			panic(err)
		}
		sb.WriteByte(charset[randomIndex.Int64()])
	}
	fmt.Println(sb.String())
}
