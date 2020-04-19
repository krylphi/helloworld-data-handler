package aws

import (
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/krylphi/helloworld-data-handler/internal/domain"
	"github.com/krylphi/helloworld-data-handler/internal/stream"
	"github.com/krylphi/helloworld-data-handler/internal/utils"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// StreamHandler handler for incoming data and passing them to respective streams
type StreamHandler struct {
	lock       chan int
	wg         *sync.WaitGroup
	awsSession *session.Session
	streamMap  *stream.QueuesMap
	bucket     string
	next       stream.Handler
}

// NewStreamHandler StreamHandler constructor
func NewStreamHandler() *StreamHandler {
	var awsConfig *aws.Config
	accessKey := utils.GetEnvDef("AWS_ACCESS_KEY", "")
	accessSecret := utils.GetEnvDef("AWS_ACCESS_SECRET", "")
	awsRegion := utils.GetEnvDef("AWS_REGION", "us-west-2")
	awsBucket := utils.GetEnvDef("AWS_BUCKET", "hw-test-chat")
	t, err := strconv.Atoi(utils.GetEnvDef("QUEUE_TIMEOUT_MIN", "60"))
	if err != nil {
		t = 60
	}
	queueTimeout := time.Minute * time.Duration(int64(t))
	if accessKey == "" || accessSecret == "" {
		//load default credentials
		awsConfig = &aws.Config{
			Region: aws.String(awsRegion),
		}
	} else {
		awsConfig = &aws.Config{
			Region:      aws.String(awsRegion),
			Credentials: credentials.NewStaticCredentials(accessKey, accessSecret, ""),
		}
	}

	// The session the S3 Uploader will use
	sess := session.Must(session.NewSession(awsConfig))
	return &StreamHandler{
		awsSession: sess,
		streamMap:  stream.NewQueuesMap(queueTimeout),
		wg:         &sync.WaitGroup{},
		bucket:     awsBucket,
		lock:       make(chan int, 1),
	}
}

// Send sends message to stream or creating new, if non existent
func (sh *StreamHandler) Send(e *domain.Entry) error {
	if sh.next != nil {
		go sh.next.Send(e)
	}
	date := utils.DateFromUnixMillis(e.Timestamp)
	key := utils.Concat("/chat/", date, "/content_logs_", date, "_", strconv.Itoa(e.ClientID))
	sh.lock <- 1
	upStream := sh.streamMap.GetQueue(e.ClientID)
	if upStream == nil {
		input := &s3.CreateMultipartUploadInput{
			Bucket:      aws.String(sh.bucket),
			Key:         aws.String(key),
			ContentType: aws.String("application/gzip"),
		}
		s, err := NewAWSStream(sh.awsSession, input)
		if err != nil {
			return err
		}
		errorChan := make(chan error)
		upStream = stream.NewQueue(s, errorChan)
		sh.streamMap.AddQueue(e.ClientID, upStream, errorChan)
		upStream.Run(sh.wg)
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
