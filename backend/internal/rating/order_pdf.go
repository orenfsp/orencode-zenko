package rating

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"
	"github.com/orenfsp/orencode-zenko.git/internal/auth"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
)

type PDFOrderGenerator struct{}

func NewPDFOrderGenerator() *PDFOrderGenerator { return &PDFOrderGenerator{} }

func (g *PDFOrderGenerator) Generate(selection Selection, coach *auth.Profile, createdAt time.Time) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(20, 15, 20)
	pdf.SetAutoPageBreak(true, 18)
	pdf.AddUTF8FontFromBytes("Go", "", goregular.TTF)
	pdf.AddUTF8FontFromBytes("Go", "B", gobold.TTF)
	pdf.SetTitle("Проект приказа о составе сборной", true)
	pdf.AddPage()
	pdf.SetFont("Go", "B", 13)
	pdf.CellFormat(0, 8, "ПРОЕКТ ПРИКАЗА", "", 1, "C", false, 0, "")
	pdf.SetFont("Go", "", 10)
	pdf.MultiCell(0, 6, fmt.Sprintf("О формировании сборной команды по дисциплине «%s» на сезон %d года", selection.Discipline.Name, selection.SeasonYear), "", "C", false)
	pdf.Ln(5)
	pdf.MultiCell(0, 6, "На основании утверждённого рейтинга спортсменов и решения Федерации спортивного программирования:", "", "L", false)
	pdf.Ln(2)
	pdf.SetFont("Go", "B", 10)
	pdf.CellFormat(0, 6, "ПРИКАЗЫВАЮ:", "", 1, "L", false, 0, "")
	pdf.SetFont("Go", "", 9)
	pdf.MultiCell(0, 5.5, fmt.Sprintf("1. Утвердить состав сборной команды в возрастной группе «%s»:", selection.AgeGroup.Name), "", "L", false)
	pdf.Ln(2)
	for _, member := range selection.Members {
		name := strings.TrimSpace(member.LastName + " " + member.FirstName)
		if member.MiddleName != nil {
			name += " " + strings.TrimSpace(*member.MiddleName)
		}
		mark := ""
		if selection.CaptainUserID != nil && *selection.CaptainUserID == member.UserID {
			mark = " — капитан команды"
		}
		pdf.CellFormat(10, 6, fmt.Sprintf("%d.", member.SelectionPosition), "", 0, "R", false, 0, "")
		pdf.MultiCell(0, 6, name+mark, "", "L", false)
	}
	if coach != nil {
		coachName := strings.TrimSpace(coach.LastName + " " + coach.FirstName)
		if coach.MiddleName != nil {
			coachName += " " + strings.TrimSpace(*coach.MiddleName)
		}
		pdf.Ln(2)
		pdf.MultiCell(0, 5.5, "2. Назначить тренером сборной команды: "+coachName+".", "", "L", false)
	}
	pdf.MultiCell(0, 5.5, "3. Контроль за исполнением настоящего приказа оставить за Федерацией спортивного программирования.", "", "L", false)
	pdf.Ln(12)
	pdf.CellFormat(65, 6, "Руководитель федерации", "", 0, "L", false, 0, "")
	pdf.CellFormat(45, 6, "________________", "", 0, "C", false, 0, "")
	pdf.CellFormat(0, 6, "________________________", "", 1, "L", false, 0, "")
	pdf.Ln(5)
	pdf.SetFont("Go", "", 8)
	pdf.CellFormat(0, 5, "Проект сформирован "+createdAt.Format("02.01.2006 15:04"), "", 1, "L", false, 0, "")

	var output bytes.Buffer
	if err := pdf.Output(&output); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}
