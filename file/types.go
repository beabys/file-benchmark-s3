package file

import (
	"time"

	"github.com/beabys/file-benchmark-s3/adapters"
)

// File is a struct to define files attributes
type File struct {
	Size         int
	MD5          string
	Path         string
	DownloadPath string
	Name         string
	Execution    *ExecutionTime
	S3Session    *adapters.S3Session
}

// Times of the execution
type ExecutionTime struct {
	CreationDuration       time.Duration
	UploadDuration         time.Duration
	DownloadDuration       time.Duration
	DeleteLocalDuration    time.Duration
	DeleteDownloadDuration time.Duration
	DeleteUploadDuration   time.Duration
}
