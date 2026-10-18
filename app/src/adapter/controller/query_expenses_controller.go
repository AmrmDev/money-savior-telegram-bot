package controller

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"money-savior-telegram-bot/app/src/adapter/presenter"
	"money-savior-telegram-bot/app/src/core/domain"
	"money-savior-telegram-bot/app/src/core/usecase"
	"money-savior-telegram-bot/app/src/core/usecase/input"
)

const queryExpensesTimeout = 5 * time.Second

var errInvalidPeriodFormat = errors.New("invalid period format")

var periodKeywords = map[string]domain.PeriodKind{
	"hoje":   domain.PeriodToday,
	"ontem":  domain.PeriodYesterday,
	"semana": domain.PeriodLast7Days,
	"mes":    domain.PeriodThisMonth,
}

type QueryExpensesUseCase interface {
	Execute(ctx context.Context, in input.QueryExpenses) (usecase.ExpensesSummary, error)
}

type QueryExpensesController struct {
	queryExpenses QueryExpensesUseCase
	sender        Sender
}

func NewQueryExpensesController(queryExpenses QueryExpensesUseCase, sender Sender) *QueryExpensesController {
	return &QueryExpensesController{queryExpenses: queryExpenses, sender: sender}
}

func (c *QueryExpensesController) Handle(ctx context.Context, message Message) error {
	spec, err := parsePeriodSpec(message.Args)
	if err != nil {
		return c.sender.Send(ctx, message.ChatID, presenter.InvalidQueryFormat())
	}

	execCtx, cancel := context.WithTimeout(ctx, queryExpensesTimeout)
	defer cancel()

	summary, err := c.queryExpenses.Execute(execCtx, input.QueryExpenses{
		UserID: message.UserID,
		Period: spec,
	})
	if err != nil {
		if sendErr := c.sender.Send(ctx, message.ChatID, presenter.QueryExpensesError(err)); sendErr != nil {
			return sendErr
		}
		return err
	}

	return c.sender.Send(ctx, message.ChatID, presenter.QuerySummary(summary))
}

func parsePeriodSpec(args []string) (domain.PeriodSpec, error) {
	words := make([]string, len(args))
	for i, arg := range args {
		words[i] = normalizeWord(arg)
	}

	switch len(words) {
	case 0:
		return domain.PeriodSpec{Kind: domain.PeriodThisMonth}, nil
	case 1:
		if kind, ok := periodKeywords[words[0]]; ok {
			return domain.PeriodSpec{Kind: kind}, nil
		}
	case 2:
		if words[0] == "mes" && words[1] == "passado" {
			return domain.PeriodSpec{Kind: domain.PeriodLastMonth}, nil
		}
		from, errFrom := parseDayMonth(words[0])
		to, errTo := parseDayMonth(words[1])
		if errFrom == nil && errTo == nil {
			return domain.PeriodSpec{Kind: domain.PeriodCustom, From: from, To: to}, nil
		}
	}

	return domain.PeriodSpec{}, errInvalidPeriodFormat
}

func normalizeWord(text string) string {
	return strings.ReplaceAll(strings.ToLower(strings.TrimSpace(text)), "ê", "e")
}

func parseDayMonth(text string) (domain.Date, error) {
	parts := strings.Split(text, "/")
	if len(parts) != 2 || !isDigits(parts[0]) || !isDigits(parts[1]) {
		return domain.Date{}, errInvalidPeriodFormat
	}

	day, errDay := strconv.Atoi(parts[0])
	month, errMonth := strconv.Atoi(parts[1])
	if errDay != nil || errMonth != nil {
		return domain.Date{}, errInvalidPeriodFormat
	}

	return domain.Date{Month: time.Month(month), Day: day}, nil
}
