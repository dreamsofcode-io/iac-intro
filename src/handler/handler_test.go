package handler_test

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"

	"github.com/dreamsofcode-io/iac-intro/handler"
	"github.com/dreamsofcode-io/iac-intro/handler/mocks"
)

func TestHandler_Handle(t *testing.T) {
	tests := []struct {
		name    string
		expects func(*mocks.MockS3Client)
		bucket  string
		input   *events.SQSEvent
		wants   error
	}{
		{
			name:   "happy path",
			bucket: "test-bucket",
			expects: func(s3Client *mocks.MockS3Client) {
				s3Client.EXPECT().PutObject(gomock.Any(), &s3.PutObjectInput{
					Bucket: aws.String("test-bucket"),
					Key:    aws.String("test-key"),
					Body:   strings.NewReader("test-body"),
				}).Return(nil, nil)
			},
			input: &events.SQSEvent{
				Records: []events.SQSMessage{
					{
						MessageId: "test-key",
						Body:      "test-body",
					},
				},
			},
		},
		{
			name:   "sad path",
			bucket: "other-bucket",
			expects: func(s3Client *mocks.MockS3Client) {
				s3Client.EXPECT().PutObject(gomock.Any(), &s3.PutObjectInput{
					Bucket: aws.String("other-bucket"),
					Key:    aws.String("other-key"),
					Body:   strings.NewReader("other-body"),
				}).Return(nil, io.ErrClosedPipe)
			},
			input: &events.SQSEvent{
				Records: []events.SQSMessage{
					{
						MessageId: "other-key",
						Body:      "other-body",
					},
				},
			},
			wants: io.ErrClosedPipe,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			s3Client := mocks.NewMockS3Client(ctrl)

			if tt.expects != nil {
				tt.expects(s3Client)
			}

			l := zaptest.NewLogger(t)

			h, err := handler.New(
				context.Background(),
				l,
				handler.WithS3Client(s3Client),
				handler.WithBucketName(tt.bucket),
			)
			assert.NoError(t, err)

			err = h.Handle(context.Background(), tt.input)
			assert.ErrorIs(t, err, tt.wants)
		})
	}
}
