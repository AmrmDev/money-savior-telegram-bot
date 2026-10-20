package controller

import (
	"errors"
	"strings"
	"testing"
)

func TestCallback_RoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		cents    int64
		category string
		method   string
	}{
		{"com método", 300, "Água", "Pix"},
		{"sem método", 4573, "Mercado", ""},
		{"categoria com espaço", 1200, "Casa De Praia", "Crédito"},
		{"valor grande", 123456789, "Carro", "Débito"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := encodeConfirmCallback(tt.cents, tt.category, tt.method)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(data) > maxCallbackDataBytes {
				t.Fatalf("data has %d bytes, the limit is %d", len(data), maxCallbackDataBytes)
			}

			action, got, err := decodeCallback(data)
			if err != nil || action != callbackConfirm {
				t.Fatalf("unexpected result: action=%v err=%v", action, err)
			}
			want := expenseData{AmountCents: tt.cents, Category: tt.category, Method: tt.method}
			if got != want {
				t.Errorf("got %+v, want %+v", got, want)
			}
		})
	}
}

func TestCallback_CancelRoundTrip(t *testing.T) {
	action, _, err := decodeCallback(cancelCallbackData)
	if err != nil || action != callbackCancel {
		t.Errorf("unexpected result: action=%v err=%v", action, err)
	}
}

func TestEncodeConfirmCallback_ByteLimit(t *testing.T) {
	// "g|300|" is 6 bytes and the empty method leaves one more "|": 57 left for the category.
	if _, err := encodeConfirmCallback(300, strings.Repeat("a", 57), ""); err != nil {
		t.Errorf("64 bytes must fit, got %v", err)
	}

	if _, err := encodeConfirmCallback(300, strings.Repeat("a", 58), ""); !errors.Is(err, errCallbackTooLong) {
		t.Errorf("65 bytes must not fit, got %v", err)
	}

	// 30 runes of 2 bytes each are 60 bytes: the limit counts bytes, not characters.
	if _, err := encodeConfirmCallback(300, strings.Repeat("á", 30), ""); !errors.Is(err, errCallbackTooLong) {
		t.Errorf("multibyte text must be counted in bytes, got %v", err)
	}
}

func TestEncodeConfirmCallback_RejectsUnsafeValues(t *testing.T) {
	tests := []struct {
		name     string
		cents    int64
		category string
		method   string
	}{
		{"categoria com separador", 300, "Ág|ua", "Pix"},
		{"método com separador", 300, "Água", "Pi|x"},
		{"categoria vazia", 300, "", "Pix"},
		{"valor zero", 0, "Água", "Pix"},
		{"valor negativo", -5, "Água", "Pix"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := encodeConfirmCallback(tt.cents, tt.category, tt.method); !errors.Is(err, errCallbackInvalid) {
				t.Errorf("expected errCallbackInvalid, got %v", err)
			}
		})
	}
}

func TestDecodeCallback_RejectsAnythingUnexpected(t *testing.T) {
	tests := []string{
		"",
		"g",
		"X",
		"x|1",
		"z|300|Água|Pix",
		"g|300",
		"g|300|Água",
		"g|abc|Água|Pix",
		"g|0|Água|Pix",
		"g|-5|Água|Pix",
		"g|+5|Água|Pix",
		"g|3.5|Água|Pix",
		"g|300||Pix",
		"g|300|Água|Pix|extra",
		"g|99999999999999999999|Água|Pix",
	}

	for _, data := range tests {
		t.Run(data, func(t *testing.T) {
			if _, _, err := decodeCallback(data); !errors.Is(err, errCallbackInvalid) {
				t.Errorf("expected errCallbackInvalid, got %v", err)
			}
		})
	}
}
