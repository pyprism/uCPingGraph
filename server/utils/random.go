package utils

import (
	"math"
	"math/rand"
)

func RandomFloat() float64 {
	// Generate a random float number between 0.0 and 1.0
	randomFloat := rand.Float64()

	// Scale and shift the random number to the desired range
	scaledRandomFloat := 2.0 + randomFloat*(20.0-2.0)

	// Round the scaled random float number to two decimal places
	roundedRandomFloat := math.Round(scaledRandomFloat*100) / 100
	return roundedRandomFloat
}

func RandomInt() int {
	randomInt := rand.Intn(10)
	return randomInt
}
