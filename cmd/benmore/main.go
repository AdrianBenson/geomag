package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"github.com/ozym/geomag/internal/ds"
	"github.com/ozym/geomag/internal/gm"
)

const timeFormat = "2006-01-02T15:04:05"

func ticks(d, o time.Duration) time.Duration {
	t := time.Now().UTC()
	switch {
	case d > 0 && o > 0:
		return t.Add(d+o).Truncate(d).Add(o).Sub(t) % d
	case d > 0:
		return t.Add(d).Truncate(d).Sub(t)
	default:
		return 0
	}
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Build geomag benmore processing files via fdsn\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "  %s [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "General Options:\n")
		fmt.Fprintf(os.Stderr, "\n")
		flag.PrintDefaults()
	}

	var fz string
	flag.StringVar(&fz, "fz", "", "benmore z channel stream srcname")

	var service string
	flag.StringVar(&service, "service", "https://service-nrt.geonet.org.nz", "fdsn service to use")

	var timeout time.Duration
	flag.DurationVar(&timeout, "timeout", time.Minute, "timeout for FDSN connections")

	var endtime string
	flag.StringVar(&endtime, "endtime", "", "optional time to process to, implies empty interval")

	var starttime string
	flag.StringVar(&endtime, "starttime", "", "optional time to process from, implies empty interval")

	var length time.Duration
	flag.DurationVar(&length, "length", time.Hour, "length of time to process")

	var offset time.Duration
	flag.DurationVar(&offset, "offset", 0, "offset from length")

	var interval time.Duration
	flag.DurationVar(&interval, "interval", 0, "interval to process continuously")

	var delay time.Duration
	flag.DurationVar(&delay, "delay", 0, "delay to remove from processing endtime ")

	var base string
	flag.StringVar(&base, "base", ".", "base directory")

	var label string
	flag.StringVar(&label, "label", "unknown", "label to use")

	var volts float64
	flag.Float64Var(&volts, "volts", math.Pow(2.0, 24)/40, "sample scale factor")

	var path string
	flag.StringVar(&path, "path", "{{year}}/{{year}}.{{yearday}}/{{year}}.{{yearday}}.{{hour}}{{minute}}.{{second}}.{{tolower .Label}}.raw", "file name template")

	var truncate time.Duration
	flag.DurationVar(&truncate, "truncate", time.Hour, "time interval to split files into")

	flag.Parse()

	var err error
	var st, et time.Time
	if starttime != "" {
		if st, err = time.Parse(timeFormat, starttime); err != nil {
			log.Fatalf("invalid starttime %s: %v", starttime, err)
		}
	}
	if endtime != "" {
		if et, err = time.Parse(timeFormat, endtime); err != nil {
			log.Fatalf("invalid endtime %s: %v", endtime, err)
		}
	}

	client := ds.NewDataselect(service, timeout)

	for {
		t, dt := func() (time.Time, time.Duration) {
			switch {
			case !st.IsZero() && !et.IsZero():
				return et, et.Sub(st)
			case !st.IsZero():
				return st.Add(length), length
			case !et.IsZero():
				return et, length
			default:
				t := <-time.After(ticks(interval, offset))
				return t.UTC().Add(-delay), length
			}
		}()

		obs, err := client.Query([]string{fz}, t, dt)
		if err != nil {
			log.Fatalf("unable to query fdsn service: %v", err)
		}

		raw := make(map[string]*gm.Benmore)
		for t, x := range obs {
			for i, v := range x {
				if !(i < 1) {
					continue
				}

				if _, ok := raw[fz]; !ok {
					raw[fz] = gm.NewBenmore(label)
				}

				if r, ok := raw[fz]; ok {
					r.Add(gm.NewReading(t, fz, float64(v)/volts))
				}
			}
		}

		for _, v := range raw {
			if err := v.Store(base, path, truncate); err != nil {
				log.Fatalf("unable to store observations: %v", err)
			}
		}

		if !(interval > 0) {
			break
		}
	}
}
