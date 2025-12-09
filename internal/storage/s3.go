package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// UploadToS3 uploads the given filePath to the given bucket/region with the provided key.
func UploadToS3(bucket, region, key, filePath string) error {
	ctx := context.Background()

	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return fmt.Errorf("load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open file for S3 upload: %w", err)
	}
	defer f.Close()

	// if key empty, just use file name
	if key == "" {
		key = filepath.Base(filePath)
	}

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   f,
		ACL:    types.ObjectCannedACLPrivate,
	})
	if err != nil {
		return fmt.Errorf("put object to S3: %w", err)
	}

	return nil
}
