package entities

type Book struct {
	ID     int64
	Title  string
	ISBN   string
	Author *Author
}