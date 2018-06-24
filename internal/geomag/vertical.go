package geomag

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Value struct {
	Timestamp time.Time
	Field     float64
}

type Vertical struct {
	Timestamp time.Time
	Prefix    string
	Readings  []Value
}

func NewVertical(prefix string, readings []Value) []Vertical {

	cache := make(map[time.Time][]Value)
	for _, r := range readings {
		t := r.Timestamp.Truncate(time.Hour)

		cache[t] = append(cache[t], r)
	}

	var verts []Vertical
	for k, v := range cache {
		verts = append(verts, Vertical{
			Timestamp: k,
			Prefix:    prefix,
			Readings:  v,
		})
	}
	sort.Slice(verts, func(i, j int) bool {
		return verts[i].Timestamp.Before(verts[j].Timestamp)
	})

	return verts
}

func (v Vertical) Filename(label string) string {
	return strings.Join([]string{
		fmt.Sprintf("%04d", v.Timestamp.Year()),
		fmt.Sprintf("%04d.%03d", v.Timestamp.Year(), v.Timestamp.YearDay()),
		fmt.Sprintf("%04d.%03d.%02d%02d.%02d.%s.raw",
			v.Timestamp.Year(),
			v.Timestamp.YearDay(),
			v.Timestamp.Hour(),
			v.Timestamp.Minute(),
			v.Timestamp.Second(),
			strings.ToLower(strings.TrimSpace(label)),
		),
	}, "/")
}

func (v Vertical) Encode() ([]byte, error) {

	var lines []string
	for _, r := range v.Readings {
		lines = append(lines, strings.Join([]string{
			v.Prefix,
			r.Timestamp.Format("2006-01-02 15:04:05.000"),
			fmt.Sprintf("%.8f", r.Field),
		}, ", "))
	}

	return []byte(strings.Join(lines, "\n") + "\n"), nil
}

func (v *Vertical) Decode(data []byte) error {

	scanner := bufio.NewScanner(bytes.NewBuffer(data))
	for scanner.Scan() {
		if parts := strings.Split(scanner.Text(), ","); len(parts) > 2 {
			v.Prefix = strings.TrimSpace(parts[0])

			t, err := time.Parse("2006-01-02 15:04:05.000", strings.TrimSpace(parts[1]))
			if err != nil {
				return err
			}
			f, err := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
			if err != nil {
				return err
			}

			v.Readings = append(v.Readings, Value{
				Timestamp: t,
				Field:     f,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func (v *Vertical) Merge(vert Vertical) error {

	cache := make(map[time.Time]Value)
	for _, r := range vert.Readings {
		cache[r.Timestamp] = r
	}
	for _, r := range v.Readings {
		cache[r.Timestamp] = r
	}

	var readings []Value
	for _, v := range cache {
		readings = append(readings, v)
	}

	sort.Slice(readings, func(i, j int) bool {
		return readings[i].Timestamp.Before(readings[j].Timestamp)
	})

	v.Readings = readings

	return nil
}

func (v *Vertical) ReadFile(path string) error {

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	if err := v.Decode(data); err != nil {
		return err
	}

	return nil
}

func (v *Vertical) WriteFile(path string) error {

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	if _, err := os.Stat(path); err == nil {
		var vert Vertical
		if err := vert.ReadFile(path); err != nil {
			return err
		}
		if err := v.Merge(vert); err != nil {
			return err
		}
	}

	data, err := v.Encode()
	if err != nil {
		return err
	}

	if err := writeFile(path, data); err != nil {
		return err
	}

	return nil
}
