package errs

import "errors"

// ErrInvalidTimestamp invalid timestamp
var ErrInvalidTimestamp = errors.New("invalid timestamp")

// ErrInvalidClientID invalid client id
var ErrInvalidClientID = errors.New("invalid client id")

// ErrInvalidContentID invalid content id
var ErrInvalidContentID = errors.New("invalid content id")

// ErrEmptyText empty text
var ErrEmptyText = errors.New("empty text")
