package util

import "math"

// Round a float to the specified precision
func Round(f float64, round int) (n float64) {
	shift := math.Pow(10, float64(round))
	return math.Floor((f*shift)+.5) / shift
}
