package main

import (
	"bufio"
	"bytes"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Decode converts a byte slice into reading lines, all other fields are lef untouched.
func (f *Fluxgate) Decode(data []byte) error {

	scanner := bufio.NewScanner(bytes.NewBuffer(data))
	var line int
	var start time.Time
	for scanner.Scan() {
		if line = line + 1; line < 26 {
			if strings.HasPrefix(scanner.Text(), "epoch=") {
				ticks, err := strconv.ParseInt(strings.TrimPrefix(scanner.Text(), "epoch="), 10, 64)
				if err != nil {
					return err
				}
				start = time.Unix(ticks, 0).UTC()
			}
			continue
		}

		if parts := strings.Fields(scanner.Text()); len(parts) == 14 {
			min, err := strconv.Atoi(parts[1])
			if err != nil {
				return err
			}
			t := start.Add(time.Minute * time.Duration(min))

			sec, err := strconv.Atoi(parts[2])
			if err != nil {
				return err
			}
			t = t.Add(time.Second * time.Duration(sec))

			ts, err := strconv.ParseFloat(parts[3], 64)
			if err != nil {
				return err
			}
			td, err := strconv.ParseFloat(parts[4], 64)
			if err != nil {
				return err
			}

			x, err := strconv.ParseFloat(parts[5], 64)
			if err != nil {
				return err
			}

			y, err := strconv.ParseFloat(parts[6], 64)
			if err != nil {
				return err
			}

			z, err := strconv.ParseFloat(parts[7], 64)
			if err != nil {
				return err
			}

			b, err := strconv.ParseFloat(parts[13], 64)
			if err != nil {
				return err
			}

			f.Readings = append(f.Readings, Full{
				Timestamp: t,
				Driver:    td,
				Sensor:    ts,
				Field:     [3]float64{x, y, z},
				Benmore:   b,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

// Merge will insert readings from one Fluxgate representation into another.
func (f *Fluxgate) Merge(data []byte) error {

	var flux Fluxgate
	if err := flux.Decode(data); err != nil {
		return err
	}

	cache := make(map[time.Time]Full)
	for _, r := range flux.Readings {
		cache[r.Timestamp] = r
	}
	for _, r := range f.Readings {
		cache[r.Timestamp] = r
	}

	var readings []Full
	for _, v := range cache {
		readings = append(readings, v)
	}

	sort.Slice(readings, func(i, j int) bool {
		return readings[i].Timestamp.Before(readings[j].Timestamp)
	})

	f.Readings = readings

	return nil
}
