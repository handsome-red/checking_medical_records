package model

type ImportData struct {
	Books []ImportBook `json:"books"`
}

type ImportBook struct {
	ID        int              `json:"id"`
	Name      string           `json:"name"`
	BookPath  string           `json:"book_path"`
	Pages     []ImportPage     `json:"pages,omitempty"`
	Questions []ImportQuestion `json:"questions,omitempty"`
}

type ImportPage struct {
	// ID     int    `json:"id"`
	// BookID int    `json:"book_id"`
	Number int    `json:"number"`
	Path   string `json:"path"`
}

type ImportQuestion struct {
	ID      int            `json:"id"`
	BookID  int            `json:"book_id"`
	Text    string         `json:"text"`
	Options []ImportOption `json:"options"`
}

type ImportOption struct {
	ID      int    `json:"id"`
	Text    string `json:"text"`
	Correct bool   `json:"correct"`
}
