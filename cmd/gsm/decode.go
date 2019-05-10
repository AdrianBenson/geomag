package main

import (
	"bufio"
	"bytes"
	"sort"
	"strconv"
	"strings"
	"time"
)

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

func (g *Gsm) Merge(data []byte) error {

	var gsm Gsm
	if err := gsm.Decode(data); err != nil {
		return err
	}

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
