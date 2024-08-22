package test

import (
	uuid "github.com/satori/go.uuid"
	"testing"
)

func TestUUID(t *testing.T) {
	uuID := uuid.NewV4().String()
	t.Log(uuID)
}
