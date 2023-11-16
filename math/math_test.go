package math_test

import (
	"testing"
	"strconv"
	"netexp/math"
)

func TestMax (t *testing.T) {
	tests := []struct{
		series []int64
		head int
		want int64
	}{
		{ []int64{ 1, 2, 3, 4, 5, 6, 7 }, 3, 7 },
		{ []int64{ 1, 2, 3, 9, 3, 2, 1 }, 2, 2 },
		{ []int64{ 1, 2, 3, 9, 3, 2, 1 }, 4, 9 },
		{ []int64{ 2, 1, 7, 8, 6, 1, 2 }, 5, 8 },
	}

	for i, tc := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := math.Max(tc.series, tc.head)
			if got != tc.want {
				t.Errorf("incorrect max; got %d want %d", got, tc.want)
			}
		})
	}
}

func TestRate (t *testing.T) {
	tests := []struct{
		series []int64
		head int
		want int64
	}{
		{ []int64{ 1, 9, 1, 9, 5, 6, 7 }, 3,  1 },
		{ []int64{ 1, 2, 3, 7, 3, 2, 1 }, 4, -2 },
		{ []int64{ 1, 2, 3, 9, 3, 2, 1 }, 7,  0 },
		{ []int64{ 2, 1, 6, 8, 6, 1, 9 }, 5,  0 },
	}

	for i, tc := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := math.Rate(tc.series, tc.head)
			if got != tc.want {
				t.Errorf("incorrect rate; got %d want %d", got, tc.want)
			}
		})
	}
}
