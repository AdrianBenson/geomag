package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/GeoNet/kit/mseed"

	"github.com/ozym/geomag/internal/raw"
)

func main() {

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Build geomag data from MSEED file(s)\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "  %s [options] [options] <mseed ...>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "General Options:\n")
		fmt.Fprintf(os.Stderr, "\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n")
	}

	var truncate time.Duration
	flag.DurationVar(&truncate, "truncate", time.Hour, "interval to store files")

	var base string
	flag.StringVar(&base, "base", ".", "base output directory")

	var path string
	flag.StringVar(&path, "path", "{{year}}/{{year}}.{{yearday}}/{{year}}.{{yearday}}.{{hour}}{{minute}}.{{second}}.{{toupper .Label}}.csv", "file name template")

	var dp int
	flag.IntVar(&dp, "dp", 0, "number of decimal places for raw data")

	var gain float64
	flag.Float64Var(&gain, "gain", 1.0, "gain to apply to raw data")

	flag.Parse()

	if err := os.MkdirAll(filepath.Dir(base), 0755); err != nil {
		log.Fatalf("unable to create base parent directory %s: %v", base, err)
	}

	cache := make(map[string]*raw.Raw)

	msr := mseed.NewMSRecord()
	defer mseed.FreeMSRecord(msr)

	for _, f := range flag.Args() {
		data, err := ioutil.ReadFile(f)
		if err != nil {
			log.Fatalf("unable to read file %s: %v", f, err)
		}

		for n := 0; n < len(data)/512; n++ {
			if err := msr.Unpack(data[n*512:(n+1)*512], 512, 1, 0); err != nil {
				log.Printf("skipping block, unable to unpack block  %s: (%d) %v", f, n, err)
				continue
			}

			srcname := msr.SrcName(0)

			sps := float64(msr.Samprate())
			if !(sps > 0) {
				log.Printf("skipping block, invalid sample rate %s: (%s) %g", f, srcname, sps)
				continue
			}

			dt := time.Duration(float64(time.Second) / sps)

			samples, err := msr.DataSamples()
			if err != nil {
				log.Printf("skipping block, unable to decode samples %s: (%s) %v", f, srcname, err)
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
	}

	for _, v := range cache {
		if err := v.Store(base, path, truncate); err != nil {
			log.Fatalf("unable to store observations: %v", err)
		}
	}

}
