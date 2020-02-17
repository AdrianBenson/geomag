package raw

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"
)

const rawFormat = "2006-01-02 15:04:05Z"

type Reading struct {
	Timestamp time.Time
	Label     string
	Field     float64
}

func NewReading(t time.Time, l string, v float64) Reading {
	return Reading{
		Timestamp: t,
		Label:     l,
		Field:     v,
	}
}

func (r Reading) At() time.Time {
	return r.Timestamp
}

func (r Reading) Tag() string {
	return r.Label
}

func (r Reading) Value() float64 {
	return r.Field
}

func (r Reading) Less(reading Reading) bool {
	return r.Timestamp.Before(reading.Timestamp)
}

type Raw struct {
	Label     string
	Precision int
	Timestamp time.Time

	Readings []Reading
}

func NewRaw(label string, precision int) *Raw {
	return &Raw{
		Label:     label,
		Precision: precision,
	}
}

func (r *Raw) Add(v Reading) {
	if t := v.Timestamp; r.Timestamp.IsZero() || r.Timestamp.After(t) {
		r.Timestamp = t
	}
	r.Readings = append(r.Readings, v)
}

func (r *Raw) At() time.Time {
	return r.Timestamp
}

func (r *Raw) Tag() string {
	return r.Label
}

func (r *Raw) Marshal() ([]byte, error) {

	var buf bytes.Buffer
	if err := r.Encode(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (r *Raw) Unmarshal(data []byte) error {

	if err := r.Decode(bytes.NewReader(data)); err != nil {
		return err
	}

	return nil
}

func (r *Raw) Merge(data []byte) error {

	var raw Raw
	if err := raw.Unmarshal(data); err != nil {
		return err
	}

	cache := make(map[time.Time]Reading)
	for _, r := range raw.Readings {
		cache[r.Timestamp] = r
	}
	for _, r := range r.Readings {
		cache[r.Timestamp] = r
	}

	var readings []Reading
	for _, v := range cache {
		readings = append(readings, v)
	}

	r.Readings = readings

	return nil
}

func (r *Raw) Decode(rd io.Reader) error {

	records, err := csv.NewReader(rd).ReadAll()
	if err != nil {
		return err
	}

	for _, l := range records {
		if len(l) < 3 {
			continue
		}

		t, err := time.Parse(rawFormat, strings.TrimSpace(l[0]))
		if err != nil {
			return err
		}

		v, err := strconv.ParseFloat(l[2], 64)
		if err != nil {
			return err
		}

		r.Readings = append(r.Readings, NewReading(t, l[1], v))
	}

	return nil
}

func (r *Raw) Encode(wr io.Writer) error {

	sort.Slice(r.Readings, func(i, j int) bool {
		return r.Readings[i].Less(r.Readings[j])
	})

	var lines [][]string
	for _, v := range r.Readings {
		lines = append(lines, []string{
			v.Timestamp.Format(rawFormat),
			v.Label,
			strconv.FormatFloat(v.Field, 'f', r.Precision, 64),
		})
	}

	w := csv.NewWriter(wr)

	w.WriteAll(lines)

	if err := w.Error(); err != nil {
		return err
	}

	return nil
}

// Filename builds a standard hourly file path name.
func (r *Raw) Filename(path string) ([]byte, error) {
	tmpl, err := template.New("raw").Funcs(
		template.FuncMap{
			"tag": func() string {
				return r.Tag()
			},
			"at": func(s string) string {
				return r.At().Format(s)
			},
			"year": func() string {
				return fmt.Sprintf("%04d", r.At().Year())
			},
			"yearday": func() string {
				return fmt.Sprintf("%03d", r.At().YearDay())
			},
			"hour": func() string {
				return fmt.Sprintf("%02d", r.At().Hour())
			},
			"minute": func() string {
				return fmt.Sprintf("%02d", r.At().Minute())
			},
			"second": func() string {
				return fmt.Sprintf("%02d", r.At().Second())
			},
			"tolower": func(s string) string {
				return strings.ToLower(s)
			},
			"toupper": func(s string) string {
				return strings.ToUpper(s)
			},
		}).Parse(path)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, r); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (r *Raw) Split(truncate time.Duration) []*Raw {

	cache := make(map[time.Time][]Reading)
	for _, v := range r.Readings {
		t := v.Timestamp.Truncate(truncate)

		cache[t] = append(cache[t], v)
	}

	var res []*Raw
	for k, v := range cache {
		sort.Slice(v, func(i, j int) bool {
			return v[i].At().Before(v[j].At())
		})

		res = append(res, &Raw{
			Label:     r.Label,
			Precision: r.Precision,
			Timestamp: k,
			Readings:  v,
		})
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].At().Before(res[j].At())
	})

	return res
}

func (r *Raw) Store(base, path string, truncate time.Duration) error {
	for _, f := range r.Split(truncate) {
		basename, err := f.Filename(path)
		if err != nil {
			return err
		}

		filename := filepath.Join(base, string(basename))

		if _, err := os.Stat(filename); err == nil {
			data, err := readFile(filename)
			if err != nil {
				return err
			}
			if err := f.Merge(data); err != nil {
				return err
			}
		}

		data, err := f.Marshal()
		if err != nil {
			return err
		}

		if err := writeFile(filename, data); err != nil {
			return err
		}
	}

	return nil
}
