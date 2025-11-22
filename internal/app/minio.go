package bootstrap

import (
	"context"
	"log"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func SetupMinIO() {

	// MinIO服务器的URL和端口
	endpoint := "localhost:9000"
	accessKeyID := "root"
	secretAccessKey := "12345678"
	useSSL := false

	// 创建一个MinIO客户端
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalln(err)
	}
	// 创建一个上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 创建存储桶
	bucketName := "my-bucket"
	location := "us-east-1"

	err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
	if err != nil {
		// 检查存储桶是否已经存在
		exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			log.Printf("Bucket %s already exists\n", bucketName)
		} else {
			log.Fatalln(err)
		}
	} else {
		log.Printf("Successfully created %s\n", bucketName)
	}
	// 上传文件
	objectName := "1.txt"
	filePath := "./static/1.txt"
	contentType := "application/text"

	// 上传文件
	uploadInfo, err := minioClient.FPutObject(ctx, bucketName, objectName, filePath, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("Successfully uploaded %s of size %d\n", objectName, uploadInfo.Size)

}
