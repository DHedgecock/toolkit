package toolkit

import "crypto/rand"

const randomStringSource = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_+"

// Tools is the type used to instantiate this module
type Tools struct{}

// RandomString returns a string of randam characters of length n
func (t *Tools) RandomString(n int) string {
	randomString := make([]rune, n)
	stringSource := []rune(randomStringSource)

	for idx := range randomString {
		primeNumber, _ := rand.Prime(rand.Reader, len(stringSource))
		x := primeNumber.Uint64()
		y := uint64(len(stringSource))

		randomString[idx] = stringSource[x%y]
	}

	return string(randomString)
}
