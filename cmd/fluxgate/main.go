package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
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
		fmt.Fprintf(os.Stderr, "Build geomag fluxgate processing files\n")
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
	flag.StringVar(&location, "location", "50", "stream location code")

	var benmore string
	flag.StringVar(&benmore, "benmore", "", "benmore stream station code")

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

	var volts float64
	flag.Float64Var(&volts, "volts", math.Pow(2.0, 24)/40, "sample scale factor")

	var gain float64
	flag.Float64Var(&gain, "gain", 1280, "sample scale factor")

	var scale float64
	flag.Float64Var(&scale, "scale", 200, "conversion unit for temperature")

	var step float64
	flag.Float64Var(&step, "step", 0.003922, "gain step")

	var sensor int
	flag.IntVar(&sensor, "sensor", 0, "sensor id")

	var driver int
	flag.IntVar(&driver, "driver", 0, "driver id")

	var xbias int
	flag.IntVar(&xbias, "xbias", 0, "sensor x bias")

	var ybias int
	flag.IntVar(&ybias, "ybias", 0, "sensor x bias")

	var zbias int
	flag.IntVar(&zbias, "zbias", 0, "sensor x bias")

	var xcoil float64
	flag.Float64Var(&xcoil, "xcoil", 0, "sensor x coil")

	var ycoil float64
	flag.Float64Var(&ycoil, "ycoil", 0, "sensor x coil")

	var zcoil float64
	flag.Float64Var(&zcoil, "zcoil", 0, "sensor x coil")

	var xres float64
	flag.Float64Var(&xres, "xres", 0, "sensor x res")

	var yres float64
	flag.Float64Var(&yres, "yres", 0, "sensor x res")

	var zres float64
	flag.Float64Var(&zres, "zres", 0, "sensor x res")

	var e0 float64
	flag.Float64Var(&e0, "e0", 0, "sensor x e0")

	var e1 float64
	flag.Float64Var(&e1, "e1", 0, "sensor x e1")

	var e2 float64
	flag.Float64Var(&e2, "e2", 0, "sensor x e2")

	var e3 float64
	flag.Float64Var(&e3, "e3", 0, "sensor x e3")

	var e4 float64
	flag.Float64Var(&e4, "e4", 0, "sensor x e4")

	var zoffset float64
	flag.Float64Var(&zoffset, "zoffset", 0, "sensor x zoffset")

	var zpolarity float64
	flag.Float64Var(&zpolarity, "zpolarity", -1, "sensor x zpolarity")

	var model string
	flag.StringVar(&model, "model", "", "loger model")

	var code string
	flag.StringVar(&code, "code", "", "ite code")

	flag.Parse()

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

		obs := make(map[time.Time][]int32)
		for _, c := range []string{"LFX", "LFY", "LFZ", "LKD", "LKS"} {
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

		var full []geomag.Full
		for t, v := range obs {
			if len(v) != 5 {
				continue
			}
			full = append(full, geomag.Full{
				Timestamp: t.Truncate(time.Second),
				Field:     [3]float64{float64(v[0]) / volts, float64(v[1]) / volts, float64(v[2]) / volts},
				//TODO: bug?
				Driver: scale*float64(v[4])/volts - 273,
				Sensor: scale*float64(v[3])/volts - 273,
			})
		}

		for _, f := range geomag.NewFluxgate(full) {
			f.Label = code
			f.Sensor = sensor
			f.Driver = driver
			f.Bias = [3]int{xbias, ybias, zbias}
			f.Coil = [3]float64{xcoil, ycoil, zcoil}
			f.Res = [3]float64{xres, yres, zres}
			f.E = [5]float64{e0, e1, e2, e3, e4}
			f.Step = step
			f.Offset = [3]float64{0, 0, zoffset}
			f.Polarity = [3]float64{1, 1, zpolarity}
			f.Model = model
			f.Gain = gain // not used other than for template

			path := filepath.Join(base, f.Filename(label))
			if err := f.WriteFile(path); err != nil {
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
