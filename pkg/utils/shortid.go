package utils

const (
	alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
	base     = len(alphabet)
	length   = 10
)

func Encode(id uint64) string {
	var result [length]byte
	base64 := uint64(base)
	for i := 0; i < length; i++ {
		result[i] = alphabet[id%base64]
		id /= base64
	}
	return string(result[:])
}
