package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOStorage struct {
	Client         *minio.Client
	SignerClient   *minio.Client // Optional client for pre-signing with a different host
	BucketName     string
	PublicEndpoint string
}

func NewMinIOStorage(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*MinIOStorage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
		Region: "us-east-1",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	storage := &MinIOStorage{
		Client:         client,
		SignerClient:   client, // Default to the same client
		BucketName:     bucket,
		PublicEndpoint: endpoint,
	}

	return storage, nil
}

func (s *MinIOStorage) SetPublicEndpoint(endpoint, accessKey, secretKey string, useSSL bool) error {
	s.PublicEndpoint = endpoint
	// Create a dedicated signer client that doesn't necessarily need to reach the server from here,
	// but is used to generate correctly signed URLs for the browser.
	signer, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
		Region: "us-east-1",
	})
	if err != nil {
		return fmt.Errorf("failed to create signer minio client: %w", err)
	}
	s.SignerClient = signer
	return nil
}

func (s *MinIOStorage) Upload(objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	info, err := s.Client.PutObject(context.Background(), s.BucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload object: %w", err)
	}
	return info.Key, nil
}

func (s *MinIOStorage) GetPresignedURL(objectName string, expires time.Duration, download bool) (string, error) {
	reqParams := make(url.Values)
	if download {
		reqParams.Set("response-content-disposition", fmt.Sprintf("attachment; filename=\"%s\"", objectName))
	}

	// Use the SignerClient which ensures the signature is calculated for the correct Host
	presignedURL, err := s.SignerClient.PresignedGetObject(context.Background(), s.BucketName, objectName, expires, reqParams)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return presignedURL.String(), nil
}
