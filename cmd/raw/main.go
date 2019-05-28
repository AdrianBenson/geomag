package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
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
		fmt.Fprintf(os.Stderr, "Build geomag raw processing files via fdsn\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "  %s [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "General Options:\n")
		fmt.Fprintf(os.Stderr, "\n")
		flag.PrintDefaults()
	}

	var streams string
	flag.StringVar(&streams, "streams", "", "comma delimited channel srcname(s)")

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

	var truncate time.Duration
	flag.DurationVar(&truncate, "truncate", time.Hour, "time interval to split files into")

	var path string
	flag.StringVar(&path, "path", "{{year}}/{{year}}.{{yearday}}/{{year}}.{{yearday}}.{{hour}}{{minute}}.{{second}}.{{toupper .Label}}.csv", "file name template")

	var dp int
	flag.IntVar(&dp, "dp", 0, "number of decimal places for raw data")

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

	var srcnames []string
	for _, s := range strings.Split(streams, ",") {
		srcnames = append(srcnames, strings.TrimSpace(s))
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

		obs, err := client.Query(srcnames, t, dt)
		if err != nil {
			log.Fatalf("unable to query fdsn service: %v", err)
		}

		raw := make(map[string]*gm.Raw)
		for t, x := range obs {
			for i, v := range x {
				if !(i < len(srcnames)) {
					continue
				}
				srcname := srcnames[i]

				if _, ok := raw[srcname]; !ok {
					raw[srcname] = gm.NewRaw(srcname, dp)
				}

				if r, ok := raw[srcname]; ok {
					r.Add(gm.NewReading(t, srcname, float64(v)))
				}
			}
		}

		for _, v := range raw {
			if err := v.Store(base, path, truncate); err != nil {
				log.Fatalf("unable to store observations: %v", err)
			}
		}

		if !st.IsZero() || !et.IsZero() {
			break
		}

		if !(interval > 0) {
			break
		}
	}
}
