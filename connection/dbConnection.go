package connection

import (
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
		// Create a new AWS session with the region and credentials from environment variables
		sess, err := session.NewSession(&aws.Config{
			Region:      aws.String(os.Getenv("AWS_REGION")),
			Credentials: credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), ""),
		})
		if err != nil {
			log.Fatalf("Failed to create AWS session: %v", err)
		}

		// Create DynamoDB client
		svc = dynamodb.New(sess)
	}

	return svc
}
