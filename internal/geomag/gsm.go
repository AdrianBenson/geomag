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

// Absolute represents and GSM field reading.
type Absolute struct {
	Timestamp time.Time
	Field     float64
	Quality   float64
}

// Gsm represents a proton magnetometer and readings.
type Gsm struct {
	Timestamp time.Time
	Prefix    string
	Readings  []Absolute
}

// NewGsm builds a slice of hourly Gsm values.
func NewGsm(prefix string, readings []Absolute) []Gsm {

	cache := make(map[time.Time][]Absolute)
	for _, r := range readings {
		t := r.Timestamp.Truncate(time.Hour)

		cache[t] = append(cache[t], r)
	}

	var gsms []Gsm
	for k, v := range cache {
		gsms = append(gsms, Gsm{
			Timestamp: k,
			Prefix:    prefix,
			Readings:  v,
		})
	}
	sort.Slice(gsms, func(i, j int) bool {
		return gsms[i].Timestamp.Before(gsms[j].Timestamp)
	})

	return gsms
}

func (g Gsm) Filename(label string) string {
	return strings.Join([]string{
		fmt.Sprintf("%04d", g.Timestamp.Year()),
		fmt.Sprintf("%04d.%03d", g.Timestamp.Year(), g.Timestamp.YearDay()),
		fmt.Sprintf("%04d.%03d.%02d%02d.%02d.%s.raw",
			g.Timestamp.Year(),
			g.Timestamp.YearDay(),
			g.Timestamp.Hour(),
			g.Timestamp.Minute(),
			g.Timestamp.Second(),
			strings.ToLower(strings.TrimSpace(label)),
		),
	}, "/")
}

func (g Gsm) Encode() ([]byte, error) {

	var lines []string
	for _, r := range g.Readings {
		lines = append(lines, strings.Join([]string{
			g.Prefix,
			r.Timestamp.Format("2006-01-02 15:04:05.000"),
			fmt.Sprintf("%.1f", r.Field),
			fmt.Sprintf("%.1f", float64(r.Quality)),
		}, ", "))
	}

	return []byte(strings.Join(lines, "\n") + "\n"), nil
}

func (g *Gsm) Decode(data []byte) error {

	scanner := bufio.NewScanner(bytes.NewBuffer(data))
	for scanner.Scan() {
		if parts := strings.Split(scanner.Text(), ","); len(parts) > 3 {
			g.Prefix = strings.TrimSpace(parts[0])

			t, err := time.Parse("2006-01-02 15:04:05.000", strings.TrimSpace(parts[1]))
			if err != nil {
				return err
			}
			f, err := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
			if err != nil {
				return err
			}
			q, err := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)
			if err != nil {
				return err
			}

			g.Readings = append(g.Readings, Absolute{
				Timestamp: t,
				Field:     f,
				Quality:   q,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func (g *Gsm) Merge(gsm Gsm) error {

	cache := make(map[time.Time]Absolute)
	for _, r := range gsm.Readings {
		cache[r.Timestamp] = r
	}
	for _, r := range g.Readings {
		cache[r.Timestamp] = r
	}

	var readings []Absolute
	for _, v := range cache {
		readings = append(readings, v)
	}

	sort.Slice(readings, func(i, j int) bool {
		return readings[i].Timestamp.Before(readings[j].Timestamp)
	})

	g.Readings = readings

	return nil
}

func (g *Gsm) ReadFile(path string) error {

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	if err := g.Decode(data); err != nil {
		return err
	}

	return nil
}

func (g *Gsm) WriteFile(path string) error {

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	if _, err := os.Stat(path); err == nil {
		var gsm Gsm
		if err := gsm.ReadFile(path); err != nil {
			return err
		}
		if err := g.Merge(gsm); err != nil {
			return err
		}
	}

	data, err := g.Encode()
	if err != nil {
		return err
	}

	if err := writeFile(path, data); err != nil {
		return err
	}

	return nil
}
