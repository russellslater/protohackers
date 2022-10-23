package boguscoin

import (
	"regexp"
	"strings"
)

const (
	// it starts with a "7"
	// it consists of at least 26, and at most 35, alphanumeric characters
	// it starts at the start of a chat message, or is preceded by a space
	// it ends at the end of a chat message, or is followed by a space
	boguscoinPattern = "(?:^|\b)7[\\w]{25,34}(?:$|\b)"
	boguscoinAddr    = "7YWHMfk9JZe0LM0g1ZauHuiSxhI"
)

type BoguscoinAddrRewriter struct {
	targetAddr string
	regex      *regexp.Regexp
}

func NewBoguscoinAddrRewriter() *BoguscoinAddrRewriter {
	return NewBoguscoinAddrRewriterWithAddr(boguscoinAddr)
}

func NewBoguscoinAddrRewriterWithAddr(targetAddr string) *BoguscoinAddrRewriter {
	return &BoguscoinAddrRewriter{
		targetAddr: targetAddr,
		regex:      regexp.MustCompile(boguscoinPattern),
	}
}

func (b *BoguscoinAddrRewriter) Rewrite(src string) string {
	spl := strings.Split(src, " ")
	for i, val := range spl {
		spl[i] = b.regex.ReplaceAllLiteralString(val, b.targetAddr)
	}
	return strings.Join(spl, " ")
}

func (b *BoguscoinAddrRewriter) RewriteBytes(src []byte) []byte {
	return []byte(b.Rewrite(string(src)))
}
