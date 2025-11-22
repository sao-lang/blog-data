// internal/minio/minio.go
package minio

import (
	"context"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Config MinIO 配置
type Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	Bucket          string
}

// Client 封装 MinIO 客户端
type Client struct {
	cfg        Config
	client     *minio.Client
	bucketName string
}

// NewClient 初始化 MinIO 客户端
func NewClient(cfg Config) (*Client, error) {
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}

	// 确保 bucket 存在
	exists, errBucketExists := minioClient.BucketExists(context.Background(), cfg.Bucket)
	if errBucketExists != nil {
		return nil, errBucketExists
	}
	if !exists {
		err = minioClient.MakeBucket(context.Background(), cfg.Bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, err
		}
	}

	return &Client{
		cfg:        cfg,
		client:     minioClient,
		bucketName: cfg.Bucket,
	}, nil
}

// Upload 上传文件
func (c *Client) Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error {
	_, err := c.client.PutObject(ctx, c.bucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

// Download 下载文件
func (c *Client) Download(ctx context.Context, objectName string) (io.ReadCloser, error) {
	return c.client.GetObject(ctx, c.bucketName, objectName, minio.GetObjectOptions{})
}

// Delete 删除文件
func (c *Client) Delete(ctx context.Context, objectName string) error {
	return c.client.RemoveObject(ctx, c.bucketName, objectName, minio.RemoveObjectOptions{})
}

// List 列出文件
func (c *Client) List(ctx context.Context, prefix string) ([]string, error) {
	objectCh := c.client.ListObjects(ctx, c.bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	var objects []string
	for obj := range objectCh {
		if obj.Err != nil {
			return nil, obj.Err
		}
		objects = append(objects, obj.Key)
	}
	return objects, nil
}

func (c *Client) PresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	if expiry == 0 {
		expiry = 15 * time.Minute
	}
	reqParams := make(url.Values)

	urlObj, err := c.client.PresignedGetObject(ctx, c.bucketName, objectName, expiry, reqParams)
	if err != nil {
		return "", err
	}
	return urlObj.String(), nil
}
