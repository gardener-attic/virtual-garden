package virtualgarden

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	dummy              = "..."
	awsAccessKeyId     = "AWS_ACCESS_KEY_ID"
	awsSecretAccessKey = "AWS_SECRET_ACCESS_KEY"
)

func TestAwsBackup(t *testing.T) {
	// init
	// set the id and key to activate the test
	os.Setenv(awsAccessKeyId, dummy)
	os.Setenv(awsSecretAccessKey, dummy)

	if os.Getenv(awsAccessKeyId) == dummy || os.Getenv(awsSecretAccessKey) == dummy {
		return
	}

	region := "eu-central-1"

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)

	if err != nil {
		exitErrorf("Unable to get session, %v", err)
	}

	svc := s3.New(sess)

	// list buckets
	listBuckets(svc)

	// create bucket
	fmt.Println("*** Create Bucket ***")

	bucket := "luggbdq6de"
	_, err = svc.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeBucketAlreadyOwnedByYou:
				fmt.Println(s3.ErrCodeBucketAlreadyOwnedByYou, aerr.Error())
			default:
				exitErrorf("Unable to create bucket %q, %v", bucket, err)
			}
		} else {
			exitErrorf("Unable to create bucket %q, %v", bucket, err)
		}
	} else {
		// Wait until bucket is created before finishing
		fmt.Printf("Waiting for bucket %q to be created...\n", bucket)

		err = svc.WaitUntilBucketExists(&s3.HeadBucketInput{
			Bucket: aws.String(bucket),
		})

		if err != nil {
			exitErrorf("Unable to wait for bucket %q ready, %v", bucket, err)
		}
	}

	listBuckets(svc)

	// delete bucket
	fmt.Println("*** Create Bucket ***")
	_, err = svc.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(bucket),
	})

	if err != nil {
		exitErrorf("Unable to delete bucket %q, %v", bucket, err)
	}

	// Wait until bucket is deleted before finishing
	fmt.Printf("Waiting for bucket %q to be deleted...\n", bucket)

	err = svc.WaitUntilBucketNotExists(&s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})

	if err != nil {
		exitErrorf("Unable to delete bucket %q, %v", bucket, err)
	}
}

func listBuckets(svc *s3.S3) {
	result, err := svc.ListBuckets(nil)
	if err != nil {
		exitErrorf("Unable to list buckets, %v", err)
	}

	fmt.Println("Buckets:")

	for _, b := range result.Buckets {
		fmt.Printf("* %s created on %s\n",
			aws.StringValue(b.Name), aws.TimeValue(b.CreationDate))
	}
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}
