package aws

import (
	"bytes"
	"compress/gzip"
	"github.com/krylphi/helloworld-data-handler/internal/errs"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/krylphi/helloworld-data-handler/internal/utils"
)

// Stream is an aws upload stream
type Stream struct {
	debug          bool
	buf            *bytes.Buffer
	gzw            *gzip.Writer
	maxRetries     int
	svc            *s3.S3
	completedParts []*s3.CompletedPart
	multipart      *s3.CreateMultipartUploadOutput
}

// NewAWSStream Stream constructor
func NewAWSStream(session *session.Session, inp *s3.CreateMultipartUploadInput) (*Stream, error) {
	svc := s3.New(session)
	multipart, err := svc.CreateMultipartUpload(inp)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	gzw := gzip.NewWriter(buf)
	return &Stream{
		debug:          utils.GetEnvDef("NO_UPLOAD", "0") == "1",
		buf:            buf,
		gzw:            gzw,
		maxRetries:     3,
		svc:            svc,
		completedParts: make([]*s3.CompletedPart, 0),
		multipart:      multipart,
	}, nil
}

// Write write byte data to stream
func (us *Stream) Write(data []byte) (err error) {
	if _, err = us.gzw.Write(data); err != nil {
		log.Print("failed to write to stream")
		return err
	}
	if us.buf.Len() >= utils.MinAWSChunk {
		payload := make([]byte, us.buf.Len())
		if _, err = us.buf.Read(payload); err != nil {
			log.Print("failed to get compressed payload")
			return err
		}
		if err = us.uploadPart(payload); err != nil {
			log.Print("failed to upload compressed payload")
			return err
		}
		if len(us.completedParts) >= 819 { // 4Gb ~ 5Mb * 819.2
			//log.Print("Maximum file size reached, flushing...")
			//err =  us.Flush()
			//if err != nil {
			//	return err
			//}
			return errs.ErrMaxFileSize
		}
	}
	return nil
}

// Flush attempt to flush stream
func (us *Stream) Flush() (err error) {
	log.Printf("Finalizing... %s", *us.multipart.Key)
	log.Printf("Flushing gzip stream... %s", *us.multipart.Key)
	if err = us.gzw.Flush(); err != nil {
		log.Print("failed to Flush gzip stream")
		return
	}
	log.Printf("Closing gzip stream... %s", *us.multipart.Key)
	if err = us.gzw.Close(); err != nil {
		log.Print("failed to Flush gzip stream")
		return
	}
	payload := make([]byte, us.buf.Len())
	log.Printf("Getting gzip footer... %s", *us.multipart.Key)
	if _, err = us.buf.Read(payload); err != nil {
		log.Print("failed to get gzip footer payload")
		return
	}
	log.Print("Uploading gzip footer...")
	if err = us.uploadPart(payload); err != nil {
		log.Print("failed to upload gzip footer payload")
		return
	}
	log.Printf("Completing multipart upload... %s", *us.multipart.Key)
	res, err := us.completeMultipartUpload()
	if err != nil {
		log.Print(utils.Concat("Error completing upload of", *us.multipart.Key, " due to error: ", err.Error()))
		err = us.abortMultipartUpload()
		if err != nil {
			log.Print(utils.Concat("Error aborting uploaded file: ", err.Error()))
		} else {
			log.Print(utils.Concat("Upload aborted:", *us.multipart.Key))
		}
		return
	}
	log.Print(utils.Concat("Successfully uploaded file:", res.String()))
	log.Printf("Stream flushed successfully %s", *us.multipart.Key)
	return
}

func (us *Stream) completeMultipartUpload() (*s3.CompleteMultipartUploadOutput, error) {
	if us.debug {
		return &s3.CompleteMultipartUploadOutput{}, nil
	}
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

func (us *Stream) uploadPart(fileBytes []byte) error {
	tryNum := 1
	partNumber := len(us.completedParts) + 1
	partInput := &s3.UploadPartInput{
		Body:          bytes.NewReader(fileBytes),
		Bucket:        us.multipart.Bucket,
		Key:           us.multipart.Key,
		PartNumber:    aws.Int64(int64(len(us.completedParts) + 1)),
		UploadId:      us.multipart.UploadId,
		ContentLength: aws.Int64(int64(len(fileBytes))),
	}

	if us.debug {
		us.completedParts = append(us.completedParts, &s3.CompletedPart{
			ETag:       aws.String("ETAG"),
			PartNumber: aws.Int64(int64(partNumber)),
		})
		return nil
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

func (us *Stream) abortMultipartUpload() error {
	abortInput := &s3.AbortMultipartUploadInput{
		Bucket:   us.multipart.Bucket,
		Key:      us.multipart.Key,
		UploadId: us.multipart.UploadId,
	}
	_, err := us.svc.AbortMultipartUpload(abortInput)
	return err
}
