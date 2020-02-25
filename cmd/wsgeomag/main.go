package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/GeoNet/kit/mseed"
	"github.com/nightlyone/lockfile"

	"github.com/AdrianBenson/geomag/internal/raw"
	// "github.com/ozym/geomag/internal/raw"
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
		fmt.Fprintf(os.Stderr, "Build geomag raw processing files via fdsn wave services\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "  %s [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "General Options:\n")
		fmt.Fprintf(os.Stderr, "\n")
		flag.PrintDefaults()
	}

	var verbose bool
	flag.BoolVar(&verbose, "verbose", false, "make noise")

	var lock string
	flag.StringVar(&lock, "lockfile", "", "provide a process lock file")

	var streams string
	flag.StringVar(&streams, "streams", "", "comma delimited channel srcname(s)")

	var service string
	flag.StringVar(&service, "service", "https://service-nrt.geonet.org.nz", "fdsn service to use")

	var timeout time.Duration
	flag.DurationVar(&timeout, "timeout", time.Minute, "timeout for FDSN connections")

	var endtime string
	flag.StringVar(&endtime, "endtime", "", "optional time to process to, implies empty interval")

	var starttime string
	flag.StringVar(&starttime, "starttime", "", "optional time to process from, implies empty interval")

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

	var gain float64
	flag.Float64Var(&gain, "gain", 1.0, "gain to apply to raw data")

	flag.Parse()

	if lock != "" {
		lf, err := lockfile.New(lock)
		if err != nil {
			log.Fatalf("unable to open lockfile %s: %v", lock, err)
		}
		if err := lf.TryLock(); err != nil {
			if verbose {
				log.Printf("unable to lock file %q: %v", lf, err)
			}
			os.Exit(1)
		}
		defer lf.Unlock()
	}

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

	client := NewDataselect(service, timeout)
	msr := mseed.NewMSRecord()
	defer mseed.FreeMSRecord(msr)

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

		if verbose {
			log.Printf("query: %s from %v to %v", strings.Join(srcnames, ","), t.Add(-dt), t)
		}

		data, err := client.Query(srcnames, t, dt)
		if err != nil {
			log.Fatalf("unable to query fdsn service: %v", err)
		}

		cache := make(map[string]*raw.Raw)

		for n := 0; n < len(data)/512; n++ {
			if err := msr.Unpack(data[n*512:(n+1)*512], 512, 1, 0); err != nil {
				log.Printf("skipping block, unable to unpack block: (%d) %v", n, err)
				continue
			}

			srcname := msr.SrcName(0)

			sps := float64(msr.Samprate())
			if !(sps > 0) {
				log.Printf("skipping block, invalid sample rate: (%s) %g", srcname, sps)
				continue
			}

			dt := time.Duration(float64(time.Second) / sps)

			samples, err := msr.DataSamples()
			if err != nil {
				log.Printf("skipping block, unable to decode samples: (%s) %v", srcname, err)
				continue
			}

			for n, s := range samples {
				t := msr.Starttime().Add(time.Duration(n) * dt)

				if _, ok := cache[srcname]; !ok {
					cache[srcname] = raw.NewRaw(srcname, dp)
				}

				if r, ok := cache[srcname]; ok {
					r.Add(raw.NewReading(t, srcname, gain*float64(s)))
				}
			}
		}

		for _, v := range cache {
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
