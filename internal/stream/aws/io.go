package aws

import (
	"bytes"
	"compress/gzip"
	"io"
	"log"
	"sync"

	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/krylphi/helloworld-data-handler/internal/domain"
)

type streamIO struct {
	buf    *bytes.Buffer
	gzw    *gzip.Writer
	queue  chan *domain.Entry
	closed bool
	mtx    chan int
	wg     *sync.WaitGroup
}

func newStreamIO() *streamIO {
	buf := &bytes.Buffer{}
	gzw := gzip.NewWriter(buf)
	return &streamIO{
		buf:    buf,
		gzw:    gzw,
		queue:  make(chan *domain.Entry, 10),
		closed: false,
		mtx:    make(chan int, 1),
	}
}

func (s *streamIO) Send(e *domain.Entry) {
	s.queue <- e
}

func (s *streamIO) write(p []byte) (n int, err error) {
	return s.gzw.Write(p)
}

func (s *streamIO) Read(p []byte) (n int, err error) {
	//return s.buf.Read(p)
	res, err := s.buf.Read(p)
	if s.closed && res == 0 && err == io.EOF {
		s.wg.Done()
		return 0, io.EOF
	}
	if res == 0 && err == io.EOF {
		return 0, nil
	}
	return res, nil
}

func (s *streamIO) Run(wg *sync.WaitGroup) {
	s.wg = wg
	s.wg.Add(1)
	var err error = nil
	go func() {
		for e := range s.queue {
			if _, err = s.write(e.Marshal()); err != nil {
				log.Print("failed to write to stream")
			}
		}
		if err = s.gzw.Flush(); err != nil {
			log.Print("failed to flush gzip stream")
		}
		if err = s.gzw.Close(); err != nil {
			log.Print("failed to flush gzip stream")
		}
		s.closed = true
	}()
}

func (s *streamIO) Flush() {
	go func() {
		close(s.queue)
	}()
}

type streamMap struct {
	lock    sync.RWMutex
	streams map[int]*uploadStream
}

type uploadStream struct {
	uploader *s3manager.Uploader
	reader   *streamIO
}

func (s *streamMap) get(key int) *uploadStream {
	s.lock.RLock()
	res := s.streams[key]
	s.lock.RUnlock()
	return res
}

func (s *streamMap) set(key int, stream *uploadStream) {
	s.lock.Lock()
	s.streams[key] = stream
	s.lock.Unlock()
}

func (s *streamMap) flush() {
	for _, v := range s.streams {
		log.Print("flushing stream")
		v.reader.Flush()
	}
}
