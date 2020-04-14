package aws

import (
	"bytes"
	"compress/gzip"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/krylphi/helloworld-data-handler/internal/utils"
	"log"
	"sync"

	"github.com/krylphi/helloworld-data-handler/internal/domain"
)

type streamIO struct {
	buf    *bytes.Buffer
	gzw    *gzip.Writer
	queue  chan *domain.Entry
	closed bool
	wg     *sync.WaitGroup
	stream *uploadStream
}

func newStreamIO(stream *uploadStream) *streamIO {
	buf := &bytes.Buffer{}
	gzw := gzip.NewWriter(buf)
	return &streamIO{
		buf:    buf,
		gzw:    gzw,
		queue:  make(chan *domain.Entry, 10),
		closed: false,
		stream: stream,
	}
}

func (s *streamIO) Send(e *domain.Entry) {
	s.queue <- e
}

func (s *streamIO) Run(wg *sync.WaitGroup) {
	s.wg = wg
	s.wg.Add(1)
	var err error = nil
	go func() {
		for e := range s.queue {
			//log.Print(utils.Concat("New entity:", string(e.Marshal())))
			if _, err = s.gzw.Write(e.Marshal()); err != nil {
				log.Print("failed to write to stream")
			}
			if s.buf.Len() >= 5 * 1024 * 1024 { // 5 mb is minimum allowed chunk
				payload := make([]byte, s.buf.Len())
				if _, err = s.buf.Read(payload); err != nil {
					log.Print("failed to get compressed payload")
				}
				if err = s.stream.uploadPart(payload); err != nil {
					log.Print("failed to upload compressed payload")
				}
			}
		}
		if err = s.gzw.Flush(); err != nil {
			log.Print("failed to flush gzip stream")
		}
		if err = s.gzw.Close(); err != nil {
			log.Print("failed to flush gzip stream")
		}
		payload := make([]byte, s.buf.Len())
		if _, err = s.buf.Read(payload); err != nil {
			log.Print("failed to get gzip footer payload")
		}
		if err = s.stream.uploadPart(payload); err != nil {
			log.Print("failed to upload gzip footer payload")
		}
		res, err := s.stream.completeMultipartUpload()
		if err != nil {
			log.Print(utils.Concat("Error completing upload of", *s.stream.multipart.Key, " due to error: ", err.Error()))
			err = s.stream.abortMultipartUpload()
			if err != nil {
				log.Print(utils.Concat("Error aborting uploaded file: ", err.Error()))
			} else {
				log.Print(utils.Concat("Upload aborted:", *s.stream.multipart.Key))
			}
		}
		log.Print(utils.Concat("Successfully uploaded file:", res.String()))
		s.closed = true
		s.wg.Done()
	}()
}

func (s *streamIO) Flush() {
	go func() {
		close(s.queue)
	}()
}

type streamMap struct {
	lock    sync.RWMutex
	streams map[int]*streamIO
}

type uploadStream struct {
	maxRetries int
	svc *s3.S3
	completedParts []*s3.CompletedPart
	multipart *s3.CreateMultipartUploadOutput
}

func newUploadStream(session *session.Session, inp *s3.CreateMultipartUploadInput) (*uploadStream, error) {
	svc:= s3.New(session)
	multipart, err := svc.CreateMultipartUpload(inp)
	if err != nil {
		return nil, err
	}
	return &uploadStream{
		maxRetries:     3,
		svc:            svc,
		completedParts: make([]*s3.CompletedPart, 0),
		multipart:      multipart,
	}, nil
}

func (us *uploadStream)completeMultipartUpload() (*s3.CompleteMultipartUploadOutput, error) {
	completeInput := &s3.CompleteMultipartUploadInput{
		Bucket:   us.multipart.Bucket,
		Key:      us.multipart.Key,
		UploadId: us.multipart.UploadId,
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: us.completedParts,
		},
	}
	return us.svc.CompleteMultipartUpload(completeInput)
}

func (us *uploadStream)uploadPart(fileBytes []byte) error {
	tryNum := 1
	partNumber := len(us.completedParts)+1
	partInput := &s3.UploadPartInput{
		Body:          bytes.NewReader(fileBytes),
		Bucket:        us.multipart.Bucket,
		Key:           us.multipart.Key,
		PartNumber:    aws.Int64(int64(len(us.completedParts)+1)),
		UploadId:      us.multipart.UploadId,
		ContentLength: aws.Int64(int64(len(fileBytes))),
	}

	for tryNum <= us.maxRetries {
		uploadResult, err := us.svc.UploadPart(partInput)
		if err != nil {
			if tryNum == us.maxRetries {
				if aerr, ok := err.(awserr.Error); ok {
					return aerr
				}
				return err
			}
			tryNum++
		} else {
			us.completedParts = append(us.completedParts, &s3.CompletedPart{
				ETag:       uploadResult.ETag,
				PartNumber: aws.Int64(int64(partNumber)),
			})
			return nil
		}
	}
	return nil
}

func (us *uploadStream)abortMultipartUpload() error {
	abortInput := &s3.AbortMultipartUploadInput{
		Bucket:   us.multipart.Bucket,
		Key:      us.multipart.Key,
		UploadId: us.multipart.UploadId,
	}
	_, err := us.svc.AbortMultipartUpload(abortInput)
	return err
}


func (s *streamMap) get(key int) *streamIO {
	s.lock.RLock()
	res := s.streams[key]
	s.lock.RUnlock()
	return res
}

func (s *streamMap) set(key int, stream *streamIO) {
	s.lock.Lock()
	s.streams[key] = stream
	s.lock.Unlock()
}

func (s *streamMap) flush() {
	for _, v := range s.streams {
		log.Print("flushing stream")
		v.Flush()
	}
}
