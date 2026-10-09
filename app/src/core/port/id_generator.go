package port

type IDGenerator interface {
	NewExpenseID() string
}