package registry

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"
	"time"
)

func TestBuildUsersXLSXContainsRegistryAndDynamicsSheets(t *testing.T) {
	content, err := buildUsersXLSX([]User{
		{ID: 1, LastName: "Иванов", FirstName: "Иван", Birth: time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC), CreatedAt: time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)},
		{ID: 2, LastName: "Петров", FirstName: "Пётр", Birth: time.Date(2001, 2, 3, 0, 0, 0, 0, time.UTC), CreatedAt: time.Date(2024, 2, 6, 0, 0, 0, 0, time.UTC)},
	})
	if err != nil {
		t.Fatal(err)
	}
	archive, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		t.Fatal(err)
	}
	files := make(map[string]string, len(archive.File))
	for _, entry := range archive.File {
		reader, openErr := entry.Open()
		if openErr != nil {
			t.Fatal(openErr)
		}
		raw, readErr := io.ReadAll(reader)
		_ = reader.Close()
		if readErr != nil {
			t.Fatal(readErr)
		}
		files[entry.Name] = string(raw)
	}
	if !strings.Contains(files["xl/workbook.xml"], `name="Динамика"`) {
		t.Fatal("dynamics sheet is missing")
	}
	dynamics := files["xl/worksheets/sheet2.xml"]
	for _, expected := range []string{"Новых пользователей", "Всего пользователей", "01.2024", "02.2024"} {
		if !strings.Contains(dynamics, expected) {
			t.Fatalf("dynamics sheet does not contain %q", expected)
		}
	}
}

func TestBuildDynamicsReturnsMonthlyAndCumulativeCounts(t *testing.T) {
	points := buildDynamics([]User{
		{CreatedAt: time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)},
		{CreatedAt: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
		{CreatedAt: time.Date(2024, 2, 6, 0, 0, 0, 0, time.UTC)},
	})
	if len(points) != 2 || points[0].Month != "01.2024" || points[0].NewUsers != 2 || points[0].TotalUsers != 2 || points[1].TotalUsers != 3 {
		t.Fatalf("unexpected dynamics: %#v", points)
	}
}
