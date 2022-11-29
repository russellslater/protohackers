package util_test

import (
	"testing"

	"github.com/matryer/is"
	"github.com/russellslater/protohackers/cmd/line-reversal/util"
)

func TestReverseLine(t *testing.T) {
	t.Parallel()
	tt := []struct {
		name  string
		input []byte
		want  []byte
	}{
		{
			"Nil array",
			nil,
			nil,
		},
		{
			"Single character",
			[]byte("A"),
			[]byte("A"),
		},
		{
			"12345",
			[]byte("12345"),
			[]byte("54321"),
		},
		{
			"Hello, world",
			[]byte("Hello, world!"),
			[]byte("!dlrow ,olleH"),
		},
		{
			"0x00, 0x14, 0x69",
			[]byte{0x00, 0x14, 0x69},
			[]byte{0x69, 0x14, 0x00},
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			got := util.Reverse(tc.input)

			is.Equal(got, tc.want)
		})
	}
}

func TestSlashUnescape(t *testing.T) {
	t.Parallel()
	tt := []struct {
		name  string
		input string
		want  string
	}{
		{
			"Forward slash and backslash",
			`foo\/bar\\baz`,
			`foo/bar\baz`,
		},
		{
			"Double slashes",
			`foo\\\/bar\\\\baz`,
			`foo\/bar\\baz`,
		},
		{
			"New lines",
			`Hello\\nworld\\n`,
			"Hello\nworld\n",
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			got := util.SlashUnescape(tc.input)

			is.Equal(got, tc.want)
		})
	}
}

func TestSlashEscape(t *testing.T) {
	t.Parallel()
	tt := []struct {
		name  string
		input string
		want  string
	}{
		{
			"Forward slash and backslash",
			`foo/bar\baz`,
			`foo\/bar\\baz`,
		},
		{
			"Double slashes",
			`foo\/bar\\baz`,
			`foo\\\/bar\\\\baz`,
		},
		{
			"New lines",
			"Hello\nworld\n",
			`Hello\\nworld\\n`,
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			got := util.SlashEscape(tc.input)

			is.Equal(got, tc.want)
		})
	}
}
