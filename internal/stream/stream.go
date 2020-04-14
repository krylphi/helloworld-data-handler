package stream

import "github.com/krylphi/helloworld-data-handler/internal/domain"

type Stream interface {
	Send(e *domain.Entry)
	Flush()
}
