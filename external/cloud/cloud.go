package cloud

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var (
	AccessKeyID     string
	SecretAccessKey string
	MyRegion        string
	// filepath        string
	// s3session       *s3.S3
	s3Endpoint string = "https://usual.s3.eu-west-2.amazonaws.com"
)

func ConnectAWS() (*session.Session) {
	AccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	SecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	MyRegion = os.Getenv("AWS_REGION")

	sess, err := session.NewSession(
		&aws.Config{
			Region: aws.String(MyRegion),
			Credentials: credentials.NewStaticCredentials(
				AccessKeyID,
				SecretAccessKey,
				"",
			),
		},
	)

	if err != nil {
		panic(err)
	}

	return sess

}


func GetImageUploadUrl(sess *session.Session, key string) (string, error) {
	svc := s3.New(sess)
	var bucket = os.Getenv("BUCKET_NAME")

	req, _ := svc.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	str, err := req.Presign(60 * time.Minute)
	if err != nil {
		return "can't sign", err
	}

	return str, nil
}

func GetObjectPresignedURL(sess *session.Session, key string, time time.Duration) (string, error) {
	svc := s3.New(sess)
	var bucket = os.Getenv("BUCKET_NAME")

	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	str, err := req.Presign(time)
	if err != nil {
		return "can't sign", err
	}

	return str, nil
}

func PutObject(
	sess *session.Session, 
	image []byte, 
	contentType string,
	key string,
) (error) {
	svc := s3.New(sess)
	var bucket = os.Getenv("BUCKET_NAME")	

	_, err := svc.PutObject(&s3.PutObjectInput{
		Body: bytes.NewReader(image),
		ContentType: &contentType,
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	return err
}

func DeleteImage(sess *session.Session, key string) (error) {
	svc := s3.New(sess)
	var bucket = os.Getenv("BUCKET_NAME")

	_, err := svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		log.Println(err)
	} else {
		fmt.Println("Successfully deleted image")
	}
	
	return err
}