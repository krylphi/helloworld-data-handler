package stream

import (
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/krylphi/helloworld-data-handler/internal/domain"
	"github.com/krylphi/helloworld-data-handler/internal/utils"
)

// StreamHandler handler for incoming data and passing them to respective streams
type StreamHandler struct {
	lock      chan int
	wg        *sync.WaitGroup
	streamMap *QueuesMap
	next      Handler
	streamGen func(filename string) (Stream, error)
}

// NewStreamHandler StreamHandler constructor
func NewStreamHandler(streamGen func(filename string) (Stream, error)) *StreamHandler {
	t, err := strconv.Atoi(utils.GetEnvDef("QUEUE_TIMEOUT_MIN", "60"))
	if err != nil {
		t = 60
	}
	queueTimeout := time.Minute * time.Duration(int64(t))
	return &StreamHandler{
		streamMap: NewQueuesMap(queueTimeout),
		wg:        &sync.WaitGroup{},
		lock:      make(chan int, 1),
		streamGen: streamGen,
	}
}

// Send sends message to stream or creating new, if non existent
func (sh *StreamHandler) Send(e *domain.Entry) error {
	if sh.next != nil {
		go sh.next.Send(e)
	}
	date := utils.DateFromUnixMillis(e.Timestamp)
	key := utils.Concat("chat/", date, "/content_logs_", date, "_", strconv.Itoa(e.ClientID))
	sh.lock <- 1
	upStream := sh.streamMap.GetQueue(e.ClientID)
	if upStream == nil {
		s, err := sh.streamGen(key)
		if err != nil {
			return err
		}
		upStream = sh.streamMap.AddQueue(e.ClientID, s, sh.wg)
	}
	<-sh.lock
	err := upStream.Send(e)
	if err != nil {
		return err
	}
	return nil
}

// Flush clean streams
func (sh *StreamHandler) Flush() {
	sh.streamMap.Flush()
	log.Print("Awaiting streams Flush to end")
	sh.wg.Wait()
	log.Print("Streams flushed successfully")
}
