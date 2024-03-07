package aws_service

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const BUCKET_NAME = "dsekt-tessera"

func NewS3Client() (*s3.S3, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-north-1"),
		Credentials: credentials.NewStaticCredentials(
			os.Getenv("AWS_ACCESS_KEY_ID"),
			os.Getenv("AWS_SECRET_ACCESS_KEY"),
			"",
		),
	})
	if err != nil {
		fmt.Println("Error creating session ", err)
		return nil, err
	}

	s3Client := s3.New(sess)
	return s3Client, nil
}

func UploadFileToS3(s3Client *s3.S3, key, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(BUCKET_NAME),
		Key:    aws.String(key),
		Body:   file,
	})
	return err
}

func DownloadFileFromS3(s3Client *s3.S3, key, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	output, err := s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(BUCKET_NAME),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}

	defer output.Body.Close()
	_, err = io.Copy(file, output.Body)
	return err
}

func GetFileURL(s3Client *s3.S3, key string) (string, error) {
	req, _ := s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(BUCKET_NAME),
		Key:    aws.String(key),
	})

	urlStr, err := req.Presign(24 * 7 * time.Hour) // Presign for 7 days

	if err != nil {
		return "", err
	}

	return urlStr, nil
}
