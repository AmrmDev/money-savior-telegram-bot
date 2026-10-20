package controller

import "context"

type AudioDownloader interface {
	Download(ctx context.Context, fileID string) ([]byte, error)
}

type MessageHandler interface {
	Handle(ctx context.Context, message Message) error
}
