package gm

import (
	"bytes"
	"encoding/csv"
	"sort"
	"strconv"
	"strings"
	"time"
)

const rawFormat = "2006-01-02 15:04:05.000000Z"

type Reading struct {
	Timestamp time.Time
	Label     string
	Field     []float64
}

func NewReading(t time.Time, l string, v []float64) Reading {
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

func (r Reading) Values() []float64 {
	return r.Field
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

func (r Raw) At() time.Time {
	return r.Timestamp
}

func (r Raw) Tag() string {
	return r.Label
}

func (r *Raw) Store(base, path string, truncate time.Duration) error {
	return Store(base, path, truncate, r)
}

func (r Raw) Split(truncate time.Duration) []Formatter {

	cache := make(map[time.Time][]Reading)
	for _, v := range r.Readings {
		t := v.Timestamp.Truncate(truncate)

		cache[t] = append(cache[t], v)
	}

	var res []Formatter
	for k, v := range cache {
		sort.Slice(v, func(i, j int) bool {
			return v[i].At().Before(v[j].At())
		})

		obs := r
		obs.Timestamp = k
		obs.Readings = v

		res = append(res, &obs)
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].At().Before(res[j].At())
	})

	return res
}

func (r Raw) Encode() ([]byte, error) {

	sort.Slice(r.Readings, func(i, j int) bool {
		return r.Readings[i].Timestamp.Before(r.Readings[j].Timestamp)
	})

	var lines [][]string
	for _, v := range r.Readings {
		values := []string{
			v.Timestamp.Format(rawFormat),
			v.Label,
		}
		for _, f := range v.Field {
			values = append(values, strconv.FormatFloat(f, 'f', r.Precision, 64))
		}
		lines = append(lines, values)
	}

	var buf bytes.Buffer

	w := csv.NewWriter(&buf)

	w.WriteAll(lines)

	if err := w.Error(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (r *Raw) Decode(data []byte) error {

	rd := csv.NewReader(bytes.NewReader(data))

	records, err := rd.ReadAll()
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

		var values []float64

		for _, f := range l[2:] {
			v, err := strconv.ParseFloat(f, 64)
			if err != nil {
				return err
			}
			values = append(values, v)
		}

		r.Readings = append(r.Readings, NewReading(t, l[1], values))
	}

	return nil
}

func (r *Raw) Merge(data []byte) error {

	var raw Raw
	if err := raw.Decode(data); err != nil {
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
