package boguscoin_test

import (
	"testing"

	"github.com/matryer/is"
	"github.com/russellslater/protohackers/cmd/mob-in-the-middle/boguscoin"
)

// Hi alice, please send payment to 7iKDZEwPZSqIvDnHvVN2r0hUWXD5rHX

func TestContainsBoguscoinAddr(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name       string
		str        string
		targetAddr string
		want       string
	}{
		{
			name:       "Empty string",
			str:        "",
			targetAddr: "7YWHMfk9JZe0LM0g1ZauHuiSxhI",
			want:       "",
		},
		{
			name:       "String without Boguscoin remains unchanged",
			str:        "A simple string without a Boguscoin address",
			targetAddr: "7YWHMfk9JZe0LM0g1ZauHuiSxhI",
			want:       "A simple string without a Boguscoin address",
		},
		{
			name:       "25-character alphanumeric string starting with a 7 remains unchanged",
			str:        "7albw8et5kg078g7dq3x7bo4i",
			targetAddr: "7YWHMfk9JZe0LM0g1ZauHuiSxhI",
			want:       "7albw8et5kg078g7dq3x7bo4i",
		},
		{
			name:       "36-character alphanumeric string starting with a 7 remains unchanged",
			str:        "7ap1ikiow520w65diywf8dsjh89dwrfuvsyh",
			targetAddr: "7YWHMfk9JZe0LM0g1ZauHuiSxhI",
			want:       "7ap1ikiow520w65diywf8dsjh89dwrfuvsyh",
		},
		{
			name:       "Valid Boguscoin address is replaced",
			str:        "76kit7xpcxtkrh87z4hobkx05oni3r",
			targetAddr: "7YWHMfk9JZe0LM0g1ZauHuiSxhI",
			want:       "7YWHMfk9JZe0LM0g1ZauHuiSxhI",
		},
		{
			name:       "Valid Boguscoin address at end of string is replaced",
			str:        "Hi alice, please send payment to 7iKDZEwPZSqIvDnHvVN2r0hUWXD5rHX",
			targetAddr: "7YWHMfk9JZe0LM0g1ZauHuiSxhI",
			want:       "Hi alice, please send payment to 7YWHMfk9JZe0LM0g1ZauHuiSxhI",
		},
		{
			name:       "Valid Boguscoin address at start of string is replaced",
			str:        "70wo1xny3yjqjxik2a75sptgdbl46qob is my Boguscoin address",
			targetAddr: "7YWHMfk9JZe0LM0g1ZauHuiSxhI",
			want:       "7YWHMfk9JZe0LM0g1ZauHuiSxhI is my Boguscoin address",
		},
		{
			name:       "Valid Boguscoin address in middle of string is replaced",
			str:        "Would you be able to send 3 Boguscoins to 7quurs6ex9m64avsqt1gfc19bpga before tomorrow?",
			targetAddr: "7YWHMfk9JZe0LM0g1ZauHuiSxhI",
			want:       "Would you be able to send 3 Boguscoins to 7YWHMfk9JZe0LM0g1ZauHuiSxhI before tomorrow?",
		},
		{
			name:       "Multiple valid Boguscoin addresses replaced",
			str:        "k56w5f5uif2ww13ai22xmpcfc7or3z 7484y99h7u1jhwnt4fz82fosaf 7a0z01cul8ksshgvy4o8lj0sicn61 9mobune4wuaqgotihpsmxoqnnnb88f",
			targetAddr: "7YWHMfk9JZe0LM0g1ZauHuiSxhI",
			want:       "k56w5f5uif2ww13ai22xmpcfc7or3z 7YWHMfk9JZe0LM0g1ZauHuiSxhI 7YWHMfk9JZe0LM0g1ZauHuiSxhI 9mobune4wuaqgotihpsmxoqnnnb88f",
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			b := boguscoin.NewBoguscoinAddrRewriterWithAddr(tc.targetAddr)
			got := b.Rewrite(tc.str)
			is.Equal(got, tc.want) // rewritten string is not a match
		})
	}
}
