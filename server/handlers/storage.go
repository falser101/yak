package handlers

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var minioClient *minio.Client
var minioBucket string
var minioEndpoint string
var minioSecure bool

// InitMinIO 初始化 MinIO 客户端
func InitMinIO(endpoint, accessKey, secretKey, bucket string, secure bool) error {
	var err error
	minioClient, err = minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: secure,
	})
	if err != nil {
		return fmt.Errorf("minio new client error: %v", err)
	}
	minioBucket = bucket
	minioEndpoint = endpoint
	minioSecure = secure

	// 确保 bucket 存在
	ctx := context.Background()
	exists, err := minioClient.BucketExists(ctx, bucket)
	if err != nil {
		return fmt.Errorf("minio bucket exists check error: %v", err)
	}
	if !exists {
		err = minioClient.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("minio make bucket error: %v", err)
		}
	}

	// 设置 bucket 为 public（允许匿名访问）
	policy := fmt.Sprintf(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::%s/*"]}]}`, bucket)
	err = minioClient.SetBucketPolicy(ctx, bucket, policy)
	if err != nil {
		// 设置 policy 失败不影响主流程，只是该 bucket 可能需要手动设置 public
		fmt.Printf("warn: set bucket policy error: %v\n", err)
	}

	return nil
}

// UploadToMinIO 上传文件到 MinIO，返回公开访问 URL
func UploadToMinIO(objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	ctx := context.Background()
	_, err := minioClient.PutObject(ctx, minioBucket, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("minio put object error: %v", err)
	}
	scheme := "http"
	if minioSecure {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", scheme, minioEndpoint, minioBucket, objectName), nil
}
