package errs

import "errors"

var ErrInvalidTimestamp = errors.New("invalid timestamp")
var ErrInvalidClientId = errors.New("invalid client id")
var ErrInvalidContentId = errors.New("invalid content id")
var ErrEmptyText = errors.New("empty text")
