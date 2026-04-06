package oss

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

var (
	AllowedImageMimeTypes = map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/jpg":  true,
	}
	AllowedDocMimeTypes = map[string]bool{
		"application/pdf":          true,
		"application/vnd.ms-excel": true,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
	}
	MaxImageSize int64 = 5 * 1024 * 1024
	MaxDocSize   int64 = 20 * 1024 * 1024
)

type Config struct {
	Endpoint        string
	AccessKeyID     string
	AccessKeySecret string
	BucketName      string
}

type Client struct {
	client *oss.Client
	bucket *oss.Bucket
	cfg    Config
}

func NewClient(cfg Config) (*Client, error) {
	client, err := oss.New(cfg.Endpoint, cfg.AccessKeyID, cfg.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("init oss client: %w", err)
	}
	bucket, err := client.Bucket(cfg.BucketName)
	if err != nil {
		return nil, fmt.Errorf("get oss bucket: %w", err)
	}
	return &Client{client: client, bucket: bucket, cfg: cfg}, nil
}

func (c *Client) GeneratePutURL(ctx context.Context, bizType string, orgID int64, filename string, fileSize int64, contentType string, expiry time.Duration) (string, error) {
	ext := path.Ext(filename)
	_ = ext
	isImage := AllowedImageMimeTypes[contentType]
	isDoc := AllowedDocMimeTypes[contentType]
	if !isImage && !isDoc {
		return "", fmt.Errorf("file type %s not allowed", contentType)
	}
	if isImage && fileSize > MaxImageSize {
		return "", fmt.Errorf("image size %d exceeds limit %d", fileSize, MaxImageSize)
	}
	if isDoc && fileSize > MaxDocSize {
		return "", fmt.Errorf("document size %d exceeds limit %d", fileSize, MaxDocSize)
	}

	now := time.Now()
	objectKey := fmt.Sprintf("%s/org_%d/%04d-%02d/%s", bizType, orgID, now.Year(), now.Month(), filename)

	signedURL, err := c.bucket.SignURL(objectKey, oss.HTTPPut, int64(expiry.Seconds()))
	if err != nil {
		return "", fmt.Errorf("sign put url: %w", err)
	}
	return signedURL, nil
}

func (c *Client) GenerateGetURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	signedURL, err := c.bucket.SignURL(objectKey, oss.HTTPGet, int64(expiry.Seconds()))
	if err != nil {
		return "", fmt.Errorf("sign get url: %w", err)
	}
	return signedURL, nil
}

func BuildObjectKey(bizType string, orgID int64, filename string) string {
	now := time.Now()
	return fmt.Sprintf("%s/org_%d/%04d-%02d/%s", bizType, orgID, now.Year(), now.Month(), filename)
}

func ValidateFileType(contentType string) bool {
	return AllowedImageMimeTypes[contentType] || AllowedDocMimeTypes[contentType]
}

func ValidateFileSize(contentType string, fileSize int64) error {
	if AllowedImageMimeTypes[contentType] && fileSize > MaxImageSize {
		return fmt.Errorf("image size %d exceeds limit %d", fileSize, MaxImageSize)
	}
	if AllowedDocMimeTypes[contentType] && fileSize > MaxDocSize {
		return fmt.Errorf("document size %d exceeds limit %d", fileSize, MaxDocSize)
	}
	return nil
}
