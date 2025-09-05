package main

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/peervault/peervault/internal/plugins"
)

// S3StoragePlugin implements the StoragePlugin interface for Amazon S3
type S3StoragePlugin struct {
	bucket     string
	region     string
	prefix     string
	session    *session.Session
	s3Client   *s3.S3
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
}

// NewS3StoragePlugin creates a new S3 storage plugin
func NewS3StoragePlugin() *S3StoragePlugin {
	return &S3StoragePlugin{}
}

// Plugin interface implementation
func (p *S3StoragePlugin) Name() string {
	return "s3-storage"
}

func (p *S3StoragePlugin) Version() string {
	return "1.0.0"
}

func (p *S3StoragePlugin) Description() string {
	return "Amazon S3 storage backend for PeerVault"
}

func (p *S3StoragePlugin) Type() plugins.PluginType {
	return plugins.PluginTypeStorage
}

func (p *S3StoragePlugin) Initialize(config map[string]interface{}) error {
	// Extract configuration
	var ok bool
	p.bucket, ok = config["bucket"].(string)
	if !ok {
		return fmt.Errorf("bucket configuration is required")
	}

	p.region, ok = config["region"].(string)
	if !ok {
		return fmt.Errorf("region configuration is required")
	}

	if prefix, exists := config["prefix"]; exists {
		p.prefix, _ = prefix.(string)
	}

	// Create AWS session
	awsConfig := &aws.Config{
		Region: aws.String(p.region),
	}

	// Add credentials if provided
	if accessKey, exists := config["access_key"]; exists {
		if secretKey, exists := config["secret_key"]; exists {
			awsConfig.Credentials = credentials.NewStaticCredentials(
				accessKey.(string),
				secretKey.(string),
				"", // token
			)
		}
	}

	// Add endpoint if provided (for S3-compatible services)
	if endpoint, exists := config["endpoint"]; exists {
		awsConfig.Endpoint = aws.String(endpoint.(string))
		awsConfig.S3ForcePathStyle = aws.Bool(true)
	}

	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return fmt.Errorf("failed to create AWS session: %w", err)
	}

	p.session = sess
	p.s3Client = s3.New(sess)
	p.uploader = s3manager.NewUploader(sess)
	p.downloader = s3manager.NewDownloader(sess)

	return nil
}

func (p *S3StoragePlugin) Start() error {
	// Test connection to S3
	_, err := p.s3Client.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(p.bucket),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to S3 bucket %s: %w", p.bucket, err)
	}

	return nil
}

func (p *S3StoragePlugin) Stop() error {
	// Cleanup resources if needed
	return nil
}

func (p *S3StoragePlugin) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"bucket": map[string]interface{}{
			"type":        "string",
			"required":    true,
			"description": "S3 bucket name",
		},
		"region": map[string]interface{}{
			"type":        "string",
			"required":    true,
			"description": "AWS region",
		},
		"prefix": map[string]interface{}{
			"type":        "string",
			"required":    false,
			"description": "Key prefix for all objects",
		},
		"access_key": map[string]interface{}{
			"type":        "string",
			"required":    false,
			"description": "AWS access key ID",
		},
		"secret_key": map[string]interface{}{
			"type":        "string",
			"required":    false,
			"description": "AWS secret access key",
		},
		"endpoint": map[string]interface{}{
			"type":        "string",
			"required":    false,
			"description": "S3-compatible endpoint URL",
		},
	}
}

func (p *S3StoragePlugin) ValidateConfig(config map[string]interface{}) error {
	// Validate required fields
	if _, exists := config["bucket"]; !exists {
		return fmt.Errorf("bucket is required")
	}
	if _, exists := config["region"]; !exists {
		return fmt.Errorf("region is required")
	}

	// Validate types
	if bucket, exists := config["bucket"]; exists {
		if _, ok := bucket.(string); !ok {
			return fmt.Errorf("bucket must be a string")
		}
	}
	if region, exists := config["region"]; exists {
		if _, ok := region.(string); !ok {
			return fmt.Errorf("region must be a string")
		}
	}

	return nil
}

// StoragePlugin interface implementation
func (p *S3StoragePlugin) Store(ctx context.Context, key string, data io.Reader) error {
	fullKey := p.getFullKey(key)

	_, err := p.uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(fullKey),
		Body:   data,
	})

	if err != nil {
		return fmt.Errorf("failed to store object %s: %w", key, err)
	}

	return nil
}

func (p *S3StoragePlugin) Retrieve(ctx context.Context, key string) (io.ReadCloser, error) {
	fullKey := p.getFullKey(key)

	result, err := p.s3Client.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(fullKey),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve object %s: %w", key, err)
	}

	return result.Body, nil
}

func (p *S3StoragePlugin) Delete(ctx context.Context, key string) error {
	fullKey := p.getFullKey(key)

	_, err := p.s3Client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(fullKey),
	})

	if err != nil {
		return fmt.Errorf("failed to delete object %s: %w", key, err)
	}

	return nil
}

func (p *S3StoragePlugin) List(ctx context.Context, prefix string) ([]string, error) {
	fullPrefix := p.getFullKey(prefix)

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(p.bucket),
		Prefix: aws.String(fullPrefix),
	}

	var keys []string
	err := p.s3Client.ListObjectsV2PagesWithContext(ctx, input, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, obj := range page.Contents {
			key := *obj.Key
			if p.prefix != "" {
				key = strings.TrimPrefix(key, p.prefix+"/")
			}
			keys = append(keys, key)
		}
		return true
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	return keys, nil
}

func (p *S3StoragePlugin) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := p.getFullKey(key)

	_, err := p.s3Client.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(fullKey),
	})

	if err != nil {
		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "NoSuchKey") || strings.Contains(err.Error(), "NotFound") {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if object exists %s: %w", key, err)
	}

	return true, nil
}

func (p *S3StoragePlugin) GetMetadata(ctx context.Context, key string) (*plugins.FileMetadata, error) {
	fullKey := p.getFullKey(key)

	result, err := p.s3Client.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(fullKey),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get metadata for object %s: %w", key, err)
	}

	metadata := &plugins.FileMetadata{
		Key:         key,
		Size:        *result.ContentLength,
		ContentType: aws.StringValue(result.ContentType),
		CreatedAt:   result.LastModified.Format(time.RFC3339),
		UpdatedAt:   result.LastModified.Format(time.RFC3339),
		Checksum:    aws.StringValue(result.ETag),
		Metadata:    make(map[string]string),
	}

	// Add custom metadata
	for key, value := range result.Metadata {
		metadata.Metadata[key] = aws.StringValue(value)
	}

	return metadata, nil
}

// Helper methods
func (p *S3StoragePlugin) getFullKey(key string) string {
	if p.prefix == "" {
		return key
	}
	return p.prefix + "/" + key
}

// Register the plugin
func init() {
	plugins.RegisterStoragePlugin(NewS3StoragePlugin())
}
