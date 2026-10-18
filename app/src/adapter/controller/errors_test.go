package controller

import (
	"context"
	"errors"
	"testing"

	"money-savior-telegram-bot/app/src/adapter/presenter"
	"money-savior-telegram-bot/app/src/core/domain/exception"
)

type fakeSender struct {
	replies []presenter.Reply
	err     error
}

func (f *fakeSender) Send(_ context.Context, _ int64, reply presenter.Reply) error {
	f.replies = append(f.replies, reply)
	return f.err
}

func TestReject_AnswersAndMarksInvalidInput(t *testing.T) {
	sender := &fakeSender{}

	err := reject(context.Background(), sender, 1, presenter.Reply{Text: "oops"}, "missing arguments")

	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
	if len(sender.replies) != 1 || sender.replies[0].Text != "oops" {
		t.Errorf("unexpected replies: %+v", sender.replies)
	}
}

func TestReject_SendFailureIsNotInvalidInput(t *testing.T) {
	boom := errors.New("telegram down")
	sender := &fakeSender{err: boom}

	err := reject(context.Background(), sender, 1, presenter.Reply{Text: "oops"}, "missing arguments")

	if !errors.Is(err, boom) {
		t.Errorf("expected the send error, got %v", err)
	}
	if errors.Is(err, ErrInvalidInput) {
		t.Error("a failed reply must not be reported as invalid input")
	}
}

func TestRespondError_DomainErrorsAreInvalidInput(t *testing.T) {
	domainErrors := []error{
		exception.ErrInvalidAmount,
		exception.ErrInvalidCategory,
		exception.ErrInvalidPeriod,
	}

	for _, domainErr := range domainErrors {
		t.Run(domainErr.Error(), func(t *testing.T) {
			sender := &fakeSender{}

			err := respondError(context.Background(), sender, 1, presenter.Reply{Text: "x"}, domainErr)

			if !errors.Is(err, ErrInvalidInput) {
				t.Errorf("expected ErrInvalidInput, got %v", err)
			}
			if len(sender.replies) != 1 {
				t.Errorf("expected 1 reply, got %d", len(sender.replies))
			}
		})
	}
}

func TestRespondError_TechnicalErrorIsReturned(t *testing.T) {
	boom := errors.New("dynamo down")
	sender := &fakeSender{}

	err := respondError(context.Background(), sender, 1, presenter.Reply{Text: "x"}, boom)

	if !errors.Is(err, boom) {
		t.Errorf("expected the original error, got %v", err)
	}
	if errors.Is(err, ErrInvalidInput) {
		t.Error("a technical error must not be reported as invalid input")
	}
	if len(sender.replies) != 1 {
		t.Errorf("expected 1 reply, got %d", len(sender.replies))
	}
}

func TestRespondError_KeepsBothErrorsWhenReplyFails(t *testing.T) {
	boom := errors.New("dynamo down")
	sendErr := errors.New("telegram down")
	sender := &fakeSender{err: sendErr}

	err := respondError(context.Background(), sender, 1, presenter.Reply{Text: "x"}, boom)

	if !errors.Is(err, boom) || !errors.Is(err, sendErr) {
		t.Errorf("expected both errors, got %v", err)
	}
}
