// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

// Package s3 provides a wrapper around the S3-compatible object storage.
package s3

import (
	"T-match_backend/internal/constants"
	"T-match_backend/internal/models"
	"context"
	"io"
	"mime/multipart"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// LoadS3 creates a MinIO client from the given config.
func LoadS3(cfg models.S3Config) (*minio.Client, error) {
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	return minioClient, err
}

// Storage is a wrapper around a MinIO client for object storage operations.
type Storage struct {
	client *minio.Client
	cfg    models.S3Config
}

// NewS3 ensures the bucket exists with public read policy and returns a Storage.
func NewS3(s3client *minio.Client, cfg models.S3Config) (*Storage, error) {
	policy := `{
        "Version": "2012-10-17",
        "Statement": [
            {
                "Effect": "Allow",
                "Principal": {"AWS": "*"},
                "Action": ["s3:GetObject"],
                "Resource": ["arn:aws:s3:::` + constants.BucketName + `/*"]
            }
        ]
    }`
	exist, err := s3client.BucketExists(context.Background(), constants.BucketName)
	if err != nil {
		return nil, err
	}
	if !exist {
		err := s3client.MakeBucket(context.Background(), constants.BucketName, minio.MakeBucketOptions{Region: "ru-central-1"})
		if err != nil {
			return nil, err
		}
	}

	err = s3client.SetBucketPolicy(context.Background(), constants.BucketName, policy)
	if err != nil {
		return nil, err
	}
	return &Storage{client: s3client, cfg: cfg}, nil
}

// SetFile uploads the given file to the bucket and returns its public URL.
func (s3 Storage) SetFile(ctx context.Context, objectName string, file io.Reader, contentType string, fileHandler *multipart.FileHeader) (string, error) {
	info, err := s3.client.PutObject(ctx, constants.BucketName, objectName, file, fileHandler.Size, minio.PutObjectOptions{ContentType: contentType})
	url := s3.GetURL(info)
	return url, err
}

// Delete removes the object with the given name from the bucket.
func (s3 Storage) Delete(ctx context.Context, objectName string) error {
	err := s3.client.RemoveObject(ctx, constants.BucketName, objectName, minio.RemoveObjectOptions{})
	return err
}

// GetURL builds the public URL for the given upload info.
func (s3 Storage) GetURL(info minio.UploadInfo) string {
	url := "http"
	if s3.cfg.UseSSL {
		url = "https"
	}
	url += "://"
	url += os.Getenv("S3_HOST")
	url += os.Getenv("S3_PORT")
	url += "/"
	url += info.Bucket
	url += "/"
	url += info.Key
	return url
}
