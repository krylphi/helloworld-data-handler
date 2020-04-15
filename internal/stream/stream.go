package stream

import "github.com/krylphi/helloworld-data-handler/internal/domain"

// Stream stream interface
type Stream interface {
	Send(e *domain.Entry)
	Flush()
}
