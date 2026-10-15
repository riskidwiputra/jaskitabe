package ulasan

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type R2Storage struct {
	client    *s3.Client
	bucket    string
	publicURL string
}

func NewR2StorageFromEnv(ctx context.Context) (*R2Storage, error) {
	accountID := strings.TrimSpace(os.Getenv("R2_ACCOUNT_ID"))
	accessKeyID := strings.TrimSpace(os.Getenv("R2_ACCESS_KEY_ID"))
	secretAccessKey := strings.TrimSpace(os.Getenv("R2_SECRET_ACCESS_KEY"))
	bucket := strings.TrimSpace(os.Getenv("R2_BUCKET_NAME"))
	publicURL := strings.TrimRight(strings.TrimSpace(os.Getenv("R2_PUBLIC_URL")), "/")
	for name, value := range map[string]string{
		"R2_ACCOUNT_ID":        accountID,
		"R2_ACCESS_KEY_ID":     accessKeyID,
		"R2_SECRET_ACCESS_KEY": secretAccessKey,
		"R2_BUCKET_NAME":       bucket,
		"R2_PUBLIC_URL":        publicURL,
	} {
		if value == "" {
			return nil, fmt.Errorf("%s belum dikonfigurasi", name)
		}
	}
	parsedPublicURL, err := url.Parse(publicURL)
	if err != nil || (parsedPublicURL.Scheme != "https" && parsedPublicURL.Scheme != "http") || parsedPublicURL.Host == "" || parsedPublicURL.RawQuery != "" || parsedPublicURL.Fragment != "" {
		return nil, errors.New("R2_PUBLIC_URL harus berupa URL HTTP(S) publik yang valid tanpa query atau fragment")
	}

	cfg, err := awsconfig.LoadDefaultConfig(
		ctx,
		awsconfig.WithRegion("auto"),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("gagal menyiapkan konfigurasi R2: %w", err)
	}
	client := s3.NewFromConfig(cfg, func(options *s3.Options) {
		options.BaseEndpoint = aws.String("https://" + accountID + ".r2.cloudflarestorage.com")
		options.UsePathStyle = true
	})
	return &R2Storage{client: client, bucket: bucket, publicURL: publicURL}, nil
}

func (s *R2Storage) Upload(ctx context.Context, key string, body io.Reader, contentType string) (string, error) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &s.bucket,
		Key:         &key,
		Body:        body,
		ContentType: &contentType,
	})
	if err != nil {
		return "", fmt.Errorf("gagal mengunggah foto ke R2: %w", err)
	}
	return s.publicURL + "/" + key, nil
}

func (s *R2Storage) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: &s.bucket, Key: &key})
	if err != nil {
		return fmt.Errorf("gagal menghapus foto dari R2: %w", err)
	}
	return nil
}

func (s *R2Storage) ObjectKey(photoURL string) (string, bool) {
	prefix := s.publicURL + "/"
	if !strings.HasPrefix(photoURL, prefix) {
		return "", false
	}
	key := strings.TrimPrefix(photoURL, prefix)
	if key == "" || strings.Contains(key, "..") || strings.Contains(key, "\\") || strings.HasPrefix(key, "/") {
		return "", false
	}
	return key, true
}
