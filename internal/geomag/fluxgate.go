package geomag

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
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
// full field recordings for a given hour.
type Fluxgate struct {
	Label    string
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
func NewFluxgate(readings []Full) []Fluxgate {

	cache := make(map[time.Time][]Full)
	for _, r := range readings {
		t := r.Timestamp.Truncate(time.Hour)

		cache[t] = append(cache[t], r)
	}

	var flux []Fluxgate
	for k, v := range cache {
		flux = append(flux, Fluxgate{
			Timestamp: k,
			Readings:  v,
		})
	}
	sort.Slice(flux, func(i, j int) bool {
		return flux[i].Timestamp.Before(flux[j].Timestamp)
	})

	return flux
}

// Filename builds a standard hourly file path name.
func (f Fluxgate) Filename(label string) string {
	return strings.Join([]string{
		fmt.Sprintf("%04d", f.Timestamp.Year()),
		fmt.Sprintf("%04d.%03d", f.Timestamp.Year(), f.Timestamp.YearDay()),
		fmt.Sprintf("%04d.%03d.%02d%02d.%02d.%s.txt",
			f.Timestamp.Year(),
			f.Timestamp.YearDay(),
			f.Timestamp.Hour(),
			f.Timestamp.Minute(),
			f.Timestamp.Second(),
			strings.ToLower(strings.TrimSpace(label)),
		),
	}, "/")
}

// Encode builds a byte slice containing the standard file header and reading lines.
func (f Fluxgate) Encode() ([]byte, error) {

	var lines []string

	var buf bytes.Buffer
	if err := f.Format(&buf); err != nil {
		return nil, err
	}

	for _, r := range f.Readings {
		var parts []string

		parts = append(parts, r.Timestamp.Format("15 04 05"))

		parts = append(parts, fmt.Sprintf("%7.2f", r.Driver))
		parts = append(parts, fmt.Sprintf("%6.2f", r.Sensor))

		parts = append(parts, fmt.Sprintf("%7.4f", r.Field[0]))
		parts = append(parts, fmt.Sprintf("%9.4f", r.Field[1]))
		parts = append(parts, fmt.Sprintf("%8.4f", r.Field[2]))

		calc := f.toCalc(r.Field)

		parts = append(parts, fmt.Sprintf("%10.3f", calc[0]))
		parts = append(parts, fmt.Sprintf("%9.3f", calc[1]))
		parts = append(parts, fmt.Sprintf("%10.3f", calc[2]))

		parts = append(parts, fmt.Sprintf("%10.3f", f.toF(calc)))
		parts = append(parts, fmt.Sprintf("%7.2f", f.toI(calc)))
		parts = append(parts, fmt.Sprintf("%8.4f", r.Benmore))

		lines = append(lines, strings.Join(parts, " "))
	}

	return append(buf.Bytes(), []byte(strings.Join(lines, "\n")+"\n")...), nil
}

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
func (f *Fluxgate) Merge(flux Fluxgate) error {

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

// ReadFile will read and decode a fluxgate file.
func (f *Fluxgate) ReadFile(path string) error {

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	if err := f.Decode(data); err != nil {
		return err
	}

	return nil
}

// WriteFile will encode and write a fluxgate file, existing values will be merged.
func (f *Fluxgate) WriteFile(path string) error {

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	if _, err := os.Stat(path); err == nil {
		var flux Fluxgate
		if err := flux.ReadFile(path); err != nil {
			return err
		}
		if err := f.Merge(flux); err != nil {
			return err
		}
	}

	data, err := f.Encode()
	if err != nil {
		return err
	}

	if err := writeFile(path, data); err != nil {
		return err
	}

	return nil
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
