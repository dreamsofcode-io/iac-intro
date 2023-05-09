package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"go.uber.org/zap"

	"github.com/dreamsofcode-io/iac-intro/handler"
)

func main() {
	lambda.Start(func(ctx context.Context, msg *events.SQSEvent) error {
		logger, err := zap.NewProduction()
		defer logger.Sync()

		h, err := handler.New(ctx, logger)
		if err != nil {
			logger.Error("new handler", zap.Error(err))
			return fmt.Errorf("new handler: %w", err)
		}

		return h.Handle(ctx, msg)
	})
}
