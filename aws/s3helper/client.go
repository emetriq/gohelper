package s3helper

import (
	"bytes"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/emetriq/gohelper/aws/ec2helper"
	"github.com/emetriq/gohelper/env"
)

// S3Client ...
type Client struct {
	Client  *s3.S3
	Session *session.Session
}

// NewS3Client ... ctor
// region is the region of the bucket you want to access
// leave it blank to use AWS_REGION environment variable
func NewS3Client(region string) *Client {
	if region == "" {
		region = env.GetStrEnv("AWS_REGION", "eu-west-1")
	}

	sess, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	if env.GetStrEnv("AWS_SDK_LOAD_CONFIG", "") == "" {
		cred, err := ec2helper.GetCredentialsFromRole(sess)
		if err != nil {
			panic(err)
		}
		sess.Config.Credentials = cred
	}

	if err != nil {
		panic(err)
	}

	client := s3.New(sess)

	s3 := Client{
		Client:  client,
		Session: sess,
	}
	return &s3
}

// NewS3ClientWithSession ... ctor
func NewS3ClientWithSession(sess *session.Session) *Client {

	if sess == nil {
		return nil
	}

	client := s3.New(sess)

	s3 := Client{
		Client:  client,
		Session: sess,
	}
	return &s3
}

func contentToSlice(contents []*s3.Object) []string {
	result := make([]string, 0, len(contents))
	for _, key := range contents {
		result = append(result, *key.Key)
	}
	return result
}

// ListBucketDir list all files in a bucket and supports a prefix
// if you want to list file of s3://mybucket/20111117/data you must call
// ListBucketDir("mybucket", "20111117/data")
func (c Client) ListBucketDir(bucket, key string) ([]string, error) {
	resp, err := c.Client.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String(key),
	})

	if err != nil {
		return nil, err
	}
	result := contentToSlice(resp.Contents)

	return result, err
}

// ListBucket list all files in a bucket
func (c Client) ListBucket(bucket string) ([]string, error) {
	resp, err := c.Client.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(bucket),
	})

	if err != nil {
		return nil, err
	}
	result := contentToSlice(resp.Contents)
	return result, err
}

// GetObject returns the content of a file as bytes
func (c Client) GetBytes(bucket string, key string) ([]byte, error) {

	requestInput := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	result, err := c.Client.GetObject(requestInput)
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()
	body, err := ioutil.ReadAll(result.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// GetObject uploads bytes to given bucket and key
func (c Client) PutBytes(bucket string, key string, contentType string, body []byte) error {
	requestInput := &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(body),
		ContentType: aws.String(contentType),
	}

	_, err := c.Client.PutObject(requestInput)
	if err != nil {
		return err
	}
	return nil
}
