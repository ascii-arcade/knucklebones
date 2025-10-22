package utils

import (
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/ascii-arcade/knucklebones/language"
)

func GenerateName(lang *language.Language) string {
AGAIN:
	a := lang.UsernameFirstWords[rand.IntN(len(lang.UsernameFirstWords))]
	b := lang.UsernameSecondWords[rand.IntN(len(lang.UsernameSecondWords))]
	n := fmt.Sprintf("%s %s", a, b)
	if len(n) > 15 {
		goto AGAIN
	}

	return n
}

func GenerateDescriminator() string {
	return fmt.Sprintf("%04d", rand.IntN(10000))
}

func GenerateCode() string {
	for {
		b := make([]byte, 7)
		letters := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		for i := range b {
			if i == 3 {
				b[i] = '-'
				continue
			}
			b[i] = letters[rand.IntN(len(letters))]
		}
		code := string(b)
		return code
	}
}

func ToPointer[T any](v T) *T {
	return &v
}

func ValidPublicKey(key string) bool {
	validPrefixes := []string{"ssh-rsa", "ssh-ed25519", "ecdsa-sha2-nistp256", "ecdsa-sha2-nistp384", "ecdsa-sha2-nistp521"}
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(key, prefix) && len(key) > len(prefix)+20 {
			return true
		}
	}
	return false
}
