package utils

import "math/rand"

// GenToken generates a random string based on the size provided
func GenToken(size int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	var token = make([]rune, size)
	for i := range token {
		token[i] = letters[rand.Intn(len(letters))]
	}
	return string(token)

}
