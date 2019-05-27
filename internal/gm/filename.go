package gm

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
)

// Filename builds a standard hourly file path name.
func Filename(base, path string, format Formatter) (string, error) {
	tmpl, err := template.New("gm").Funcs(
		template.FuncMap{
			"tag": func() string {
				return format.Tag()
			},
			"at": func(s string) string {
				return format.At().Format(s)
			},
			"year": func() string {
				return fmt.Sprintf("%04d", format.At().Year())
			},
			"yearday": func() string {
				return fmt.Sprintf("%03d", format.At().YearDay())
			},
			"hour": func() string {
				return fmt.Sprintf("%02d", format.At().Hour())
			},
			"minute": func() string {
				return fmt.Sprintf("%02d", format.At().Minute())
			},
			"second": func() string {
				return fmt.Sprintf("%02d", format.At().Second())
			},
			"tolower": func(s string) string {
				return strings.ToLower(s)
			},
			"toupper": func(s string) string {
				return strings.ToUpper(s)
			},
		}).Parse(path)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, format); err != nil {
		return "", err
	}

	filename := filepath.Join(base, buf.String())

	return filename, nil
}
