package gm

import (
	"bufio"
	"bytes"
	//	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

const gsmTimeFormat = "2006-01-02 15:04:05.000"

type Gsm struct {
	Label string

	Timestamp time.Time
	Readings  []Reading
}

func NewGsm(label string) *Gsm {
	return &Gsm{
		Label: label,
	}
}

func (g *Gsm) Add(v Reading) {
	if t := v.Timestamp; g.Timestamp.IsZero() || g.Timestamp.After(t) {
		g.Timestamp = t
	}
	g.Readings = append(g.Readings, v)
}

func (g Gsm) At() time.Time {
	return g.Timestamp
}

func (g Gsm) Tag() string {
	return g.Label
}

func (g *Gsm) Store(base, path string, truncate time.Duration) error {
	return Store(base, path, truncate, g)
}

// NewGsm builds a slice of hourly Gsm values.
func (g Gsm) Split(truncate time.Duration) []Formatter {

	cache := make(map[time.Time][]Reading)
	for _, r := range g.Readings {
		t := r.Timestamp.Truncate(truncate)

		cache[t] = append(cache[t], r)
	}

	var res []Formatter
	for k, v := range cache {
		obs := g

		obs.Timestamp = k
		obs.Readings = v

		res = append(res, &obs)
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].At().Before(res[j].At())
	})

	return res
}

func (g Gsm) Encode() ([]byte, error) {

	sort.Slice(g.Readings, func(i, j int) bool {
		return g.Readings[i].Timestamp.Before(g.Readings[j].Timestamp)
	})

	var lines []string
	for _, r := range g.Readings {
		lines = append(lines, strings.Join([]string{
			r.Label,
			r.Timestamp.Format(gsmTimeFormat),
			//fmt.Sprintf("%.1f", r.Field),
			//fmt.Sprintf("%.1f", float64(r.Quality)),
		}, ", "))
	}

	return []byte(strings.Join(lines, "\n") + "\n"), nil

}

func (g *Gsm) Decode(data []byte) error {

	scanner := bufio.NewScanner(bytes.NewBuffer(data))
	for scanner.Scan() {
		if parts := strings.Split(scanner.Text(), ","); len(parts) > 3 {

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

			g.Readings = append(g.Readings, NewReading(t, strings.TrimSpace(parts[0]), []float64{f, q}))
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func (g *Gsm) Merge(data []byte) error {

	var gsm Gsm
	if err := gsm.Decode(data); err != nil {
		return err
	}

	cache := make(map[time.Time]Reading)
	for _, r := range gsm.Readings {
		cache[r.Timestamp] = r
	}
	for _, r := range g.Readings {
		cache[r.Timestamp] = r
	}

	var readings []Reading
	for _, v := range cache {
		readings = append(readings, v)
	}

	sort.Slice(readings, func(i, j int) bool {
		return readings[i].Timestamp.Before(readings[j].Timestamp)
	})

	g.Readings = readings

	return nil
}
