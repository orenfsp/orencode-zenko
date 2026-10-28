package registry

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/orenfsp/orencode-zenko.git/internal/file"
	"github.com/orenfsp/orencode-zenko.git/internal/rbac"
)

type ReportHandler struct {
	registry *Service
	files    *file.Service
}

func NewReportHandler(registry *Service, files *file.Service) *ReportHandler {
	return &ReportHandler{registry: registry, files: files}
}

func (h *ReportHandler) UsersXLSX(w http.ResponseWriter, r *http.Request) {
	identity, ok := rbac.IdentityFromContext(r.Context())
	if !ok || identity.UserID <= 0 {
		writeError(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}
	users, err := h.allUsers(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "не удалось сформировать отчёт")
		return
	}
	content, err := buildUsersXLSX(users)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "не удалось сформировать отчёт")
		return
	}
	record, err := h.files.Upload(r.Context(), file.UploadInput{
		Purpose: file.PurposeFederationReport, Filename: "federation-users.xlsx",
		Size: int64(len(content)), Content: bytes.NewReader(content),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "не удалось сохранить отчёт")
		return
	}
	download, err := h.files.DownloadStored(r.Context(), record.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "не удалось открыть отчёт")
		return
	}
	defer download.Content.Close()
	w.Header().Set("Content-Type", record.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(record.SizeBytes, 10))
	w.Header().Set("Content-Disposition", `attachment; filename="federation-users.xlsx"`)
	w.Header().Set("X-File-ID", strconv.FormatInt(record.ID, 10))
	_, _ = io.Copy(w, download.Content)
}

func (h *ReportHandler) allUsers(ctx context.Context) ([]User, error) {
	users := make([]User, 0)
	for offset := int32(0); ; offset += MaxLimit {
		page, err := h.registry.ListUsers(ctx, UserFilter{Limit: MaxLimit, Offset: offset})
		if err != nil {
			return nil, err
		}
		users = append(users, page.Items...)
		if len(page.Items) < int(MaxLimit) {
			return users, nil
		}
	}
}

func buildUsersXLSX(users []User) ([]byte, error) {
	var buffer bytes.Buffer
	archive := zip.NewWriter(&buffer)
	files := map[string]string{
		"[Content_Types].xml":        `<?xml version="1.0" encoding="UTF-8"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/><Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/></Types>`,
		"_rels/.rels":                `<?xml version="1.0" encoding="UTF-8"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/></Relationships>`,
		"xl/workbook.xml":            `<?xml version="1.0" encoding="UTF-8"?><workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets><sheet name="Пользователи" sheetId="1" r:id="rId1"/></sheets></workbook>`,
		"xl/_rels/workbook.xml.rels": `<?xml version="1.0" encoding="UTF-8"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/></Relationships>`,
	}
	for name, content := range files {
		if err := writeZipFile(archive, name, []byte(content)); err != nil {
			return nil, err
		}
	}
	var sheet strings.Builder
	sheet.WriteString(`<?xml version="1.0" encoding="UTF-8"?><worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>`)
	headers := []string{"ID", "Фамилия", "Имя", "Отчество", "Дата рождения", "Организация", "Муниципалитет", "Роли", "Дисциплины"}
	writeRow(&sheet, 1, headers)
	for i, user := range users {
		roles := make([]string, 0, len(user.Roles))
		for _, role := range user.Roles {
			roles = append(roles, role.Name)
		}
		disciplines := make([]string, 0, len(user.Disciplines))
		for _, discipline := range user.Disciplines {
			disciplines = append(disciplines, discipline.Name)
		}
		middle := ""
		if user.MiddleName != nil {
			middle = *user.MiddleName
		}
		organization, municipality := "", ""
		if user.Organization != nil {
			organization = user.Organization.Name
		}
		if user.Municipality != nil {
			municipality = user.Municipality.Name
		}
		writeRow(&sheet, i+2, []string{strconv.FormatInt(user.ID, 10), user.LastName, user.FirstName, middle, user.Birth.Format("2006-01-02"), organization, municipality, strings.Join(roles, ", "), strings.Join(disciplines, ", ")})
	}
	sheet.WriteString(`</sheetData></worksheet>`)
	if err := writeZipFile(archive, "xl/worksheets/sheet1.xml", []byte(sheet.String())); err != nil {
		return nil, err
	}
	if err := archive.Close(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func writeRow(sheet *strings.Builder, row int, values []string) {
	fmt.Fprintf(sheet, `<row r="%d">`, row)
	for column, value := range values {
		cell := string(rune('A' + column))
		fmt.Fprintf(sheet, `<c r="%s%d" t="inlineStr"><is><t xml:space="preserve">%s</t></is></c>`, cell, row, xmlEscape(value))
	}
	sheet.WriteString(`</row>`)
}

func writeZipFile(archive *zip.Writer, name string, content []byte) error {
	writer, err := archive.Create(name)
	if err != nil {
		return err
	}
	_, err = writer.Write(content)
	return err
}

func xmlEscape(value string) string {
	var buffer bytes.Buffer
	_ = xml.EscapeText(&buffer, []byte(value))
	return buffer.String()
}
