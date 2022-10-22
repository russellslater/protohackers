package boguscoin

import (
	"github.com/dlclark/regexp2"
)

const (
	// it starts with a "7"
	// it consists of at least 26, and at most 35, alphanumeric characters
	// it starts at the start of a chat message, or is preceded by a space
	// it ends at the end of a chat message, or is followed by a space
	boguscoinPattern = "(?:^|(?<= ))7[\\w]{25,34}(?:$|(?= ))"
	boguscoinAddr    = "7YWHMfk9JZe0LM0g1ZauHuiSxhI"
)

type BoguscoinAddrRewriter struct {
	targetAddr string
	regex      *regexp2.Regexp
}

func NewBoguscoinAddrRewriter() *BoguscoinAddrRewriter {
	return NewBoguscoinAddrRewriterWithAddr(boguscoinAddr)
}

func NewBoguscoinAddrRewriterWithAddr(targetAddr string) *BoguscoinAddrRewriter {
	return &BoguscoinAddrRewriter{
		targetAddr: targetAddr,
		regex:      regexp2.MustCompile(boguscoinPattern, 0),
	}
}

func (b *BoguscoinAddrRewriter) Rewrite(src string) string {
	str, _ := b.regex.Replace(src, b.targetAddr, -1, -1)
	return str
}

func (b *BoguscoinAddrRewriter) RewriteBytes(src []byte) []byte {
	return []byte(b.Rewrite(string(src)))
}
