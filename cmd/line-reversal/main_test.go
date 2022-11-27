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

func TestSessionCompletedLinesOutOfBounds(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	s := NewSession(0, nil)

	gotStrs, gotLen := s.CompletedLines(100) // out of bounds

	is.Equal(gotStrs, nil)
	is.Equal(gotLen, 0)
}

func TestSessionCompletedLines(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name     string
		data     string
		pos      int
		wantStrs []string
		wantLen  int
	}{
		{
			"Incomplete line",
			"Hello, world!",
			0,
			nil,
			0,
		},
		{
			"Single line",
			"Hello, world!\n",
			0,
			[]string{"Hello, world!"},
			14,
		},
		{
			"Two lines",
			"Hello, world!\nHere's another line.\n",
			0,
			[]string{"Hello, world!", "Here's another line."},
			35,
		},
		{
			"Empty string",
			"",
			0,
			nil,
			0,
		},
		{
			"Start part way through second line",
			"Apple\nOrange\nPear\n",
			8,
			[]string{"ange", "Pear"},
			10,
		},
		{
			"Start part way through third line",
			"Apple\nOrange\nPear\n",
			15,
			[]string{"ar"},
			3,
		},
		{
			"Four empty lines",
			"\n\n\n\n",
			0,
			[]string{"", "", "", ""},
			4,
		},
		{
			"Last two of four empty lines",
			"\n\n\n\n",
			2,
			[]string{"", ""},
			2,
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			s := NewSession(0, nil)
			s.AppendData(tc.data)

			gotStrs, gotLen := s.CompletedLines(tc.pos)

			is.Equal(gotStrs, tc.wantStrs)
			is.Equal(gotLen, tc.wantLen)
		})
	}
}
