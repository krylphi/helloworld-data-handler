//go:generate mockgen -source=stream.go -destination=../mock/stream.go -package=mock
package stream

import (
	"github.com/krylphi/helloworld-data-handler/internal/domain"
	"sync"
)

// Handler handler
type Handler interface {
	Send(e *domain.Entry) error
	Flush()
}

// Queue writing queue
type Queue interface {
	Send(e *domain.Entry) error
	Run(wg *sync.WaitGroup)
	IsClosed() bool
	Flush()
}

// Stream stream interface
type Stream interface {
	Write(data []byte) error
	Flush() (err error)
}
