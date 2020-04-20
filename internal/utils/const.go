package utils

const (
	// MinDataChunk is minimum valid data length to upload
	MinDataChunk = 5 * 1024 * 1024
	MaxDataLen   = MinDataChunk * 1024 * 4
)

var (
	CRLF = []byte{13, 10}
)
