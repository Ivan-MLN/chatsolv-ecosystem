package knowledge

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/ledongthuc/pdf"
	"github.com/nguyenthenguyen/docx"
)

var ErrUnsupportedType = errors.New("unsupported document type")

type Source struct{ ID, WorkspaceID, Type, Title string }
type Note struct{ Path, Title, Category, Content string }

var slugCleaner = regexp.MustCompile(`[^a-z0-9]+`)
var xmlTagCleaner = regexp.MustCompile(`<[^>]+>`)

func Convert(source Source, reader io.Reader, mime string) ([]Note, error) {
	switch mime {
	case "text/plain", "text/markdown", "application/json":
		return convertStructuredText(source, reader, "knowledge")
	case "text/csv", "application/csv":
		return convertCSV(source, reader)
	case "application/pdf":
		data, err := io.ReadAll(io.LimitReader(reader, 20<<20))
		if err != nil {
			return nil, err
		}
		parsed, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
		if err != nil {
			return nil, err
		}
		plain, err := parsed.GetPlainText()
		if err != nil {
			return nil, err
		}
		return convertStructuredText(source, plain, "knowledge")
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		data, err := io.ReadAll(io.LimitReader(reader, 20<<20))
		if err != nil {
			return nil, err
		}
		document, err := docx.ReadDocxFromMemory(bytes.NewReader(data), int64(len(data)))
		if err != nil {
			return nil, err
		}
		defer document.Close()
		return convertStructuredText(source, strings.NewReader(stripXML(document.Editable().GetContent())), "knowledge")
	default:
		return nil, ErrUnsupportedType
	}
}

func convertStructuredText(source Source, reader io.Reader, category string) ([]Note, error) {
	data, err := io.ReadAll(io.LimitReader(reader, 20<<20))
	if err != nil {
		return nil, err
	}
	body := strings.TrimSpace(string(data))
	if body == "" {
		return nil, errors.New("document contains no extractable text")
	}
	slug := slugify(source.Title)
	if slug == "" {
		slug = "document"
	}
	return []Note{{Path: category + "/" + slug + ".md", Title: source.Title, Category: category, Content: frontmatter(source, category, body)}}, nil
}

func convertCSV(source Source, reader io.Reader) ([]Note, error) {
	records, err := csv.NewReader(io.LimitReader(reader, 20<<20)).ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) < 2 {
		return nil, errors.New("CSV has no product rows")
	}
	headers := records[0]
	var notes []Note
	for index, row := range records[1:] {
		fields := map[string]string{}
		for i, h := range headers {
			if i < len(row) {
				fields[strings.ToLower(strings.TrimSpace(h))] = strings.TrimSpace(row[i])
			}
		}
		identifier := fields["sku"]
		if identifier == "" {
			identifier = fields["name"]
		}
		if identifier == "" {
			identifier = fmt.Sprintf("product-%d", index+1)
		}
		title := fields["name"]
		if title == "" {
			title = identifier
		}
		keys := make([]string, 0, len(fields))
		for key := range fields {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		var body strings.Builder
		body.WriteString("# " + title + "\n\n## Informasi Produk\n\n")
		for _, key := range keys {
			if key != "name" && fields[key] != "" {
				body.WriteString("- " + key + ": " + fields[key] + "\n")
			}
		}
		now := time.Now().UTC().Format(time.RFC3339)
		content := fmt.Sprintf("---\nid: %s-%d\nworkspace_id: %s\nsource_type: document\nsource_id: %s\ncategory: product\nversion: 1\nstatus: active\nsku: %s\ncreated_at: %s\nupdated_at: %s\n---\n\n%s", source.ID, index+1, source.WorkspaceID, source.ID, fields["sku"], now, now, body.String())
		notes = append(notes, Note{Path: "products/" + slugify(identifier) + ".md", Title: title, Category: "product", Content: content})
	}
	return notes, nil
}

func frontmatter(source Source, category, body string) string {
	now := time.Now().UTC().Format(time.RFC3339)
	return fmt.Sprintf("---\nid: %s\nworkspace_id: %s\nsource_type: %s\nsource_id: %s\ncategory: %s\nversion: 1\nstatus: active\ncreated_at: %s\nupdated_at: %s\n---\n\n# %s\n\n%s\n", source.ID, source.WorkspaceID, source.Type, source.ID, category, now, now, source.Title, strings.TrimSpace(body))
}
func slugify(value string) string {
	return strings.Trim(slugCleaner.ReplaceAllString(strings.ToLower(strings.TrimSpace(value)), "-"), "-")
}
func stripXML(value string) string {
	value = strings.ReplaceAll(value, "</w:p>", "\n\n")
	value = strings.ReplaceAll(value, "</w:tr>", "\n")
	return strings.TrimSpace(xmlTagCleaner.ReplaceAllString(value, ""))
}
