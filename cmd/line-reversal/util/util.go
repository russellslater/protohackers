package util

import (
	"math"
	"strings"
)

func Reverse(arr []byte) []byte {
	if arr == nil {
		return nil
	}

	res := make([]byte, len(arr))
	copy(res, arr)

	for i := 0; i < len(res)/2; i++ {
		j := len(res) - i - 1
		res[i], res[j] = res[j], res[i]
	}

	return res
}

func replacements() [][]string {
	return [][]string{
		{`\/`, "/"},
		{`\\`, `\`},
		{`\n`, "\n"},
	}
}

func SlashUnescape(str string) string {
	res := str
	for _, r := range replacements() {
		res = strings.ReplaceAll(res, r[0], r[1])
	}
	return res
}

func SlashEscape(str string) string {
	res := str
	rep := replacements()
	for i := len(rep) - 1; i >= 0; i-- {
		res = strings.ReplaceAll(res, rep[i][1], rep[i][0])
	}
	return res
}

func Chunks(str string, maxSize int) []string {
	if len(str) == 0 || maxSize == 0 {
		return nil
	}

	chunks := make([]string, int(math.Ceil(float64(len(str))/float64(maxSize))))
	for i := 0; i < len(chunks)-1; i++ {
		idx := i * maxSize
		chunks[i] = str[idx : idx+maxSize]
	}
	chunks[len(chunks)-1] = str[(len(chunks)-1)*maxSize:]

	return chunks
}
