package main

import (
	"testing"

	"github.com/matryer/is"
)

func TestSessionAppendData(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name string
		data []string
		want []int
	}{
		{
			"Empty string",
			[]string{""},
			[]int{0},
		},
		{
			"Hello, world",
			[]string{"Hello,", " world!", "\n"},
			[]int{6, 13, 14},
		},
		{
			"Single new line character",
			[]string{"\n", "", "\n", "", "Foo", "Bar", "", "\n"},
			[]int{1, 1, 2, 2, 5, 8, 8, 9},
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			s := NewSession(0, nil)

			var got int
			for i, d := range tc.data {
				got = s.AppendData(d)
				is.Equal(got, tc.want[i])
			}

			is.Equal(s.ReceivedPos, got)
		})
	}
}
