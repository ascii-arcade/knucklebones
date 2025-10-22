package utils

import "math/rand/v2"

func Color() string {
	return "#" + randomHex(6)
}
func randomHex(length int) string {
	const charset = "0123456789abcdef"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.IntN(len(charset))]
	}
	return string(result)
}
