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
	q.Run(wg)
	return q, stream, mockCtrl, wg, ch
}

func TestQueuesMap_AddQueue(t *testing.T) {
	qm := initQueuesMap(10 * time.Second)
	_, s, c, wg, _ := initQueuesMapMocks(t)
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
	_, s, c, wg, _ := initQueuesMapMocks(t)
	_, s2, c2, wg2, _ := initQueuesMapMocks(t)
	t.Run("QueuesMap_Flush", func(t *testing.T) {
		qm.AddQueue(1, s, wg)
		qm.AddQueue(2, s2, wg2)
		if len(qm.streams) < 2 {
			t.Fatalf("Queues was not added")
		}
		//q.EXPECT().Flush().Times(1)
		//q2.EXPECT().Flush().Times(1)
		qm.Flush()
	})
	c2.Finish()
	c.Finish()
}

func TestQueuesMap_GetQueue(t *testing.T) {
	qm := initQueuesMap(10 * time.Second)
	_, s, c, wg, _ := initQueuesMapMocks(t)
	t.Run("Empty", func(t *testing.T) {
		res := qm.GetQueue(1)
		if res != nil {
			t.Fatalf("Queues not empty")
		}
	})

	t.Run("Not Empty", func(t *testing.T) {
		qm.AddQueue(1, s, wg)
		res := qm.GetQueue(1)
		if res == nil {
			t.Fatalf("Queues is empty")
		}
	})
	c.Finish()
}

func TestQueuesMap_RemoveQueue(t *testing.T) {
	qm := initQueuesMap(10 * time.Second)
	_, s, c, wg, _ := initQueuesMapMocks(t)
	t.Run("Empty", func(t *testing.T) {
		before := len(qm.streams)
		qm.RemoveQueue(1)
		if before != len(qm.streams) {
			t.Fatalf("Queues not empty")
		}
	})

	t.Run("Not Empty", func(t *testing.T) {
		qm.AddQueue(1, s, wg)
		qm.RemoveQueue(1)
		res := qm.GetQueue(1)
		if res != nil {
			t.Fatalf("Queues not deleted")
		}
	})
	c.Finish()
}
