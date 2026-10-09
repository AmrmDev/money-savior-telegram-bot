package input

type RegisterExpense struct {
	UserID   int64
	AmountCents int64
	Category string
	Method   string
}