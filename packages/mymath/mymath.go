package mymath

func Exponentiate(x float64, n int64) float64 {

	Is_Neg := false

	if n == 0 {
		return 1
	} else if n < 0 {
		Is_Neg = true
		n -= 2 * n
	}

	result := 1.0
	for n > 0 {
		if n%2 == 1 {
			result *= x
		}
		x *= x
		n /= 2
	}

	if Is_Neg {
		result = 1 / result
	}

	return result
}

func SquareRoot(x float64) float64 {
	var bottom float64 = 0.0
	var top float64 = x
	var middle float64 = (x + bottom) / 2
	var bin_search float64
	var best float64 = x
	var Exp_Best = best
	var goal float64 = x

	for (top - bottom) > 1e-9 {
		bin_search = Exponentiate(middle, 2)
		if abs(bin_search-goal) < abs(Exp_Best-goal) {
			best = middle
			Exp_Best = bin_search
		}

		if bin_search > goal {
			top = middle
		} else {
			bottom = middle
		}
		middle = (top + bottom) / 2
	}
	return best
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
