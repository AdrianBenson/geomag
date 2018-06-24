package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ozym/geomag/internal/fdsn"
	"github.com/ozym/geomag/internal/geomag"
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
		fmt.Fprintf(os.Stderr, "Build geomag benmore processing files\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "  %s [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "General Options:\n")
		fmt.Fprintf(os.Stderr, "\n")
		flag.PrintDefaults()
	}

	var network string
	flag.StringVar(&network, "network", "NZ", "stream network code")

	var station string
	flag.StringVar(&station, "station", "SMHS", "stream station code")

	var location string
	flag.StringVar(&location, "location", "50", "stream location code")

	var service string
	flag.StringVar(&service, "service", "https://beta-service-nrt.geonet.org.nz", "fdsn service to use")

	var timeout time.Duration
	flag.DurationVar(&timeout, "timeout", time.Minute, "timeout for FDSN connections")

	var endtime string
	flag.StringVar(&endtime, "endtime", "", "time to process to")

	var length time.Duration
	flag.DurationVar(&length, "length", time.Hour, "length of time to process")

	var offset time.Duration
	flag.DurationVar(&offset, "offset", 0, "offset from length")

	var interval time.Duration
	flag.DurationVar(&interval, "interval", 0, "interval to process")

	var delay time.Duration
	flag.DurationVar(&delay, "delay", 0, "add delay to end time")

	var base string
	flag.StringVar(&base, "base", ".", "base directory")

	var label string
	flag.StringVar(&label, "label", "fge-benmore", "label to use")

	var volts float64
	flag.Float64Var(&volts, "volts", math.Pow(2.0, 24)/40, "sample scale factor")

	flag.Parse()

	if volts == 0 {
		log.Fatalf("volts argument cannot be zero")
	}

	leader := strings.Join([]string{network, station, location, "LFZ"}, "_")

	client := fdsn.NewClient(timeout)

	for {
		tock := func() time.Time {
			if endtime != "" {
				t, err := time.Parse(timeFormat, endtime)
				if err != nil {
					log.Fatalf("invalid end time %s: %v", endtime, err)
				}
				return t
			}

			<-time.After(ticks(interval, offset))

			return time.Now().UTC().Add(-delay)
		}()

		var values []geomag.Value
		r, err := (fdsn.Source{
			Network:  network,
			Station:  station,
			Location: location,
			Channel:  "LFZ",
		}).Request(service, tock, length)
		if err != nil {
			log.Fatalf("invalid service %s: %v", service, err)
		}
		if err := client.Sample(r, func(s string, t time.Time, v int32) error {
			values = append(values, geomag.Value{
				Timestamp: t,
				Field:     float64(v) / volts,
			})
			return nil
		}); err != nil {
			log.Fatalf("invalid query %s: %v", r, err)
		}

		for _, g := range geomag.NewVertical(leader, values) {
			path := filepath.Join(base, g.Filename(label))
			if err := g.WriteFile(path); err != nil {
				log.Fatalf("unable to store file %s: %v", path, err)
			}
		}

		if endtime != "" {
			break
		}

		if !(interval > 0) {
			break
		}
	}
}
