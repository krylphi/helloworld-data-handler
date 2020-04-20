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

func initQueuesMapMocks(t *testing.T) (*mock.MockQueue, *gomock.Controller, chan error) {
	mockCtrl := gomock.NewController(t)
	//stream := mock.NewMockStream(mockCtrl)
	ch := make(chan error)
	q := mock.NewMockQueue(mockCtrl)
	wg := &sync.WaitGroup{}
	q.EXPECT().Run(wg).Times(1)
	q.Run(wg)
	return q, mockCtrl, ch
}

func TestQueuesMap_AddQueue(t *testing.T) {
	qm := initQueuesMap(10 * time.Second)
	q, c, ch := initQueuesMapMocks(t)
	t.Run("AddQueue", func(t *testing.T) {
		qm.AddQueue(1, q, ch)
		if len(qm.streams) < 1 {
			t.Fatalf("Queue was not added")
		}
	})
	c.Finish()
}

func TestQueuesMap_Flush(t *testing.T) {
	qm := initQueuesMap(10 * time.Second)
	q, c, ch := initQueuesMapMocks(t)
	q2, c2, ch2 := initQueuesMapMocks(t)
	t.Run("QueuesMap_Flush", func(t *testing.T) {
		qm.AddQueue(1, q, ch)
		qm.AddQueue(2, q2, ch2)
		if len(qm.streams) < 2 {
			t.Fatalf("Queues was not added")
		}
		q.EXPECT().Flush().Times(1)
		q2.EXPECT().Flush().Times(1)
		qm.Flush()
	})
	c2.Finish()
	c.Finish()
}

func TestQueuesMap_GetQueue(t *testing.T) {
	qm := initQueuesMap(10 * time.Second)
	q, c, ch := initQueuesMapMocks(t)
	t.Run("Empty", func(t *testing.T) {
		res := qm.GetQueue(1)
		if res != nil {
			t.Fatalf("Queues not empty")
		}
	})

	t.Run("Not Empty", func(t *testing.T) {
		qm.AddQueue(1, q, ch)
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
	q, c, ch := initQueuesMapMocks(t)
	t.Run("Empty", func(t *testing.T) {
		before := len(qm.streams)
		qm.RemoveQueue(1)
		if before != len(qm.streams) {
			t.Fatalf("Queues not empty")
		}
	})

	t.Run("Not Empty", func(t *testing.T) {
		qm.AddQueue(1, q, ch)
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
