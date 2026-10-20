package controller

type Callback struct {
	ID        string
	ChatID    int64
	UserID    int64
	MessageID int
	Data      string
}
