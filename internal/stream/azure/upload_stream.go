package azure

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"net/url"
	"sync"

	"github.com/Azure/azure-storage-blob-go/azblob"

	"github.com/krylphi/helloworld-data-handler/internal/errs"
	"github.com/krylphi/helloworld-data-handler/internal/stream"
	"github.com/krylphi/helloworld-data-handler/internal/utils"
)

type azureStream struct {
	mx sync.Mutex
	containerURL azblob.ContainerURL
	buf          *bytes.Buffer
	gzw          *gzip.Writer
	wg sync.WaitGroup
	closed       bool
}

func NewStreamHandler(errChan chan error) *stream.StreamHandler {
	return stream.NewStreamHandler(func(filename string) (stream.Stream, error) {
		return newAzureBlobStream(filename, errChan)
	})
}

func newAzureBlobStream(path string, errChan chan error) (stream.Stream, error) {
	account := utils.GetEnvDef("AZURE_ACC", "")
	accessKey := utils.GetEnvDef("AZURE_KEY", "")
	if account == "" {
		return nil, errs.ErrEmptyAzureAcc
	}

	if accessKey == "" {
		return nil, errs.ErrEmptyAzureKey
	}

	credential, err := azblob.NewSharedKeyCredential(account, accessKey)
	if err != nil {
		return nil, fmt.Errorf("can't create azure credentials: %w", err)
	}
	res, err := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net/%s", account, path))
	if err != nil {
		return nil, fmt.Errorf("can't parse azure container url: %w", err)
	}
	p := azblob.NewPipeline(credential, azblob.PipelineOptions{})
	containerURL := azblob.NewContainerURL(*res, p)
	buf := &bytes.Buffer{}
	gzw := gzip.NewWriter(buf)
	aStream := &azureStream{
		buf: buf,
		gzw: gzw,
		containerURL: containerURL,
		wg: sync.WaitGroup{},
	}
	blobURL := containerURL.NewBlockBlobURL(path)
	aStream.wg.Add(1)
	go func() {
		_, err := azblob.UploadStreamToBlockBlob(context.Background(), aStream, blobURL, azblob.UploadStreamToBlockBlobOptions{
			BufferSize: utils.MinDataChunk,
			MaxBuffers: 819,
			BlobHTTPHeaders: azblob.BlobHTTPHeaders{
				ContentEncoding: "gzip",
			},
		})
		if err != nil {
			if errChan != nil {
				errChan <- err
			}
		}
		aStream.wg.Done()
	}()
	return aStream, nil
}

func (s *azureStream) Write(data []byte) (err error) {
	if _, err = s.gzw.Write(data); err != nil {
		log.Print("failed to write to stream")
		return err
	}
	if s.buf.Cap() >= utils.MaxDataLen {
		return errs.ErrMaxFileSize
	}
	return nil
}

func (s *azureStream) Read(p []byte) (n int, err error) {
	//return s.buf.Read(p)
	res, err := s.buf.Read(p)
	if s.isClosed() && res == 0 && err == io.EOF {
		return 0, io.EOF
	}
	if res == 0 && err == io.EOF {
		return 0, nil
	}
	return res, nil
}

func (s *azureStream) Flush() (err error) {
	log.Printf("Finalizing... %s", s.containerURL.String())
	log.Printf("Flushing gzip stream... %s", s.containerURL.String())
	if err = s.gzw.Flush(); err != nil {
		log.Print("failed to Flush gzip stream")
		return
	}
	log.Printf("Closing gzip stream... %s", s.containerURL.String())
	if err = s.gzw.Close(); err != nil {
		log.Print("failed to Flush gzip stream")
		return
	}
	s.wg.Wait()
	s.mx.Lock()
	defer s.mx.Unlock()
	s.closed = true
	return
}

func (s *azureStream) isClosed() bool {
	s.mx.Lock()
	defer s.mx.Unlock()
	return s.closed
}
