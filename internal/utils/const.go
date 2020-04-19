package utils

const (
	// MinAWSChunk is minimum valid data length to upload for AWS. Only the last (or the only) chunk can be any length
	MinAWSChunk = 5 * 1024 * 1024
)
