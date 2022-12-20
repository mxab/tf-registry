package s3moduleservice

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/mxab/tf-registry/internal/module/service"
)

type S3ModuleService struct {
	s3            *s3.Client
	presignClient *s3.PresignClient
	bucketName    string
}

func buildS3Key(module service.ModuleDescriptor, version string) string {

	return fmt.Sprintf("modules/namespaces/%s/%s/%s/%s/module.tar.gz", module.Namespace, module.Name, module.System, version)
}

// implement the interface
func (s *S3ModuleService) List(req *service.ListParams) (*service.ModuleResult, error) {
	panic("implement me")
}
func (s *S3ModuleService) Seach(req *service.SearchParams) (*service.ModuleResult, error) {
	panic("implement me")
}
func (s *S3ModuleService) Versions(modul service.ModuleDescriptor) ([]string, error) {
	ctx := context.Background()
	baseKey := fmt.Sprintf("modules/namespaces/%s/%s/%s/", modul.Namespace, modul.Name, modul.System)
	resp, err := s.s3.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucketName),
		Prefix: aws.String(baseKey),
	})
	if err != nil {
		return nil, err
	}

	versions := make([]string, 0, len(resp.Contents))
	for _, obj := range resp.Contents {
		key := strings.TrimSuffix(strings.TrimPrefix(*obj.Key, baseKey), "/module.tar.gz")
		versions = append(versions, key)
	}
	return versions, nil
}
func (s *S3ModuleService) GetModuleDownloadUrl(modul service.ModuleDescriptor, version string) (string, error) {

	ctx := context.Background()
	req, err := s.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(buildS3Key(modul, version)),
	})
	if err != nil {
		return "", err
	}

	return req.URL, nil
}

// implment upload
func (s *S3ModuleService) UploadModule(modul service.ModuleDescriptor, version string, content io.Reader) error {
	ctx := context.Background()
	_, err := s.s3.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(buildS3Key(modul, version)),
		Body:   content,
	})
	return err
}

func NewS3ModuleService(s3 *s3.Client, bucketName string, presignClient *s3.PresignClient) *S3ModuleService {

	return &S3ModuleService{s3: s3, bucketName: bucketName, presignClient: presignClient}
}
