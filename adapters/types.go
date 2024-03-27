package adapters

import "github.com/aws/aws-sdk-go/aws/session"

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
