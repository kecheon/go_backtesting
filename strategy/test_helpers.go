package strategy

import "math"

// CloseEnough compares two floats for near-equality.
func CloseEnough(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}
