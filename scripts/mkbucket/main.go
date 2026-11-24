package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func main() {
	client := s3.New(s3.Options{
		Region:       "us-east-1",
		Credentials:  credentials.NewStaticCredentialsProvider("admin", "infra123!", ""),
		BaseEndpoint: aws.String("http://127.0.0.1:9000"),
		UsePathStyle: true,
	})
	_, err := client.CreateBucket(context.Background(), &s3.CreateBucketInput{
		Bucket: aws.String("light-gozero"),
	})
	if err != nil {
		fmt.Println("create:", err)
	} else {
		fmt.Println("bucket ok")
	}
	out, err := client.ListBuckets(context.Background(), &s3.ListBucketsInput{})
	if err != nil {
		log.Fatal(err)
	}
	for _, b := range out.Buckets {
		fmt.Println("-", aws.ToString(b.Name))
	}
}
