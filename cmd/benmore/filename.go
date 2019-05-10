package main

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
)

// Filename builds a standard hourly file path name.
func (b Benmore) Filename(base, path string) (string, error) {
	tmpl, err := template.New("txt").Funcs(
		template.FuncMap{
			"stamp": func(s string) string {
				return b.Timestamp.Format(s)
			},
			"year": func() string {
				return fmt.Sprintf("%04d", b.Timestamp.Year())
			},
			"yearday": func() string {
				return fmt.Sprintf("%03d", b.Timestamp.YearDay())
			},
			"hour": func() string {
				return fmt.Sprintf("%02d", b.Timestamp.Hour())
			},
			"minute": func() string {
				return fmt.Sprintf("%02d", b.Timestamp.Minute())
			},
			"second": func() string {
				return fmt.Sprintf("%02d", b.Timestamp.Second())
			},
			"tolower": func(s string) string {
				return strings.ToLower(s)
			},
		}).Parse(path)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, b); err != nil {
		return "", err
	}

	filename := filepath.Join(base, buf.String())

	return filename, nil

}
