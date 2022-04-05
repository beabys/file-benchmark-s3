package file

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
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

// Create the file and return  error if any
func (f *File) Create() error {
	buf := make([]byte, f.Size)
	md5Sum := md5.Sum(buf)
	err := os.WriteFile(fmt.Sprintf("%s%s", f.Path, f.Name), buf, 0666)
	if err != nil {
		return err
	}
	f.MD5 = hex.EncodeToString(md5Sum[:])
	return nil
}

// GetMD5 return a md5 string of the file path provided
func GetMD5(filePath string) (string, error) {
	var md5Verify string
	buf, err := os.ReadFile(filePath)
	if err != nil {
		return md5Verify, err
	}
	md5Sum := md5.Sum(buf)

	//Convert the bytes to a string and sabe the Md5
	md5Verify = hex.EncodeToString(md5Sum[:])

	return md5Verify, nil
}

// Copy the file to a especific location and return  error if any
func (f *File) Copy() error {
	src := fmt.Sprintf("%s%s", f.Path, f.Name)
	dst := fmt.Sprintf("%s%s", f.DownloadPath, f.Name)
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}
