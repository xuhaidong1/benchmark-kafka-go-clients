package config

import "math/rand"

var Brokers = []string{"n-adx-kafka-1.adx.opera.com:1901", "n-adx-kafka-2.adx.opera.com:1902", "n-adx-kafka-3.adx.opera.com:1903"}
var Topic = "topic_adx_test"

// Characters Rune to use in the random string generator
var Characters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// GenMessage generates a random message of X bytes to use in the benchmarks.
func GenMessage(size int) []byte {
	b := make([]rune, size)
	for i := range b {
		b[i] = Characters[rand.Intn(len(Characters))]
	}
	return []byte(string(b))
}
