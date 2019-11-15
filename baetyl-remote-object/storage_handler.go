package main

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/baidubce/bce-sdk-go/bce"
	"github.com/baidubce/bce-sdk-go/services/bos"
	"github.com/baidubce/bce-sdk-go/services/bos/api"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// IObjectStorage interface
type IObjectStorage interface {
	PutObjectFromFile(Bucket, remotePath, filename string, meta map[string]string) error
	FileExists(Bucket, remotePath, md5 string) bool
}

// NewObjectStorageHandler NewObjectStorageHandler
func NewObjectStorageHandler(cfg ClientInfo) (IObjectStorage, error) {
	switch cfg.Kind {
	case Bos:
		return NewBosHandler(cfg)
	case Ceph:
		return NewCephClient(cfg)
	case S3:
		return NewS3Client(cfg)
	default:
		return nil, fmt.Errorf("kind type unexpected")
	}
}

// BosHandler BosHandler
type BosHandler struct {
	bos *bos.Client
	cfg ClientInfo
}

// NewBosHandler creates a new newBosClient
func NewBosHandler(cfg ClientInfo) (*BosHandler, error) {
	bos, err := bos.NewClient(cfg.Ak, cfg.Sk, cfg.Address)
	bos.MultipartSize = cfg.MultiPart.PartSize
	bos.MaxParallel = (int64)(cfg.MultiPart.Concurrency)
	if err != nil {
		return nil, fmt.Errorf("failed to create boe client (%s): %s", cfg.Name, err.Error())
	}
	bos.Config.ConnectionTimeoutInMillis = (int)(cfg.Timeout / time.Millisecond)
	bos.Config.Retry = bce.NewBackOffRetryPolicy(cfg.Retry.Max, (int64)(cfg.Retry.Delay/time.Millisecond), (int64)(cfg.Retry.Base/time.Millisecond))
	b := &BosHandler{
		bos: bos,
		cfg: cfg,
	}
	return b, nil
}

// PutObjectFromFile upload file
func (cli *BosHandler) PutObjectFromFile(Bucket, remotePath, filename string, meta map[string]string) error {
	args := new(api.PutObjectArgs)
	args.UserMeta = meta
	_, err := cli.bos.PutObjectFromFile(Bucket, remotePath, filename, args)
	return err
}

// FileExists FileExists
func (cli *BosHandler) FileExists(Bucket, remotePath, md5 string) bool {
	res, _ := cli.bos.GetObjectMeta(Bucket, remotePath)
	if res != nil {
		if res.ObjectMeta.ContentMD5 == md5 {
			return true
		}
	}
	return false
}

// S3Handler S3Handler
type S3Handler struct {
	s3Client *s3.S3
	uploader *s3manager.Uploader
	cfg      ClientInfo
}

// NewCephClient creates a new NewCephClient
func NewCephClient(cfg ClientInfo) (*S3Handler, error) {
	// Configure to use S3 Server
	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(cfg.Ak, cfg.Sk, ""),
		Endpoint:         aws.String(cfg.Address),
		Region:           aws.String("us-east-1"),
		DisableSSL:       aws.Bool(!strings.HasPrefix(cfg.Address, "https")),
		S3ForcePathStyle: aws.Bool(true),
	}
	newSession := session.New(s3Config)
	s3Client := s3.New(newSession)
	uploader := s3manager.NewUploader(newSession)
	c := &S3Handler{
		s3Client: s3Client,
		cfg:      cfg,
		uploader: uploader,
	}
	return c, nil
}

// NewS3Client creates a new NewS3Client
func NewS3Client(cfg ClientInfo) (*S3Handler, error) {
	// Configure to use S3 Server
	s3Config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(cfg.Ak, cfg.Sk, ""),
		Region:      aws.String(cfg.Region),
	}
	newSession := session.New(s3Config)
	s3Client := s3.New(newSession)
	uploader := s3manager.NewUploader(newSession)
	c := &S3Handler{
		s3Client: s3Client,
		cfg:      cfg,
		uploader: uploader,
	}
	return c, nil
}

// PutObjectFromFile upload file
func (cli *S3Handler) PutObjectFromFile(Bucket, remotePath, filename string, meta map[string]string) error {
	Metadata := make(map[string]*string)
	for k, v := range meta {
		Metadata[k] = &v
	}
	f, err := os.Open(filename)
	defer f.Close()
	if err != nil {
		return err
	}
	params := &s3manager.UploadInput{
		Bucket:   aws.String(Bucket),     // Required
		Key:      aws.String(remotePath), // Required
		Body:     f,
		Metadata: Metadata,
	}
	ctx, cancel := context.WithTimeout(context.Background(), cli.cfg.Timeout)
	defer cancel()
	// _, err = cli.ceph.PutObjectWithContext(ctx, params)
	_, err = cli.uploader.UploadWithContext(ctx, params, func(u *s3manager.Uploader) {
		u.PartSize = cli.cfg.MultiPart.PartSize
		u.LeavePartsOnError = true
		u.Concurrency = cli.cfg.MultiPart.Concurrency
	}) //并发数

	return err
}

// FileExists FileExists
func (cli *S3Handler) FileExists(Bucket, remotePath, md5 string) bool {
	cparams := &s3.HeadObjectInput{
		Bucket: aws.String(Bucket),
		Key:    aws.String(remotePath),
	}
	ho, err := cli.s3Client.HeadObject(cparams)
	if err != nil {
		return false
	}
	input, _ := hex.DecodeString(strings.Replace(*ho.ETag, "\"", "", -1))
	encodeString := base64.StdEncoding.EncodeToString(input)
	if encodeString != md5 {
		return false
	}
	return true
}
