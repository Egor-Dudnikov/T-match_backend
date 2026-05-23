// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package s3

import (
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

type S3Storge struct {
	client *minio.Client
	cfg    models.S3Config
}

func NewS3(s3client *minio.Client, cfg models.S3Config) (*S3Storge, error) {
	policy := `{
        "Version": "2012-10-17",
        "Statement": [
            {
                "Effect": "Allow",
                "Principal": {"AWS": "*"},
                "Action": ["s3:GetObject"],
                "Resource": ["arn:aws:s3:::` + "t-match-storge" + `/*"]
            }
        ]
    }`
	exist, err := s3client.BucketExists(context.Background(), "t-match-storge")
	if err != nil {
		return nil, err
	}
	if !exist {
		err := s3client.MakeBucket(context.Background(), "t-match-storge", minio.MakeBucketOptions{Region: "ru-central-1"})
		if err != nil {
			return nil, err
		}
	}

	err = s3client.SetBucketPolicy(context.Background(), "t-match-storge", policy)
	if err != nil {
		return nil, err
	}
	return &S3Storge{client: s3client, cfg: cfg}, nil
}

func (s3 S3Storge) SetFile(ctx context.Context, objectName string, file io.Reader, contentType string, fileHandler *multipart.FileHeader) (string, error) {
	info, err := s3.client.PutObject(ctx, "t-match-storge", objectName, file, fileHandler.Size, minio.PutObjectOptions{ContentType: contentType})
	url := s3.GetURL(info)
	return url, err
}

func (s3 S3Storge) Delete(ctx context.Context, objectName string) error {
	err := s3.client.RemoveObject(ctx, "t-match-storge", objectName, minio.RemoveObjectOptions{})
	return err
}

func (s3 S3Storge) GetURL(info minio.UploadInfo) string {
	url := "http"
	if s3.cfg.UseSSL {
		url = "https"
	}
	url += "://"
	url += os.Getenv("SERVER_HOST")
	url += "/"
	url += info.Bucket
	url += "/"
	url += info.Key
	return url
}
