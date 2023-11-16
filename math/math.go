package math

func Max(series []int64, head int) int64 {
	var max int64
	for i := 1; i <= head; i++ {
		v := series[len(series) - i]
		if i == 1 || v > max {
			max = v
		}
	}
	return max
}

func Rate(series []int64, head int) int64 {
	l := len(series)
	return ( series[l - 1] - series[l - head] ) / int64(head - 1)
}
