package result

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
)

const protocolTasksPerSection = 12

type ProtocolData struct {
	Table Table
	Date  time.Time
}

type PDFProtocolGenerator struct{}

func NewPDFProtocolGenerator() *PDFProtocolGenerator {
	return &PDFProtocolGenerator{}
}

func (g *PDFProtocolGenerator) Generate(data ProtocolData) ([]byte, error) {
	pdf := fpdf.New("L", "mm", "A4", "")
	pdf.SetMargins(10, 9, 10)
	pdf.SetAutoPageBreak(false, 10)
	pdf.AddUTF8FontFromBytes("Go", "", goregular.TTF)
	pdf.AddUTF8FontFromBytes("Go", "B", gobold.TTF)
	pdf.SetTitle("Протокол соревнования", true)

	entries := append([]Entry(nil), data.Table.Entries...)
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].Place == nil {
			return false
		}
		if entries[j].Place == nil {
			return true
		}
		if *entries[i].Place != *entries[j].Place {
			return *entries[i].Place < *entries[j].Place
		}
		return entries[i].ID < entries[j].ID
	})

	for firstTask := 1; firstTask <= int(data.Table.Sheet.TaskCount); firstTask += protocolTasksPerSection {
		lastTask := firstTask + protocolTasksPerSection - 1
		if lastTask > int(data.Table.Sheet.TaskCount) {
			lastTask = int(data.Table.Sheet.TaskCount)
		}
		g.addPageHeader(pdf, data, firstTask, lastTask)
		g.addTableHeader(pdf, firstTask, lastTask)
		for index, entry := range entries {
			pdf.SetFont("Go", "", 6.5)
			lines := pdf.SplitText(entryLabel(entry), 72)
			if len(lines) == 0 {
				lines = []string{""}
			}
			for len(lines) > 0 {
				count := len(lines)
				if count > 30 {
					count = 30
				}
				height := float64(count)*3.5 + 2
				if height < 8 {
					height = 8
				}
				if pdf.GetY()+height > 190 {
					g.addPageHeader(pdf, data, firstTask, lastTask)
					g.addTableHeader(pdf, firstTask, lastTask)
				}
				g.addEntryRow(pdf, index+1, entry, firstTask, lastTask, strings.Join(lines[:count], "\n"), height)
				lines = lines[count:]
			}
		}
		if lastTask == int(data.Table.Sheet.TaskCount) {
			g.addSummaryAndSignatures(pdf, data, entries)
		}
	}

	var output bytes.Buffer
	if err := pdf.Output(&output); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func (g *PDFProtocolGenerator) addPageHeader(
	pdf *fpdf.Fpdf, data ProtocolData, firstTask, lastTask int,
) {
	pdf.AddPage()
	pdf.SetFont("Go", "B", 13)
	pdf.CellFormat(0, 7, "ПРОТОКОЛ РЕЗУЛЬТАТОВ", "", 1, "C", false, 0, "")
	pdf.SetFont("Go", "", 9)
	pdf.MultiCell(0, 5, data.Table.Context.CompetitionName, "", "C", false)
	eventName := fmt.Sprintf("Мероприятие №%d", data.Table.Context.EventID)
	if data.Table.Context.AgeGroupName != nil {
		eventName += ", " + *data.Table.Context.AgeGroupName
	}
	pdf.MultiCell(0, 5, eventName+"; "+data.Table.Context.DisciplineName, "", "C", false)
	if data.Table.Sheet.TaskCount > protocolTasksPerSection {
		pdf.SetFont("Go", "", 7)
		pdf.CellFormat(0, 4, fmt.Sprintf("Задачи %d-%d из %d", firstTask, lastTask, data.Table.Sheet.TaskCount), "", 1, "C", false, 0, "")
	}
	pdf.Ln(2)
}

func (g *PDFProtocolGenerator) addTableHeader(pdf *fpdf.Fpdf, firstTask, lastTask int) {
	n := lastTask - firstTask + 1
	const (
		rowNumberWidth = 8.0
		nameWidth      = 74.0
		solvedWidth    = 18.0
		timeWidth      = 29.0
		placeWidth     = 14.0
	)
	taskWidth := (277 - rowNumberWidth - nameWidth - solvedWidth - timeWidth - placeWidth) / float64(n)
	pdf.SetFont("Go", "B", 6.5)
	pdf.CellFormat(rowNumberWidth, 12, "№", "1", 0, "C", false, 0, "")
	pdf.CellFormat(nameWidth, 12, "Команда (состав) или участник", "1", 0, "C", false, 0, "")
	for task := firstTask; task <= lastTask; task++ {
		pdf.CellFormat(taskWidth, 12, fmt.Sprintf("%d", task), "1", 0, "C", false, 0, "")
	}
	pdf.CellFormat(solvedWidth, 12, "Решено", "1", 0, "C", false, 0, "")
	pdf.CellFormat(timeWidth, 12, "Время", "1", 0, "C", false, 0, "")
	pdf.CellFormat(placeWidth, 12, "Место", "1", 1, "C", false, 0, "")
}

