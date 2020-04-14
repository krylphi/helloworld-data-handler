package aws

import (
	"github.com/krylphi/helloworld-data-handler/internal/stream"
	"log"
	"strconv"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/krylphi/helloworld-data-handler/internal/domain"
	"github.com/krylphi/helloworld-data-handler/internal/utils"
)

type StreamHandler struct {
	wg         *sync.WaitGroup
	awsSession *session.Session
	streamMap  streamMap
	bucket     string
	next       stream.Stream
}

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
			streams: make(map[int]*uploadStream, 10),
		},
		wg:     &sync.WaitGroup{},
		bucket: awsBucket,
	}
}

func (sh *StreamHandler) Send(e *domain.Entry) {
	if sh.next != nil {
		go sh.next.Send(e)
	}
	date := utils.DateFromUnixMillis(e.Timestamp)
	key := utils.Concat("/chat/", date, "/content_logs_", date, "_", strconv.Itoa(e.ClientId))
	upStream := sh.streamMap.get(e.ClientId)
	if upStream == nil {
		s := newStreamIO()
		upStream = &uploadStream{
			uploader: s3manager.NewUploader(sh.awsSession, func(u *s3manager.Uploader) {
				u.PartSize = 5 * 1024 * 1024 // The minimum/default allowed part size is 5MB
				u.Concurrency = 2            // default is 5
			}),
			reader: s,
		}
		sh.wg.Add(1)
		go func() {
			res, err := upStream.uploader.Upload(&s3manager.UploadInput{
				Bucket: aws.String(sh.bucket),
				Key:    aws.String(key),
				Body:   s,
			})
			if err != nil {
				log.Print(utils.Concat("writing to bucket: ", sh.bucket, " key: ", key, " completed with error: ", err.Error()))
			} else {
				log.Print(utils.Concat("writing to bucket: ", sh.bucket, " key: ", key, " completed with Location: ", res.Location))
			}

			sh.wg.Done()
		}()
		sh.streamMap.set(e.ClientId, upStream)
		s.Run(sh.wg)
	}
	upStream.reader.Send(e)
}

func (sh *StreamHandler) Flush() {
	sh.streamMap.flush()
	log.Print("Awaiting streams flush to end")
	sh.wg.Wait()
	log.Print("Streams flushed successfully")
}
