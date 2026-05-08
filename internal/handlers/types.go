// internal/handlers/types.go
package handlers

import "med_book/internal/service"

type PageData struct {
	CurrentQuestion service.Question
	SelectedAnswers map[int][]int
	Step            int
	Total           int
	HasPrev         bool
	HasNext         bool
	AllAnswers      map[int][]int
}
