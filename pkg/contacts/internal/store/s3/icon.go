package s3

import (
	"bytes"
	"context"
	"image"
	"image/png"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rs/zerolog/log"
)

type IconStore struct {
	client *s3.Client
}

func NewIconStore(endpoint string) (*IconStore, error) {
	region := "us-east-1"
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(region),
		config.WithEndpointResolver(
			aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
				return aws.Endpoint{URL: endpoint, HostnameImmutable: true}, nil
			}),
		),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider("minioadmin", "minioadmin", ""),
		),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg)

	return &IconStore{client: client}, nil
}

func (store *IconStore) UploadIcon(ctx context.Context, name string, img image.Image) error {
	filename := name + ".png"

	data := new(bytes.Buffer)
	err := png.Encode(data, img)
	if err != nil {
		return err
	}

	uploader := manager.NewUploader(store.client)
	_, err = uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String("profile-icons"),
		Key:         aws.String(filename),
		Body:        data,
		ContentType: aws.String("image/png"),
	})
	if err != nil {
		return err
	}
	log.Info().Str("filename", filename).Msg("uploaded icon to s3")
	return nil
}
