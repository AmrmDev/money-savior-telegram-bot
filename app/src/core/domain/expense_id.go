package domain

import "regexp"

var expenseIDPattern = regexp.MustCompile(`^\d{3}MSB\d{3}PA$`)

func IsValidExpenseID(id string) bool {
	return expenseIDPattern.MatchString(id)
}