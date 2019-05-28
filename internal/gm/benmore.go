package gm

import (
	"bufio"
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

const benmoreFormat = "2006-01-02 15:04:05.000"

type Benmore struct {
	Label string

	Timestamp time.Time
	Readings  []Reading
}

func NewBenmore(label string) *Benmore {
	return &Benmore{
		Label: label,
	}
}

func (b *Benmore) Add(v Reading) {
	if t := v.Timestamp; b.Timestamp.IsZero() || b.Timestamp.After(t) {
		b.Timestamp = t
	}
	b.Readings = append(b.Readings, v)
}

func (b Benmore) At() time.Time {
	return b.Timestamp
}

func (b Benmore) Tag() string {
	return b.Label
}

func (b *Benmore) Store(base, path string, truncate time.Duration) error {
	return Store(base, path, truncate, b)
}

func (b Benmore) Split(truncate time.Duration) []Formatter {

	cache := make(map[time.Time][]Reading)
	for _, r := range b.Readings {
		t := r.Timestamp.Truncate(truncate)

		cache[t] = append(cache[t], r)
	}

	var res []Formatter
	for k, v := range cache {
		obs := b

		obs.Timestamp = k
		obs.Readings = v

		res = append(res, &obs)
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].At().Before(res[j].At())
	})

	return res
}

func (b Benmore) Encode() ([]byte, error) {

	sort.Slice(b.Readings, func(i, j int) bool {
		return b.Readings[i].Timestamp.Before(b.Readings[j].Timestamp)
	})

	var lines []string
	for _, r := range b.Readings {
		lines = append(lines, strings.Join([]string{
			r.Label,
			r.Timestamp.Format(benmoreFormat),
			fmt.Sprintf("%.8f", r.Field),
		}, ", "))
	}

	return []byte(strings.Join(lines, "\n") + "\n"), nil
}

func (b *Benmore) Decode(data []byte) error {

	scanner := bufio.NewScanner(bytes.NewBuffer(data))
	for scanner.Scan() {
		if parts := strings.Split(scanner.Text(), ","); len(parts) > 2 {

			t, err := time.Parse("2006-01-02 15:04:05.000", strings.TrimSpace(parts[1]))
			if err != nil {
				return err
			}
			f, err := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
			if err != nil {
				return err
			}

			b.Readings = append(b.Readings, Reading{
				Timestamp: t,
				Label:     strings.TrimSpace(parts[0]),
				Field:     f,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func (b *Benmore) Merge(data []byte) error {

	var benmore Benmore
	if err := benmore.Decode(data); err != nil {
		return err
	}

	cache := make(map[time.Time]Reading)
	for _, r := range benmore.Readings {
		cache[r.Timestamp] = r
	}
	for _, r := range b.Readings {
		cache[r.Timestamp] = r
	}

	var readings []Reading
	for _, v := range cache {
		readings = append(readings, v)
	}

	sort.Slice(readings, func(i, j int) bool {
		return readings[i].Timestamp.Before(readings[j].Timestamp)
	})

	b.Readings = readings

	return nil
}
