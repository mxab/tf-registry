package s3moduleservice

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/mxab/tf-registry/internal/module/service"
	"github.com/stretchr/testify/assert"

	"github.com/testcontainers/testcontainers-go"
)

func startMinio(t *testing.T) (func(), *s3.Client, string, *s3.PresignClient) {
	t.Helper()
	ctx := context.Background()

	minioUser := "minio"
	minioPassword := "minio123"
	bucketName := "test"
	req := testcontainers.ContainerRequest{
		Image: "minio/minio:latest",

		Env: map[string]string{
			"MINIO_ROOT_USER":     minioUser,
			"MINIO_ROOT_PASSWORD": minioPassword,
		},
		Cmd: []string{"server", "/data", "--console-address", ":9001"},

		ExposedPorts: []string{"9000/tcp", "9001/tcp"},
		//WaitingFor:   wait.ForHTTP("/").WithPort("9001/tcp"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatal(err)
	}
	// create s3 client from minio container
	ip, err := container.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}
	port, err := container.MappedPort(ctx, "9000/tcp")
	if err != nil {
		t.Fatal(err)
	}
	const defaultRegion = "us-east-1"
	hostAddress := fmt.Sprintf("http://%s:%s", ip, port.Port())

	resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...any) (aws.Endpoint, error) {
		return aws.Endpoint{
			PartitionID:       "aws",
			URL:               hostAddress,
			SigningRegion:     defaultRegion,
			HostnameImmutable: true,
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(defaultRegion),
		config.WithEndpointResolverWithOptions(resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(minioUser, minioPassword, "")),
	)
	if err != nil {
		t.Fatal(err)
	}

	s3Client := s3.NewFromConfig(cfg)
	presignClient := s3.NewPresignClient(s3Client)

	_, err = s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		t.Fatal(err)
	}
	return func() {
		container.Terminate(ctx)
	}, s3Client, bucketName, presignClient
}

// test DownloadUrl function
// create a minio testcontainer, upload a file for namespace/name/system/version, then call the DownloadUrl function and check if it returns a signed s3 url to that path

func TestDownloadUrl(t *testing.T) {
	//skip if short
	if testing.Short() {
		t.Skip("skipping integration test in short mode.")
	}

	cleanup, s3Client, bucketName, presignClient := startMinio(t)
	defer cleanup()

	ctx := context.Background()

	uploadArtifact(t, s3Client, ctx, bucketName, "hashicorp", "aws", "aws", "3.0.0")

	s3Service := NewS3ModuleService(s3Client, bucketName, presignClient)
	result, err := s3Service.GetModuleDownloadUrl(service.ModuleDescriptor{
		Namespace: "hashicorp",
		Name:      "aws",
		System:    "aws",
	}, "3.0.0")

	assert.NoError(t, err)
	url, err := url.Parse(result)
	assert.NoError(t, err)
	//check for signed s3 url params
	assert.Equal(t, "AWS4-HMAC-SHA256", url.Query().Get("X-Amz-Algorithm"))
	assert.Equal(t, "minio/20221219/us-east-1/s3/aws4_request", url.Query().Get("X-Amz-Credential"))
	assert.Equal(t, "900", url.Query().Get("X-Amz-Expires"))
	assert.NotEqual(t, "", url.Query().Get("X-Amz-Signature"))
	assert.Equal(t, "host", url.Query().Get("X-Amz-SignedHeaders"))

	assert.NotEqual(t, "", result)

}

func TestListModuleVersions(t *testing.T) {
	//skip if short
	if testing.Short() {
		t.Skip("skipping integration test in short mode.")
	}

	cleanup, s3Client, bucketName, _ := startMinio(t)
	defer cleanup()

	ctx := context.Background()

	uploadArtifact(t, s3Client, ctx, bucketName, "hashicorp", "aws", "aws", "3.0.0")
	uploadArtifact(t, s3Client, ctx, bucketName, "hashicorp", "aws", "aws", "3.0.1")
	uploadArtifact(t, s3Client, ctx, bucketName, "hashicorp", "aws", "aws", "3.0.2")

	s3Service := NewS3ModuleService(s3Client, bucketName, nil)
	result, err := s3Service.Versions(service.ModuleDescriptor{
		Namespace: "hashicorp",
		Name:      "aws",
		System:    "aws",
	})

	assert.NoError(t, err)
	assert.Equal(t, []string{"3.0.0", "3.0.1", "3.0.2"}, result)
}

func uploadArtifact(t *testing.T, s3Client *s3.Client, ctx context.Context, bucketName, namespace, name, system, version string) {
	t.Helper()

	_, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fmt.Sprintf("modules/namespaces/%s/%s/%s/%s/module.zip", namespace, name, system, version)),
		Body:   bytes.NewReader([]byte("module data")),
	})

	if err != nil {
		t.Fatal(err)
	}

}

func TestUpload(t *testing.T) {
	//skip if short
	if testing.Short() {
		t.Skip("skipping integration test in short mode.")
	}

	cleanup, s3Client, bucketName, _ := startMinio(t)
	defer cleanup()

	ctx := context.Background()

	s3Service := NewS3ModuleService(s3Client, bucketName, nil)
	err := s3Service.UploadModule(service.ModuleDescriptor{
		Namespace: "hashicorp",
		Name:      "aws",
		System:    "aws",
	}, "3.0.0", bytes.NewReader([]byte("module data")))
	assert.NoError(t, err)

	_, err = s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String("modules/namespaces/hashicorp/aws/aws/3.0.0/module.zip"),
	})
	assert.NoError(t, err)
}
