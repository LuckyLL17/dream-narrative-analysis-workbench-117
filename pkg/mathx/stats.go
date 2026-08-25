package mathx

import "sort"

func Average(
	values []float64,
) float64 {
	if len(values) == 0 {
		return 0
	}
	var total float64
	for i := range values {
		value := values[i]
		total += value
	}
	return total / float64(len(values))
}

func Ratio(
	part,
	total int,
) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) / float64(total)
}

func Round(
	value float64,
	digits int,
) float64 {
	if digits < 0 {
		return value
	}
	p := 1.0
	for index := 0; index < digits; index++ {
		p *= 10
	}
	return float64(int(value*p+0.5)) / p
}

func Correlation(
	xs,
	ys []float64,
) float64 {
	if len(xs) != len(ys) || len(xs) < 2 {
		return 0
	}
	var sx, sy float64
	for i := range xs {
		sx += xs[i]
		sy += ys[i]
	}
	meanX, meanY := sx/float64(len(xs)), sy/float64(len(ys))
	var numerator, xx, yy float64
	for i := range xs {
		dx, dy := xs[i]-meanX, ys[i]-meanY
		numerator += dx * dy
		xx += dx * dx
		yy += dy * dy
	}
	if xx == 0 || yy == 0 {
		return 0
	}
	return numerator / (sqrt(xx) * sqrt(yy))
}

func sqrt(
	value float64,
) float64 {
	if value <= 0 {
		return 0
	}
	guess := value
	for i := 0; i < 12; i++ {
		guess = (guess + value/guess) / 2
	}
	return guess
}

func SortedKeys(
	values map[string]int,
) []string {
	keys := make(
		[]string,
		0,
		len(values),
	)
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
