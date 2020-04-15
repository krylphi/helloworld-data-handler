package aws

import (
	"log"
	"strconv"
	"sync"

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
	wg         *sync.WaitGroup
	awsSession *session.Session
	streamMap  streamMap
	bucket     string
	next       stream.Stream
}

// NewStreamHandler StreamHandler constructor
func NewStreamHandler() *StreamHandler {
	var awsConfig *aws.Config
	accessKey := utils.GetEnvDef("AWS_ACCESS_KEY", "")
	accessSecret := utils.GetEnvDef("AWS_ACCESS_SECRET", "")
	awsRegion := utils.GetEnvDef("AWS_REGION", "us-west-2")
	awsBucket := utils.GetEnvDef("AWS_BUCKET", "hw-test-chat")
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
		streamMap: streamMap{
			lock:    sync.RWMutex{},
			streams: make(map[int]*streamIO, 10),
		},
		wg:     &sync.WaitGroup{},
		bucket: awsBucket,
	}
}

// Send sends message to stream or creating new, if non existent
func (sh *StreamHandler) Send(e *domain.Entry) error {
	if sh.next != nil {
		go sh.next.Send(e)
	}
	date := utils.DateFromUnixMillis(e.Timestamp)
	key := utils.Concat("/chat/", date, "/content_logs_", date, "_", strconv.Itoa(e.ClientID))
	upStream := sh.streamMap.get(e.ClientID)
	if upStream == nil {
		input := &s3.CreateMultipartUploadInput{
			Bucket:      aws.String(sh.bucket),
			Key:         aws.String(key),
			ContentType: aws.String("application/gzip"),
		}
		s, err := newUploadStream(sh.awsSession, input)
		if err != nil {
			return err
		}
		upStream = newStreamIO(s)
		sh.streamMap.set(e.ClientID, upStream)
		upStream.Run(sh.wg)
	}
	upStream.Send(e)
	return nil
}

// Flush clean streams
func (sh *StreamHandler) Flush() {
	sh.streamMap.flush()
	log.Print("Awaiting streams flush to end")
	sh.wg.Wait()
	log.Print("Streams flushed successfully")
}
