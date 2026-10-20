package port

import "context"

type IntentParser interface {
	Parse(ctx context.Context, text string) (Intent, error)
}
