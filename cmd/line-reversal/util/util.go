package util

import (
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

func SlashUnescape(str string) string {
	res := strings.ReplaceAll(str, `\/`, "/")
	res = strings.ReplaceAll(res, `\\`, `\`)
	return res
}
