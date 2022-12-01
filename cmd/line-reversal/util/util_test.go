package util_test

import (
	"reflect"
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

func TestChunks(t *testing.T) {
	t.Parallel()
	tt := []struct {
		name    string
		input   string
		maxSize int
		want    []string
	}{
		{
			"Empty string",
			"",
			0,
			nil,
		},
		{
			"Max size of zero",
			"Hello, world!",
			0,
			nil,
		},
		{
			"Hello, world",
			"Hello, world!",
			100,
			[]string{"Hello, world!"},
		},
		{
			"Split in two",
			"Hello, world!",
			7,
			[]string{"Hello, ", "world!"},
		},
		{
			"Split in three",
			`lcvpWuQrNckK7rMaeGz9BcLRZNH4agyoHdFIJAsez0LNNb9SU6rxYnQnPwbW8uXpXPhp1vtolgDBr8vpz3iXQ2g0lbDYmwLQv7dd\\n`,
			50,
			[]string{"lcvpWuQrNckK7rMaeGz9BcLRZNH4agyoHdFIJAsez0LNNb9SU6", "rxYnQnPwbW8uXpXPhp1vtolgDBr8vpz3iXQ2g0lbDYmwLQv7dd", `\\n`},
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			got := util.Chunks(tc.input, tc.maxSize)
			is.True(reflect.DeepEqual(got, tc.want)) // arrays do not match
		})
	}
}

func BenchmarkChunks(b *testing.B) {
	input := `lcvpWuQrNckK7rMaeGz9BcLRZNH4agyoHdFIJAsez0LNNb9SU6rxYnQnPwbW8uXpXPhp1vtolgDBr8vpz3iXQ2g0lbDYmwLQv7dd\\n`
	maxSize := 50

	for i := 0; i < b.N; i++ {
		util.Chunks(input, maxSize)
	}
}
