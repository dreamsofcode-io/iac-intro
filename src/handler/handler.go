package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
)

// ErrNoBucketName is returned when the BUCKET_NAME environment variable is not set.
var ErrNoBucketName = errors.New("no bucket name")

// Handler is the main entrypoint for the lambda function.
type Handler struct {
	logger     *zap.Logger
	client     S3Client
	bucketName string
}

// New creates a new handler with the given context and logger.
// It also creates a new S3 client and sets the bucket name from the BUCKET_NAME environment variable.
// If the client or bucket name are already set by the options, they will not be overwritten.
func New(ctx context.Context, logger *zap.Logger, opts ...Option) (*Handler, error) {
	h := &Handler{
		logger: logger,
	}

	for _, opt := range opts {
		opt(h)
	}

	if h.client == nil {
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return nil, fmt.Errorf("load config: %w", err)
		}

		h.client = s3.NewFromConfig(cfg)
	}

	if h.bucketName == "" {
		bucketName, exists := os.LookupEnv("BUCKET_NAME")
		if !exists {
			return nil, ErrNoBucketName
		}
		h.bucketName = bucketName
	}

	return h, nil
}

// Handle handles the given SQS event, writing the body of each record to S3.
func (h *Handler) Handle(ctx context.Context, msg *events.SQSEvent) error {
	for _, record := range msg.Records {
		data := strings.NewReader(record.Body)
		if err := h.writeToS3(ctx, record.MessageId, data); err != nil {
			return err
		}
	}

	return nil
}

// writeToS3 writes the given data to S3 with the given name.
func (h *Handler) writeToS3(ctx context.Context, name string, data io.Reader) error {
	_, err := h.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &h.bucketName,
		Key:    &name,
		Body:   data,
	})
	if err != nil {
		return err
	}

	return nil
}
