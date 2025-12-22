package avatar

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image/png"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/elug3/gochat/shared/events"
	"github.com/elug3/identicon"
)

type Handler struct {
	s3Client *s3.Client
	bucket   string
}

func NewHandler(s3Client *s3.Client, bucket string) (*Handler, error) {
	if s3Client == nil {
		return nil, fmt.Errorf("s3 client cannot be nil")
	}
	if bucket == "" {
		return nil, fmt.Errorf("bucket cannot be empty")
	}
	return &Handler{
		s3Client: s3Client,
		bucket:   bucket,
	}, nil
}

func (h *Handler) HandleProfileCreated(ctx context.Context, subject string, data []byte) error {
	var event events.ProfileCreated
	if err := json.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("unmarshal event data: %w", err)
	}
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(event.UserId))

	img := identicon.New(b, 256)
	key := fmt.Sprintf("%d.png", event.UserId)

	buf := new(bytes.Buffer)
	if err := png.Encode(buf, img); err != nil {
		return fmt.Errorf("encode png: %w", err)
	}

	uploader := manager.NewUploader(h.s3Client)
	_, err := uploader.Upload(context.Background(), &s3.PutObjectInput{
		Bucket:      aws.String(h.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String("image/png"),
	})
	if err != nil {
		return fmt.Errorf("upload to s3: %w", err)
	}
	return nil
}
