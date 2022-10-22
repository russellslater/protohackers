package boguscoin

import "regexp"

const (
	// it starts with a "7"
	// it consists of at least 26, and at most 35, alphanumeric characters
	// it starts at the start of a chat message, or is preceded by a space
	// it ends at the end of a chat message, or is followed by a space
	boguscoinAddr    = "7YWHMfk9JZe0LM0g1ZauHuiSxhI"
	boguscoinPattern = "(?:^|\\b)7[\\w]{25,34}(?:$|\\b)"
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
	return b.regex.ReplaceAllLiteralString(src, b.targetAddr)
}
