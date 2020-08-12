package util

import (
	"math/rand"
	"time"
)

// RandomString returns a random string of length n.
func RandomString(n int) string {
	rand.Seed(time.Now().UnixNano())
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// Turns a string into a type that implements the Stringer interface. Example: util.Stringer("my string").
type Stringer string

func (s Stringer) String() string { return string(s) }
