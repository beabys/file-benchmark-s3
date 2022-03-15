package adapters

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type S3Session struct {
	Bucket  string
	session *session.Session
}

type S3Config struct {
	Region           string
	AccessKeyID      string
	SecretAccessKey  string
	Endpoint         string
	Bucket           string
	DisableSSL       bool
	S3ForcePathStyle bool
}

// New create a new session to connect to s3 Object storage
func NewS3(s *S3Config) (*S3Session, error) {
	s3Config := &aws.Config{
		Region:           aws.String(s.Region),
		Credentials:      credentials.NewStaticCredentials(s.AccessKeyID, s.SecretAccessKey, ""),
		Endpoint:         aws.String(s.Endpoint),
		DisableSSL:       aws.Bool(s.DisableSSL),
		S3ForcePathStyle: aws.Bool(s.S3ForcePathStyle),
	}

	sess, err := session.NewSession(s3Config)
	if err != nil {
		return nil, err
	}

	s3Session := &S3Session{
		Bucket:  s.Bucket,
		session: sess,
	}
	return s3Session, nil
}

// try to connect
func (s *S3Session) Connect() error {
	svc := s3.New(s.session)
	_, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(s.Bucket),
	})
	return err
}

//UploadObject is a function to upload to s3 onject storage
func (s *S3Session) UploadObject(src, objectKey string) error {
	// Create an uploader with the session and default options
	u := s3manager.NewUploader(s.session)
	var err error
	file, err := os.Open(src)
	if err != nil {
		return err
	}

	defer file.Close()

	// Upload the file to S3
	_, err = u.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(objectKey),
		Body:   file,
	})
	return err
}

// Download file from s3 bucket, dst includes filename
func (s *S3Session) DownloadObject(objectKey, dst string) error {
	// create neccessary folder adn file
	// if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
	// 	return err
	// }
	file, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create an downloader with the session and default options
	downloader := s3manager.NewDownloader(s.session)

	// Download the file from S3
	_, err = downloader.Download(file, &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		fmt.Println("Failed to download file", err)
	}
	return err
}

// DeleteObject file on s3 bucket
func (s *S3Session) DeleteObject(objectKey string) error {
	svc := s3.New(s.session)

	_, err := svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return err
	}
	return svc.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(objectKey),
	})
}
