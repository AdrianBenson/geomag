package main

import (
	"math"
	"sort"
	"time"
)

// Full represents a full field reading with sensor, driver temperatures,
// and a benmore (DC) correction (which is generally not used).
type Full struct {
	Timestamp time.Time
	Field     [3]float64
	Driver    float64
	Sensor    float64
	Benmore   float64
}

// Fluxgate represents a FGE magnetometer and associated settings, and
// full field recordings for a given time interval.
type Fluxgate struct {
	Label    string
	Code     string
	Sensor   int
	Driver   int
	Bias     [3]int
	Coil     [3]float64
	Res      [3]float64
	E        [5]float64
	Step     float64
	Gain     float64
	Offset   [3]float64
	Polarity [3]float64
	Model    string

	Timestamp time.Time
	Readings  []Full
}

// NewFluxgate splits a set of full field recordings into a slice
// of Flugate elements with the Timestamp and Readings fields
// filled. Other Fluxgate settings will need to be added externally.
func (f Fluxgate) Split(readings []Full, truncate time.Duration) []Fluxgate {

	cache := make(map[time.Time][]Full)
	for _, r := range readings {
		t := r.Timestamp.Truncate(truncate)

		cache[t] = append(cache[t], r)
	}

	var flux []Fluxgate
	for k, v := range cache {
		obs := f

		obs.Timestamp = k
		obs.Readings = v

		flux = append(flux, obs)
	}

	sort.Slice(flux, func(i, j int) bool {
		return flux[i].Timestamp.Before(flux[j].Timestamp)
	})

	return flux
}

// toCalc converts to calculated units from raw units
func (f Fluxgate) toCalc(v [3]float64) [3]float64 {
	var r [3]float64
	for i := 0; i < 3; i++ {
		if f.Res[i] == 0.0 {
			continue
		}
		r[i] = f.Polarity[i] * (f.Coil[i]*(v[i]/f.Res[i]+f.Step*float64(f.Bias[i])) + f.Offset[i])
	}
	return r
}

// toF builts the full field from calculated units
func (f Fluxgate) toF(v [3]float64) float64 {
	var sum2 float64
	for i := 0; i < 3; i++ {
		sum2 += v[i] * v[i]
	}
	return math.Sqrt(sum2)
}

// toI builds the declination from calculated units
func (f Fluxgate) toI(v [3]float64) float64 {
	angle := 180.0 * math.Atan2(v[2], v[0]) / math.Pi
	if angle < -180.0 {
		angle += 360
	}
	if angle > 180.0 {
		angle -= 360
	}
	return angle
}
