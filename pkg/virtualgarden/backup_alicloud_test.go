package virtualgarden

import (
	"fmt"
	"os"
	"testing"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

func TestAliCloudBackup(t *testing.T) {
	// init
	// set the id and key to activate the test
	accessKeyId := ""
	accessSecretKey := ""

	if accessKeyId == "" || accessSecretKey == "" {
		return
	}

	endpoint := "oss-eu-central-1.aliyuncs.com"

	// Create an OSSClient instance.
	client, err := oss.New(endpoint, accessKeyId, accessSecretKey)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}

	listAliCloudBuckets(client)

	bucketName := "testbucket-virtual-garden-5514364316"

	// Create a bucket (the default storage class is Standard) and set the ACL of the bucket to public read (the default ACL is private).
	err = client.CreateBucket(bucketName, oss.ACL(oss.ACLPrivate))
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}

	listAliCloudBuckets(client)

	err = deleteAliCloudBucket(client, bucketName)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}

	listAliCloudBuckets(client)
}

func listAliCloudBuckets(client *oss.Client) {
	fmt.Println("Listing all buckets: ")

	marker := ""
	for {
		lsRes, err := client.ListBuckets(oss.Marker(marker))
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(-1)
		}

		// By default, 100 buckets are listed each time.
		for _, bucket := range lsRes.Buckets {
			fmt.Println("Bucket: ", bucket.Name)
		}

		if lsRes.IsTruncated {
			marker = lsRes.NextMarker
		} else {
			break
		}
	}
}

func deleteAliCloudBucket(client *oss.Client, bucketName string) error {
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	fmt.Printf("Deleting objects of bucket %s\n", bucketName)
	marker := ""
	for {
		lsRes, err := bucket.ListObjects(oss.Marker(marker))
		if err != nil {
			if aerr, ok := err.(oss.ServiceError); ok {
				switch aerr.Code {
				case "NoSuchBucket":
					fmt.Println("Bucket does not exist")
					return nil
				default:
					return err
				}
			} else {
				return err
			}
		}

		for _, object := range lsRes.Objects {
			fmt.Printf("Deleting object %s of bucket %s\n", object.Key, bucketName)

			err = bucket.DeleteObject(object.Key)
			if err != nil {
				fmt.Println("Error:", err)
				return err
			}
		}

		if lsRes.IsTruncated {
			marker = lsRes.NextMarker
		} else {
			break
		}
	}

	fmt.Printf("Deleting bucket %s\n", bucketName)
	err = client.DeleteBucket(bucketName)
	if err != nil {
		if aerr, ok := err.(oss.ServiceError); ok {
			switch aerr.Code {
			case "NoSuchBucket":
				fmt.Println("Bucket does not exist")
				return nil
			default:
				return err
			}
		} else {
			return err
		}
	}

	return nil
}
