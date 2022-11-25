package reverse_test

import (
	"testing"

	"github.com/matryer/is"
	"github.com/russellslater/protohackers/cmd/line-reversal/reverse"
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

			got := reverse.Reverse(tc.input)

			is.Equal(got, tc.want)
		})
	}
}
