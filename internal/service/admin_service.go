package service

import (
	"bytes"
	"fmt"
	"time"

	"med_book/internal/repository"

	"github.com/xuri/excelize/v2"
)

type AdminService struct {
	sessionService *SessionService
}

func NewAdminService(sessionService *SessionService) *AdminService {
	return &AdminService{sessionService: sessionService}
}

func (s *AdminService) ExportReportsExcel(reports []repository.AdminSessionReport) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheet := "Results"
	index, err := f.NewSheet(sheet)
	if err != nil {
		return nil, err
	}
	f.SetActiveSheet(index)
	_ = f.DeleteSheet("Sheet1")

	headers := []string{
		"ФИО",
		"Дата",
		"Начало",
		"Окончание",
		"Тест",
		"Результат",
	}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, header)
	}

	maxAnswers := 0
	for _, report := range reports {
		if len(report.Answers) > maxAnswers {
			maxAnswers = len(report.Answers)
		}
	}

	for i := 0; i < maxAnswers; i++ {
		cell, _ := excelize.CoordinatesToCellName(len(headers)+i+1, 1)
		_ = f.SetCellValue(sheet, cell, fmt.Sprintf("Ответ %d", i+1))
	}

	for rowIdx, report := range reports {
		row := rowIdx + 2
		fio := fmt.Sprintf("%s %s %s", report.LastName, report.FirstName, report.Patronymic)

		startDate := report.StartedAt.Format("02.01.2006")
		startTime := report.StartedAt.Format("15:04")
		endTime := ""
		if report.CompletedAt != nil {
			endTime = report.CompletedAt.Format("15:04")
		}

		result := fmt.Sprintf("%d/%d", report.Score, report.MaxScore)

		values := []any{fio, startDate, startTime, endTime, report.BookTitle, result}
		for colIdx, value := range values {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, row)
			_ = f.SetCellValue(sheet, cell, value)
		}

		for i, answer := range report.Answers {
			cell, _ := excelize.CoordinatesToCellName(len(headers)+i+1, row)
			_ = f.SetCellValue(sheet, cell, joinAnswerTexts(answer.SelectedTexts))
		}
	}

	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})
	lastCol, _ := excelize.CoordinatesToCellName(len(headers)+maxAnswers, 1)
	_ = f.SetCellStyle(sheet, "A1", lastCol, style)

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func joinAnswerTexts(texts []string) string {
	if len(texts) == 0 {
		return "—"
	}
	result := texts[0]
	for i := 1; i < len(texts); i++ {
		result += " | " + texts[i]
	}
	return result
}

func FormatReportTime(t time.Time) string {
	return t.Format("15:04")
}
