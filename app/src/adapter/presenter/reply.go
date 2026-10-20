package presenter

type Button struct {
	Label string
	Data  string
}

type Reply struct {
	Text    string
	Image   []byte
	Buttons []Button
}
