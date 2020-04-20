package stream

import (
	"github.com/golang/mock/gomock"
	"github.com/krylphi/helloworld-data-handler/internal/mock"
	"sync"
	"testing"
	"time"
)

func initQueuesMap(d time.Duration) *QueuesMap {
	return NewQueuesMap(d)
}

func initQueuesMapMocks(t *testing.T) (*mock.MockQueue, *mock.MockStream, *gomock.Controller, *sync.WaitGroup, chan error) {
	mockCtrl := gomock.NewController(t)
	stream := mock.NewMockStream(mockCtrl)
	ch := make(chan error)
	q := mock.NewMockQueue(mockCtrl)
	wg := &sync.WaitGroup{}
	q.EXPECT().Run(wg).Times(1)
	return q, stream, mockCtrl, wg, ch
}

func TestQueuesMap_AddQueue(t *testing.T) {
	qm := initQueuesMap(10 * time.Second)
	var q *mock.MockQueue
	var s *mock.MockStream
	var c *gomock.Controller
	var wg *sync.WaitGroup
	q, s, c, wg, _ = initQueuesMapMocks(t)
	qm.QueueGen = func(stream Stream, errChan chan error) Queue {
		return q
	}
	t.Run("AddQueue", func(t *testing.T) {
		qm.AddQueue(1, s, wg)
		if len(qm.streams) < 1 {
			t.Fatalf("Queue was not added")
		}
	})
	c.Finish()
}

func TestQueuesMap_Flush(t *testing.T) {
	qm := initQueuesMap(10 * time.Second)
	qm.QueueGen = func(stream Stream, errChan chan error) Queue {
		q, _, _, _, _ := initQueuesMapMocks(t)
		return q
	}
	_, s, _, wg, _ := initQueuesMapMocks(t)
	_, s2, _, _, _ := initQueuesMapMocks(t)
	t.Run("QueuesMap_Flush", func(t *testing.T) {
		q1 := qm.AddQueue(1, s, wg)
		q2 := qm.AddQueue(2, s2, wg)
		if len(qm.streams) < 2 {
			t.Fatalf("Queues was not added")
		}
		q1.(*mock.MockQueue).EXPECT().Flush().Times(1)
		q2.(*mock.MockQueue).EXPECT().Flush().Times(1)
		qm.Flush()
		//wg.Wait()
		//if !q1.IsClosed() {
		//	t.Fatalf("Queues was not closed")
		//}
		//if !q2.IsClosed() {
		//	t.Fatalf("Queues was not closed")
		//}
	})
	//c2.Finish()
	//c.Finish()
}

func TestQueuesMap_GetQueue(t *testing.T) {
	qm := initQueuesMap(10 * time.Second)
	var q *mock.MockQueue
	var s *mock.MockStream
	var c *gomock.Controller
	var wg *sync.WaitGroup
	q, s, c, wg, _ = initQueuesMapMocks(t)
	qm.QueueGen = func(stream Stream, errChan chan error) Queue {
		return q
	}
	t.Run("Empty", func(t *testing.T) {
		res := qm.GetQueue(1)
		if res != nil {
			t.Fatalf("Queues not empty")
		}
	})

	t.Run("Not Empty", func(t *testing.T) {
		qm.AddQueue(1, s, wg)
		q.EXPECT().IsClosed().Times(1)
		res := qm.GetQueue(1)
		if res == nil {
			t.Fatalf("Queues is empty")
		}
	})
	c.Finish()
}

func TestQueuesMap_RemoveQueue(t *testing.T) {
	qm := initQueuesMap(10 * time.Second)
	var q *mock.MockQueue
	var s *mock.MockStream
	var c *gomock.Controller
	var wg *sync.WaitGroup
	q, s, c, wg, _ = initQueuesMapMocks(t)
	qm.QueueGen = func(stream Stream, errChan chan error) Queue {
		return q
	}
	t.Run("Empty", func(t *testing.T) {
		before := len(qm.streams)
		qm.RemoveQueue(1)
		if before != len(qm.streams) {
			t.Fatalf("Queues not empty")
		}
	})

	t.Run("Not Empty", func(t *testing.T) {
		qm.AddQueue(1, s, wg)
		first := q.EXPECT().IsClosed().Times(1)
		q.EXPECT().Flush().Times(1).After(first)
		qm.RemoveQueue(1)
		res := qm.GetQueue(1)
		if res != nil {
			t.Fatalf("Queues not deleted")
		}
	})
	c.Finish()
}
