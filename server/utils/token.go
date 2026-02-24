package utils

import (
	"crypto/rand"
	"math/big"
)

// GenToken generates a cryptographically secure random string based on the size provided.
func GenToken(size int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	token := make([]byte, size)
	maxIdx := big.NewInt(int64(len(letters)))
	for i := range token {
		n, err := rand.Int(rand.Reader, maxIdx)
		if err != nil {
			panic("crypto/rand failed: " + err.Error())
		}
		token[i] = letters[n.Int64()]
	}
	return string(token)

}
