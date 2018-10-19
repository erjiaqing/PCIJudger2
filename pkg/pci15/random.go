package pci15

import (
	"crypto/rand"
	"fmt"
)

func GetRandomString() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
