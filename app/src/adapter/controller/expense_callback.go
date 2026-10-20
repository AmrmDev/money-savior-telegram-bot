package controller

import (
	"errors"
	"strconv"
	"strings"
)

const (
	maxCallbackDataBytes = 64

	callbackSeparator  = "|"
	confirmPrefix      = "g"
	cancelCallbackData = "x"
)

var (
	errCallbackTooLong = errors.New("callback data too long")
	errCallbackInvalid = errors.New("invalid callback data")
)

type callbackAction int

const (
	callbackConfirm callbackAction = iota + 1
	callbackCancel
)

type expenseData struct {
	AmountCents int64
	Category    string
	Method      string
}

func encodeConfirmCallback(amountCents int64, category, method string) (string, error) {
	if amountCents <= 0 || category == "" ||
		strings.Contains(category, callbackSeparator) || strings.Contains(method, callbackSeparator) {
		return "", errCallbackInvalid
	}

	data := strings.Join(
		[]string{confirmPrefix, strconv.FormatInt(amountCents, 10), category, method},
		callbackSeparator,
	)
	if len(data) > maxCallbackDataBytes {
		return "", errCallbackTooLong
	}
	return data, nil
}

func decodeCallback(data string) (callbackAction, expenseData, error) {
	if data == cancelCallbackData {
		return callbackCancel, expenseData{}, nil
	}

	parts := strings.Split(data, callbackSeparator)
	if len(parts) != 4 || parts[0] != confirmPrefix || !isDigits(parts[1]) || parts[2] == "" {
		return 0, expenseData{}, errCallbackInvalid
	}

	cents, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || cents <= 0 {
		return 0, expenseData{}, errCallbackInvalid
	}

	return callbackConfirm, expenseData{AmountCents: cents, Category: parts[2], Method: parts[3]}, nil
}
