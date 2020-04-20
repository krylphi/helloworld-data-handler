package stream

import (
	"log"
	"sync"

	"github.com/krylphi/helloworld-data-handler/internal/domain"
	"github.com/krylphi/helloworld-data-handler/internal/errs"
)

// UploadQueue queue incoming data for uploader
type UploadQueue struct {
	mx      sync.Mutex
	queue   chan *domain.Entry
	errChan chan error
	closed  bool
	wg      *sync.WaitGroup
	stream  Stream
}

// NewQueue UploadQueue constructor
func NewQueue(stream Stream, errChan chan error) Queue {
	return &UploadQueue{
		queue:   make(chan *domain.Entry, 10),
		errChan: errChan,
		closed:  false,
		stream:  stream,
	}
}

// Send send data to queue
func (q *UploadQueue) Send(e *domain.Entry) error {
	if q.IsClosed() {
		return errs.ErrQueueIsClosed
	}
	q.queue <- e
	return nil
}

// Run start queue execution
func (q *UploadQueue) Run(wg *sync.WaitGroup) {
	q.wg = wg
	q.wg.Add(1)
	var err error = nil
	go func() {
		for e := range q.queue {
			if err = q.stream.Write(e.Marshal()); err != nil {
				q.pushError(err)
				continue
			}
		}
		if err = q.stream.Flush(); err != nil {
			log.Print("failed to flush the stream")
		}
		q.wg.Done()
	}()
}

// Flush flush the data
func (q *UploadQueue) Flush() {
	q.mx.Lock()
	go func() {
		defer q.mx.Unlock()
		q.closed = true
		close(q.queue)
	}()
}

// IsClosed returns queue status
func (q *UploadQueue) IsClosed() bool {
	q.mx.Lock()
	defer q.mx.Unlock()
	return q.closed
}

func (q *UploadQueue) pushError(err error) {
	q.errChan <- err
}
