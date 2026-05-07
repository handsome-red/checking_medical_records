package model

type Option struct {
	Text 	string
	Correct bool
}

type Question struct {
	ID		int
	Text 	string
	Options	[]Option
}