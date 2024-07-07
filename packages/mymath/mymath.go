package mymath

import "math"

func Exponentiate(x, y int64) float64 {

	//TODO implement my own exponentiation
	//It's not that bad, most libs use e^yln(x) as the identity

	return math.Pow(float64(x), float64(y))
}
