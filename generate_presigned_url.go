package main

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func generatePresignedURL(s3Client *s3.Client, bucket, key string, expiredTime time.Duration) (string, error) {
	newPresignClient := s3.NewPresignClient(s3Client)
	httpReq, err := newPresignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{Bucket: &bucket, Key: &key}, s3.WithPresignExpires(expiredTime))
	if err != nil {
		log.Fatal("Could not generate http request")
		return "", err
	}
	return httpReq.URL, nil
}