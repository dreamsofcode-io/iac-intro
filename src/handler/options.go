package handler

// Option is a function that can be passed to NewHandler to configure it.
type Option func(*Handler)

// WithS3Client sets the S3 client on the handler.
func WithS3Client(client S3Client) Option {
	return func(h *Handler) {
		h.client = client
	}
}

func WithBucketName(bucket string) Option {
	return func(h *Handler) {
		h.bucketName = bucket
	}
}
