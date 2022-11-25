package reverse

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
