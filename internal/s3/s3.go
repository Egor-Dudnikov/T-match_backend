// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

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

func LoadS3(cfg models.S3Config) (*minio.Client, error) {
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	return minioClient, err
}

type S3Storage struct {
	client *minio.Client
	cfg    models.S3Config
}

func NewS3(s3client *minio.Client, cfg models.S3Config) (*S3Storage, error) {
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
	return &S3Storage{client: s3client, cfg: cfg}, nil
}

func (s3 S3Storage) SetFile(ctx context.Context, objectName string, file io.Reader, contentType string, fileHandler *multipart.FileHeader) (string, error) {
	info, err := s3.client.PutObject(ctx, constants.BucketName, objectName, file, fileHandler.Size, minio.PutObjectOptions{ContentType: contentType})
	url := s3.GetURL(info)
	return url, err
}

func (s3 S3Storage) Delete(ctx context.Context, objectName string) error {
	err := s3.client.RemoveObject(ctx, constants.BucketName, objectName, minio.RemoveObjectOptions{})
	return err
}

func (s3 S3Storage) GetURL(info minio.UploadInfo) string {
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
