package controller

import (
	"context"
	"errors"
	"fmt"

	"money-savior-telegram-bot/app/src/adapter/presenter"
	"money-savior-telegram-bot/app/src/core/usecase/input"
)

type ConfirmExpenseController struct {
	usecase   RegisterExpenseUseCase
	responder CallbackResponder
}

func NewConfirmExpenseController(usecase RegisterExpenseUseCase, responder CallbackResponder) *ConfirmExpenseController {
	return &ConfirmExpenseController{usecase: usecase, responder: responder}
}

func (c *ConfirmExpenseController) Handle(ctx context.Context, callback Callback) error {
	answerErr := c.responder.AnswerCallback(ctx, callback.ID)
	return errors.Join(c.handle(ctx, callback), answerErr)
}

func (c *ConfirmExpenseController) handle(ctx context.Context, callback Callback) error {
	if callback.MessageID == 0 {
		return fmt.Errorf("%w: callback without message", ErrInvalidInput)
	}

	action, data, err := decodeCallback(callback.Data)
	if err != nil {
		return c.finish(ctx, callback, presenter.CallbackInvalid(), fmt.Errorf("%w: invalid callback data", ErrInvalidInput))
	}

	if action == callbackCancel {
		return c.finish(ctx, callback, presenter.ExpenseCancelled(), nil)
	}
	return c.confirm(ctx, callback, data)
}

func (c *ConfirmExpenseController) confirm(ctx context.Context, callback Callback, data expenseData) error {
	execCtx, cancel := context.WithTimeout(ctx, registerExpenseTimeout)
	defer cancel()

	expense, err := c.usecase.Execute(execCtx, input.RegisterExpense{
		UserID:      callback.UserID,
		AmountCents: data.AmountCents,
		Category:    data.Category,
		Method:      data.Method,
	})
	if err != nil {
		if isInvalidInput(err) {
			return c.finish(ctx, callback, presenter.RegisterExpenseError(err), fmt.Errorf("%w: %s", ErrInvalidInput, err.Error()))
		}
		return c.finish(ctx, callback, presenter.RegisterExpenseError(err), err)
	}

	return c.finish(ctx, callback, presenter.ExpenseRegistered(expense), nil)
}

func (c *ConfirmExpenseController) finish(ctx context.Context, callback Callback, reply presenter.Reply, result error) error {
	if err := c.responder.Edit(ctx, callback.ChatID, callback.MessageID, reply); err != nil {
		return errors.Join(result, err)
	}
	return result
}
