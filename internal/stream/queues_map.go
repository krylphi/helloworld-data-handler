package stream

import (
	"log"
	"sync"
	"time"
)

// QueuesMap is a map for active queues
type QueuesMap struct {
	lock         sync.RWMutex
	queueTimeout time.Duration
	streams      map[int]Queue
}

// NewQueuesMap QueuesMap constructor
func NewQueuesMap(queueTimeout time.Duration) *QueuesMap {
	return &QueuesMap{
		lock:         sync.RWMutex{},
		queueTimeout: queueTimeout,
		streams:      make(map[int]Queue, 10),
	}
}

// GetQueue get queue by id
func (s *QueuesMap) GetQueue(id int) Queue {
	s.lock.RLock()
	res := s.streams[id]
	if res != nil && res.IsClosed() {
		s.lock.Lock()
		s.streams[id] = nil
		delete(s.streams, id)
		s.lock.Unlock()
		s.lock.RUnlock()
		return nil
	}
	s.lock.RUnlock()
	return res
}

// AddQueue add new queue to map
func (s *QueuesMap) AddQueue(id int, stream Stream, wg *sync.WaitGroup) Queue {
	errorChan := make(chan error)
	go func() {
		finalize := func() {
			close(errorChan)
			s.RemoveQueue(id)
		}
		select {
		case <-errorChan:
			finalize()
		case <-time.After(s.queueTimeout):
			finalize()
		}
	}()
	s.lock.Lock()
	defer s.lock.Unlock()
	upStream := NewQueue(stream, errorChan)
	upStream.Run(wg)
	s.streams[id] = upStream
	return upStream
}

// RemoveQueue removes queue from map
func (s *QueuesMap) RemoveQueue(id int) {
	s.lock.RLock()
	q := s.GetQueue(id)
	if q == nil {
		s.lock.RUnlock()
		return
	}
	s.lock.RUnlock()
	s.lock.Lock()
	q.Flush()
	s.streams[id] = nil
	delete(s.streams, id)
	s.lock.Unlock()
}

// Flush make an attempt to graceful shutdown of all streams
func (s *QueuesMap) Flush() {
	for _, v := range s.streams {
		log.Print("flushing stream")
		v.Flush()
	}
}
