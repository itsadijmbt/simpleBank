package util

import (
	"math/rand"
	"strings"
	"time"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

func init() {

	// runs every time for once
	// without see the gebeate values will be random but same as always

	rand.Seed(time.Now().UnixNano())

}

func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

func RandomString(n int) string {

	var sb strings.Builder
	k := len(alphabet)

	for i := 0; i < n; i++ {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}
	return sb.String()

}

// random strng of six letters

func RandomName(n int) string {

	var name strings.Builder

	k := len(alphabet)

	for i := 0; i < n; i++ {
		c := alphabet[rand.Intn(k)]
		name.WriteByte(c)
	}

	return name.String()

}

func RandomMoney() int64 {

	return RandomInt(0, 10000000000)
}

func RandomCurrency() string {
	var curr strings.Builder
	curry := []string{"USD", "EUR", "INR"}

	c := curry[rand.Int31n(3)]
	curr.WriteString(c)

	return curr.String()

}
