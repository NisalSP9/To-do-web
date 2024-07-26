package connection

import (
	"encoding/base64"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

var svc *dynamodb.DynamoDB

func GetSVC() *dynamodb.DynamoDB {
	if svc == nil {
		// Decode base64 encoded AWS access key ID and secret access key
		encodedAccessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
		encodedSecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

		accessKeyID, err := base64.StdEncoding.DecodeString(encodedAccessKeyID)
		if err != nil {
			log.Fatalf("Failed to decode AWS access key ID: %v", err)
		}

		secretAccessKey, err := base64.StdEncoding.DecodeString(encodedSecretAccessKey)
		if err != nil {
			log.Fatalf("Failed to decode AWS secret access key: %v", err)
		}

		// Create a new AWS session with the region and decoded credentials
		sess, err := session.NewSession(&aws.Config{
			Region:      aws.String(os.Getenv("AWS_REGION")),
			Credentials: credentials.NewStaticCredentials(string(accessKeyID), string(secretAccessKey), ""),
		})
		if err != nil {
			log.Fatalf("Failed to create AWS session: %v", err)
		}

		// Create DynamoDB client
		svc = dynamodb.New(sess)
	}

	return svc
}
