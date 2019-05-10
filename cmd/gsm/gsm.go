package main

import (
	"sort"
	"time"
)

const gsmTimeFormat = "2006-01-02 15:04:05.000"

// Absolute represents and GSM field reading.
type Absolute struct {
	Timestamp time.Time
	Field     float64
	Quality   float64
}

// Gsm represents a proton magnetometer and readings.
type Gsm struct {
	Label  string
	Prefix string

	Timestamp time.Time
	Readings  []Absolute
}

// NewGsm builds a slice of hourly Gsm values.
func (g Gsm) Split(readings []Absolute, truncate time.Duration) []Gsm {

	cache := make(map[time.Time][]Absolute)
	for _, r := range readings {
		t := r.Timestamp.Truncate(truncate)

		cache[t] = append(cache[t], r)
	}

	var res []Gsm
	for k, v := range cache {
		obs := g

		obs.Timestamp = k
		obs.Readings = v

		res = append(res, obs)
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].Timestamp.Before(res[j].Timestamp)
	})

	return res
}
