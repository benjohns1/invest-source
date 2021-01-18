package provider

import (
	"bytes"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	awsS3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// S3 provider for a key-value cache.
type S3 struct {
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
}

// NewS3 creates a new S3 key-value cache provider.
func NewS3(cfgProvider client.ConfigProvider) (*S3, error) {
	uploader := s3manager.NewUploader(cfgProvider)
	if uploader == nil {
		return nil, fmt.Errorf("could not construct S3 uploader")
	}
	downloader := s3manager.NewDownloader(cfgProvider)
	if downloader == nil {
		return nil, fmt.Errorf("could not construct S3 downloader")
	}
	return &S3{uploader, downloader}, nil
}

// Upload a byte array to an S3 bucket at the given key location.
func (s3 S3) Upload(bucket, key string, value []byte) error {
	if _, err := s3.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(value),
	}); err != nil {
		return fmt.Errorf("error uploading S3 object: %v", err)
	}

	return nil
}

// Download a byte array from an S3 bucket with the given key.
func (s3 S3) Download(bucket, key string) ([]byte, error) {
	buf := &aws.WriteAtBuffer{}
	if _, err := s3.downloader.Download(buf, &awsS3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}); err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == awsS3.ErrCodeNoSuchKey {
				return nil, nil // return nil if key doesn't exist
			}
		}
		return nil, fmt.Errorf("error downloading S3 object: %v", err)
	}

	return buf.Bytes(), nil
}
