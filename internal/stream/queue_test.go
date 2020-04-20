package stream

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"github.com/krylphi/helloworld-data-handler/internal/domain"
	"github.com/krylphi/helloworld-data-handler/internal/mock"
)

func initQueue(t *testing.T) (*UploadQueue, *gomock.Controller, *mock.MockStream, *sync.WaitGroup, chan error) {
	mockCtrl := gomock.NewController(t)
	stream := mock.NewMockStream(mockCtrl)
	ch := make(chan error)
	q := NewQueue(stream, ch)
	wg := &sync.WaitGroup{}
	q.Run(wg)
	return q, mockCtrl, stream, wg, ch
}

func TestUploadQueue_Flush(t *testing.T) {
	t.Run("Flush", func(t *testing.T) {
		q, _, stream, wg, _ := initQueue(t)
		stream.EXPECT().Flush().Times(1)
		q.Flush()
		wg.Wait()
		if got := q.IsClosed(); got != true {
			t.Fatalf("IsClosed() = %v, want %v", got, true)
		}
	})
}

func TestUploadQueue_IsClosed(t *testing.T) {
	q, _, stream, wg, _ := initQueue(t)
	t.Run("Before flush", func(t *testing.T) {
		if got := q.IsClosed(); got != false {
			t.Fatalf("IsClosed() = %v, want %v", got, true)
		}
	})

	t.Run("After flush", func(t *testing.T) {
		stream.EXPECT().Flush().Times(1)
		q.Flush()
		wg.Wait()
		if got := q.IsClosed(); got != true {
			t.Fatalf("IsClosed() = %v, want %v", got, true)
		}
	})
}

func TestUploadQueue_Send(t *testing.T) {
	q, ctrl, stream, _, _ := initQueue(t)
	data := []byte("{\"text\":\"hello world\",\"content_id\":1,\"client_id\":1,\"timestamp\":1586846680064}")
	entry, _ := domain.ParseEntry(data)
	t.Run("Send data", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)
		if got := q.IsClosed(); got != false {
			t.Fatalf("IsClosed() = %v, want %v", got, true)
		}
		do := func(data []byte) {
			wg.Done()
		}
		stream.EXPECT().Write(entry.Marshal()).Times(1).Do(do)
		err := q.Send(entry)
		if err != nil {
			t.Fatalf("Error sending data: %v", err.Error())
		}
		wg.Wait()
	})
	ctrl.Finish()
}

func TestUploadQueue_pushError(t *testing.T) {
	q, ctrl, _, _, errChan := initQueue(t)
	t.Run("Push error", func(t *testing.T) {
		if got := q.IsClosed(); got != false {
			t.Fatalf("IsClosed() = %v, want %v", got, true)
		}
		err := errors.New("some err")
		go q.pushError(err)
		select {
		case <-time.After(3 * time.Second):
			t.Fatalf("Did not recieve error after 3 seconds")
		case act := <-errChan:
			if err != act {
				t.Fatalf("Errors not match: expected %v, got: %v", err.Error(), act.Error())
			}
		}
	})
	ctrl.Finish()
}
