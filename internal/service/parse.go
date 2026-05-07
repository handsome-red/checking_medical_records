package service

import (
	"os"
	"fmt"
	"encoding/json"
)

type Test struct {
	Questions []Question `json:"questions"`
}

type Question struct {
	ID 		int			`json:"id"`
	Text 	string		`json:"text"`
	Options	[]Option	`json:"options"`
}

type Option struct {
	Text 	string	`json:"text"`
	Correct	bool	`json:"correct"`
}

func NewTest(path string) (*Test, error) {
	if path == "" {
		path = "pkg/questions/questions.json"
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("Cannot open file: %w", err)
	}
	defer file.Close()

	var test Test
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&test)
	if err != nil {
		return nil, fmt.Errorf("Cannot decode file: %w", err)
	}
	
	return &test, nil
}