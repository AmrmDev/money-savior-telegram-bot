package input

type RegisterExpense struct {
	UserID   int64
	Amount   float64
	Category string
	Method   string
}