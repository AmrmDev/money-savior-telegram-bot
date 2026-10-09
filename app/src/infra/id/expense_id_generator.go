package id

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

type ExpenseIDGenerator struct{}

func NewExpenseIDGenerator() *ExpenseIDGenerator {
	return &ExpenseIDGenerator{}
}

func (g *ExpenseIDGenerator) NewExpenseID() string {
	return fmt.Sprintf("%03dMSB%03dPA", randomThreeDigits(), randomThreeDigits())
}

func randomThreeDigits() int64 {
	n, err := rand.Int(rand.Reader, big.NewInt(1000))
	if err != nil {
		panic(fmt.Sprintf("failed to generate random number: %v", err))
	}
	return n.Int64()
}