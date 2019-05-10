package main

import (
	"sort"
	"time"
)

const benmoreFormat = "2006-01-02 15:04:05.000"

type Value struct {
	Timestamp time.Time
	Field     float64
}

type Benmore struct {
	Label string

	Timestamp time.Time
	Prefix    string
	Readings  []Value
}

func (b Benmore) Split(readings []Value, truncate time.Duration) []Benmore {

	cache := make(map[time.Time][]Value)
	for _, r := range readings {
		t := r.Timestamp.Truncate(truncate)

		cache[t] = append(cache[t], r)
	}

	var res []Benmore
	for k, v := range cache {
		obs := b

		obs.Timestamp = k
		obs.Readings = v

		res = append(res, obs)
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].Timestamp.Before(res[j].Timestamp)
	})

	return res
}
