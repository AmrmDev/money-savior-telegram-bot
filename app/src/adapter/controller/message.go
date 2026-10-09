package controller

type Message struct {
	ChatID    int64
	UserID    int64
	Command   string
	Args      []string
}
