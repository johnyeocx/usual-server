package cloud

import (
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

	// _ := "https://bool-m1.s3.ap-southeast-1.amazonaws.com/events/" + eventId + "/" + photoId

	return str, nil
}