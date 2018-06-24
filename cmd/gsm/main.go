package main

import (
	"flag"
	"fmt"
	"log"
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
		fmt.Fprintf(os.Stderr, "Build geomag gsm processing files\n")
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
	flag.StringVar(&station, "station", "", "stream station code")

	var location string
	flag.StringVar(&location, "location", "51", "stream location code")

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
	flag.DurationVar(&delay, "delay", 0, "delay to remove from processing endtime ")

	var base string
	flag.StringVar(&base, "base", ".", "base directory")

	var label string
	flag.StringVar(&label, "label", "unknown", "label to use")

	flag.Parse()

	client := fdsn.NewClient(timeout)

	leader := strings.Join([]string{network, station, location, "LFF"}, "_")

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

		obs := make(map[time.Time][]int32)
		for _, c := range []string{"LFF", "LEQ"} {
			r, err := (fdsn.Source{
				Network:  network,
				Station:  station,
				Location: location,
				Channel:  c,
			}).Request(service, tock, length)
			if err != nil {
				log.Fatalf("invalid service %s: %v", service, err)
			}
			if err := client.Sample(r, func(s string, t time.Time, v int32) error {
				obs[t] = append(obs[t], v)
				return nil
			}); err != nil {
				log.Fatalf("invalid query %s: %v", r, err)
			}
		}

		var abs []geomag.Absolute
		for t, v := range obs {
			if len(v) != 2 {
				continue
			}

			abs = append(abs, geomag.Absolute{
				Timestamp: t,
				Field:     float64(v[0]),
				Quality:   float64(v[1]),
			})
		}

		for _, g := range geomag.NewGsm(leader, abs) {
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
