package main

import (
	"bufio"
	"bytes"
	"sort"
	"strconv"
	"strings"
	"time"
)

func (b *Benmore) Decode(data []byte) error {

	scanner := bufio.NewScanner(bytes.NewBuffer(data))
	for scanner.Scan() {
		if parts := strings.Split(scanner.Text(), ","); len(parts) > 2 {
			b.Prefix = strings.TrimSpace(parts[0])

			t, err := time.Parse("2006-01-02 15:04:05.000", strings.TrimSpace(parts[1]))
			if err != nil {
				return err
			}
			f, err := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
			if err != nil {
				return err
			}

			b.Readings = append(b.Readings, Value{
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

func (b *Benmore) Merge(data []byte) error {

	var benmore Benmore
	if err := benmore.Decode(data); err != nil {
		return err
	}

	cache := make(map[time.Time]Value)
	for _, r := range benmore.Readings {
		cache[r.Timestamp] = r
	}
	for _, r := range b.Readings {
		cache[r.Timestamp] = r
	}

	var readings []Value
	for _, v := range cache {
		readings = append(readings, v)
	}

	sort.Slice(readings, func(i, j int) bool {
		return readings[i].Timestamp.Before(readings[j].Timestamp)
	})

	b.Readings = readings

	return nil
}
