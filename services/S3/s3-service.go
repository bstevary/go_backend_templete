package S3

import (
	"context" // Required for config.LoadDefaultAWSConfig
	"fmt"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

// S3Uploader defines the interface for our S3 file upload service.
type S3Uploader interface {
	UploadFiles(ctx context.Context, files []*multipart.FileHeader, rootPath string) ([]string, error)
	UploadFile(ctx context.Context, file *multipart.FileHeader, rootPath string) (string, error)
	UploadFileUpdate(ctx context.Context, file *multipart.FileHeader, key string) (string, error)
	GeneratePresignedURL(ctx context.Context, s3Key string, expiry time.Duration) (string, error)
	DeleteFile(ctx context.Context, s3Key string) error
	DeleteFiles(ctx context.Context, s3Keys []string) error
}

// S3Service implements the Uploader interface.
type S3Service struct {
	s3Client        *s3.Client
	s3BucketName    string
	uploader        *manager.Uploader
	presigner       *s3.PresignClient
	fileStoragePath string
}

// NewS3Service creates a new S3Service instance, handling the S3 client initialization internally.
// It takes the AWS region and S3 bucket name as arguments.
func NewS3Service(awsRegion, s3BucketName string, fileStoragePath string) (S3Uploader, error) {
	if awsRegion == "" || s3BucketName == "" || fileStoragePath == "" {
		return nil, fmt.Errorf("AWS region, S3 bucket name, or file storage path is empty")
	}
	sdkConfig, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(awsRegion),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client
	s3Client := s3.NewFromConfig(sdkConfig)

	return &S3Service{
		s3Client:     s3Client,
		s3BucketName: s3BucketName,
		uploader:     manager.NewUploader(s3Client),
		presigner:    s3.NewPresignClient(s3Client),
	}, nil
}

// UploadFile uploads a single file to S3 and returns its S3 Key.
func (s *S3Service) UploadFile(ctx context.Context, file *multipart.FileHeader, rootPath string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	fileName := uuid.New().String() + filepath.Ext(file.Filename)
	s3Key := fmt.Sprintf("%s/%s/%s", s.fileStoragePath, rootPath, fileName)

	_, err = s.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.s3BucketName),
		Key:         aws.String(s3Key),
		Body:        src,
		ContentType: aws.String(file.Header.Get("Content-Type")),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	return s3Key, nil
}
func (s *S3Service) UploadFileUpdate(ctx context.Context, file *multipart.FileHeader, key string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	_, err = s.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.s3BucketName),
		Key:    aws.String(key),
		Body:   src,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file update to S3: %w", err)
	}

	return key, nil
}

// UploadFiles uploads multiple files to S3 and returns a map of original filenames to their S3 Keys.
func (s *S3Service) UploadFiles(ctx context.Context, files []*multipart.FileHeader, rootPath string) ([]string, error) {
	var uploadedS3Keys []string
	for _, file := range files {
		s3Key, err := s.UploadFile(ctx, file, rootPath)
		if err != nil {
			return nil, fmt.Errorf("failed to upload file %s: %w", file.Filename, err)
		}
		uploadedS3Keys = append(uploadedS3Keys, s3Key)
	}
	return uploadedS3Keys, nil
}

// GeneratePresignedURL generates a time-limited presigned URL for a private S3 object.
func (s *S3Service) GeneratePresignedURL(ctx context.Context, s3Key string, expiry time.Duration) (string, error) {
	request, err := s.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.s3BucketName),
		Key:    aws.String(s3Key),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("failed to presign URL: %w", err)
	}
	return request.URL, nil
}

// deleteFile deletes a file from S3.
func (s *S3Service) DeleteFile(ctx context.Context, s3Key string) error {
	_, err := s.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.s3BucketName),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}
	return nil
}

// DeleteFiles deletes multiple files from S3.
func (s *S3Service) DeleteFiles(ctx context.Context, s3Keys []string) error {
	for _, s3Key := range s3Keys {
		if err := s.DeleteFile(ctx, s3Key); err != nil {
			return err
		}
	}
	return nil
}
