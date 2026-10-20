package port

import "context"

// Transcriber turns a voice message (OGG/Opus, as sent by Telegram) into text.
// An empty text with no error means no speech was recognized.
type Transcriber interface {
	Transcribe(ctx context.Context, audio []byte) (string, error)
}
