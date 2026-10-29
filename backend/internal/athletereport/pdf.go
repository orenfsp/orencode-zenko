package athletereport

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
)

type PDFGenerator struct{}

func NewPDFGenerator() *PDFGenerator { return &PDFGenerator{} }

func (g *PDFGenerator) Generate(data Data) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 13, 15)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddUTF8FontFromBytes("Go", "", goregular.TTF)
	pdf.AddUTF8FontFromBytes("Go", "B", gobold.TTF)
	pdf.SetTitle("Отчёт по спортсмену", true)
	pdf.AddPage()

	pdf.SetFont("Go", "B", 14)
	pdf.CellFormat(0, 8, "ФЕДЕРАЦИЯ СПОРТИВНОГО ПРОГРАММИРОВАНИЯ", "", 1, "C", false, 0, "")
	pdf.CellFormat(0, 8, "ОТЧЁТ ПО СПОРТСМЕНУ", "", 1, "C", false, 0, "")
	pdf.Ln(3)

	fullName := strings.TrimSpace(data.Profile.LastName + " " + data.Profile.FirstName)
	if data.Profile.MiddleName != nil {
		fullName += " " + strings.TrimSpace(*data.Profile.MiddleName)
	}
	pdf.SetFont("Go", "B", 11)
	pdf.CellFormat(0, 6, fullName, "", 1, "L", false, 0, "")
	pdf.SetFont("Go", "", 9)
	pdf.CellFormat(0, 5, "Дата рождения: "+data.Profile.BirthDate.Format("02.01.2006"), "", 1, "L", false, 0, "")
	if data.Profile.Organization != nil {
		pdf.CellFormat(0, 5, "Организация: "+data.Profile.Organization.Name, "", 1, "L", false, 0, "")
	}
	if data.Profile.Municipality != nil {
		pdf.CellFormat(0, 5, "Муниципалитет: "+data.Profile.Municipality.Name, "", 1, "L", false, 0, "")
	}

	pdf.Ln(5)
	pdf.SetFont("Go", "B", 11)
	pdf.CellFormat(0, 7, fmt.Sprintf("Результаты соревнований (%d)", len(data.Results)), "", 1, "L", false, 0, "")
	if len(data.Results) == 0 {
		pdf.SetFont("Go", "", 9)
		pdf.CellFormat(0, 6, "Утверждённых результатов пока нет.", "", 1, "L", false, 0, "")
	} else {
		for _, item := range data.Results {
			pdf.SetFont("Go", "B", 8.5)
			pdf.MultiCell(0, 5, item.CompetitionName+" · "+item.DisciplineName, "", "L", false)
			pdf.SetFont("Go", "", 8)
			details := item.StartAt.Format("02.01.2006")
			if item.Place != nil {
				details += fmt.Sprintf(" · место: %d", *item.Place)
			}
			if item.Score != nil {
				details += fmt.Sprintf(" · баллы: %g", *item.Score)
			}
			pdf.CellFormat(0, 4.5, details, "", 1, "L", false, 0, "")
			pdf.Ln(1.5)
		}
	}

	pdf.Ln(3)
	pdf.SetFont("Go", "B", 11)
	pdf.CellFormat(0, 7, fmt.Sprintf("Спортивные разряды (%d)", len(data.Ranks)), "", 1, "L", false, 0, "")
	if len(data.Ranks) == 0 {
		pdf.SetFont("Go", "", 9)
		pdf.CellFormat(0, 6, "Присвоенных разрядов пока нет.", "", 1, "L", false, 0, "")
	} else {
		for _, rank := range data.Ranks {
			pdf.SetFont("Go", "B", 8.5)
			pdf.CellFormat(0, 5, rank.RankName+" · "+rank.DisciplineName, "", 1, "L", false, 0, "")
			pdf.SetFont("Go", "", 8)
			details := ""
			if rank.AwardedAt != nil {
				details = "присвоен: " + rank.AwardedAt.Format("02.01.2006")
			}
			if rank.ExpiresAt != nil {
				details += " · действует до: " + rank.ExpiresAt.Format("02.01.2006")
			}
			if rank.OrderNumber != nil {
				details += " · приказ № " + *rank.OrderNumber
			}
			pdf.CellFormat(0, 4.5, strings.TrimPrefix(details, " · "), "", 1, "L", false, 0, "")
			pdf.Ln(1.5)
		}
	}

	pdf.Ln(5)
	pdf.SetFont("Go", "", 8)
	pdf.CellFormat(0, 5, "Сформировано платформой: "+data.CreatedAt.In(time.Local).Format("02.01.2006 15:04"), "", 1, "L", false, 0, "")

	var output bytes.Buffer
	if err := pdf.Output(&output); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}