func (g *PDFProtocolGenerator) addEntryRow(
	pdf *fpdf.Fpdf, number int, entry Entry, firstTask, lastTask int, label string, rowHeight float64,
) {
	n := lastTask - firstTask + 1
	const (
		rowNumberWidth = 8.0
		nameWidth      = 74.0
		solvedWidth    = 18.0
		timeWidth      = 29.0
		placeWidth     = 14.0
	)
	taskWidth := (277 - rowNumberWidth - nameWidth - solvedWidth - timeWidth - placeWidth) / float64(n)
	pdf.SetFont("Go", "", 6.5)
	pdf.CellFormat(rowNumberWidth, rowHeight, fmt.Sprintf("%d", number), "1", 0, "C", false, 0, "")
	x, y := pdf.GetXY()
	pdf.Rect(x, y, nameWidth, rowHeight, "D")
	pdf.MultiCell(nameWidth, 3.5, label, "", "L", false)
	pdf.SetXY(x+nameWidth, y)
	tasks := make(map[int16]TaskResult, len(entry.Tasks))
	for _, task := range entry.Tasks {
		tasks[task.TaskNumber] = task
	}
	for taskNumber := firstTask; taskNumber <= lastTask; taskNumber++ {
		pdf.CellFormat(taskWidth, rowHeight, protocolTaskValue(tasks[int16(taskNumber)]), "1", 0, "C", false, 0, "")
	}
	pdf.CellFormat(solvedWidth, rowHeight, fmt.Sprintf("%d", entry.SolvedCount), "1", 0, "C", false, 0, "")
	pdf.CellFormat(timeWidth, rowHeight, formatDeciseconds(entry.TotalTimeDeciseconds), "1", 0, "C", false, 0, "")
	place := "-"
	if entry.Place != nil {
		place = fmt.Sprintf("%d", *entry.Place)
	}
	pdf.CellFormat(placeWidth, rowHeight, place, "1", 1, "C", false, 0, "")
}

func (g *PDFProtocolGenerator) addSummaryAndSignatures(
	pdf *fpdf.Fpdf, data ProtocolData, entries []Entry,
) {
	if pdf.GetY()+43 > 200 {
		pdf.AddPage()
	}
	pdf.Ln(4)
	allSolved, noneSolved := 0, 0
	for _, entry := range entries {
		if entry.ResultStatus != ResultFinished {
			continue
		}
		if entry.SolvedCount == int32(data.Table.Sheet.TaskCount) {
			allSolved++
		}
		if entry.SolvedCount == 0 {
			noneSolved++
		}
	}
	pdf.SetFont("Go", "", 8)
	stats := fmt.Sprintf(
		"К решению были предложены %d задач; участников/команд, решивших все задачи: %d; не решивших ни одной задачи: %d.",
		data.Table.Sheet.TaskCount, allSolved, noneSolved,
	)
	pdf.MultiCell(0, 5, stats, "", "L", false)
	pdf.Ln(3)
	pdf.CellFormat(65, 6, "Главный судья соревнований", "", 0, "L", false, 0, "")
	pdf.CellFormat(55, 6, "____________________", "", 0, "C", false, 0, "")
	pdf.CellFormat(0, 6, judgeName(data.Table.Judges, "главный судья"), "", 1, "L", false, 0, "")
	pdf.CellFormat(65, 6, "Главный секретарь соревнований", "", 0, "L", false, 0, "")
	pdf.CellFormat(55, 6, "____________________", "", 0, "C", false, 0, "")
	pdf.CellFormat(0, 6, judgeName(data.Table.Judges, "главный секретарь"), "", 1, "L", false, 0, "")
	pdf.SetFont("Go", "", 7)
	pdf.CellFormat(65, 4, "", "", 0, "", false, 0, "")
	pdf.CellFormat(55, 4, "(подпись)", "", 0, "C", false, 0, "")
	pdf.CellFormat(0, 4, "(инициалы, фамилия)", "", 1, "L", false, 0, "")
	pdf.Ln(2)
	pdf.SetFont("Go", "", 8)
	pdf.CellFormat(0, 6, data.Date.Format("02.01.2006")+" г.                                      М.П.", "", 1, "L", false, 0, "")
}

func entryLabel(entry Entry) string {
	parts := make([]string, 0, len(entry.Applications))
	for _, application := range entry.Applications {
		name := strings.TrimSpace(application.LastName + " " + application.FirstName)
		if application.MiddleName != nil {
			name += " " + strings.TrimSpace(*application.MiddleName)
		}
		parts = append(parts, strings.TrimSpace(name))
	}
	label := ""
	if entry.TeamName != nil {
		if len(parts) == 0 {
			label = *entry.TeamName
		} else {
			label = *entry.TeamName + ": " + strings.Join(parts, ", ")
		}
	} else {
		label = strings.Join(parts, ", ")
	}
	switch entry.ResultStatus {
	case ResultDidNotStart:
		label += " [не стартовал]"
	case ResultDisqualified:
		label += " [дисквалифицирован]"
	}
	return label
}

func protocolTaskValue(task TaskResult) string {
	switch task.Status {
	case TaskSolved:
		if task.SolveTimeDeciseconds == nil {
			return "+"
		}
		if task.NegativeAttempts > 0 {
			return fmt.Sprintf("+%d", task.NegativeAttempts)
		}
		return "+"
	case TaskAttempted:
		if task.NegativeAttempts > 0 {
			return fmt.Sprintf("-%d", task.NegativeAttempts)
		}
		return "-"
	default:
		return "."
	}
}

func formatDeciseconds(value int64) string {
	return fmt.Sprintf("%d.%d с", value/10, value%10)
}

func judgeName(judges []Judge, position string) string {
	for _, judge := range judges {
		if !strings.EqualFold(strings.TrimSpace(judge.PositionName), position) {
			continue
		}
		initials := ""
		if judge.FirstName != "" {
			initials += string([]rune(judge.FirstName)[0]) + "."
		}
		if judge.MiddleName != nil && *judge.MiddleName != "" {
			initials += string([]rune(*judge.MiddleName)[0]) + "."
		}
		return strings.TrimSpace(initials + " " + judge.LastName)
	}
	return "____________________________"
}
