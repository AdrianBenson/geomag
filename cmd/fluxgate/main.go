package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"github.com/ozym/geomag/internal/ds"
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
		fmt.Fprintf(os.Stderr, "Build geomag fluxgate processing files via fdsn\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "  %s [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "General Options:\n")
		fmt.Fprintf(os.Stderr, "\n")
		flag.PrintDefaults()
	}

	var fx string
	flag.StringVar(&fx, "fx", "", "fluxgate x channel stream srcname")

	var fy string
	flag.StringVar(&fy, "fy", "", "fluxgate y channel stream srcname")

	var fz string
	flag.StringVar(&fz, "fz", "", "fluxgate z channel stream srcname")

	var kd string
	flag.StringVar(&kd, "kd", "", "driver channel stream srcname")

	var ks string
	flag.StringVar(&ks, "ks", "", "sensor channel stream srcname")

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
	flag.StringVar(&model, "model", "", "logger model")

	var code string
	flag.StringVar(&code, "code", "", "ite code")

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

	flux := Fluxgate{
		Label:    label,
		Code:     code,
		Sensor:   sensor,
		Driver:   driver,
		Bias:     [3]int{xbias, ybias, zbias},
		Coil:     [3]float64{xcoil, ycoil, zcoil},
		Res:      [3]float64{xres, yres, zres},
		E:        [5]float64{e0, e1, e2, e3, e4},
		Step:     step,
		Offset:   [3]float64{0, 0, zoffset},
		Polarity: [3]float64{1, 1, zpolarity},
		Model:    model,
		Gain:     gain,
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

		obs, err := client.Query([]string{fx, fy, fz, kd, ks}, t, dt)
		if err != nil {
			log.Fatalf("unable to query fdsn service: %v", err)
		}
		var full []Full
		for t, v := range obs {
			var f []float64
			for _, i := range v {
				f = append(f, float64(i)/volts)
			}
			if len(f) < 5 {
				continue
			}
			full = append(full, Full{
				Timestamp: t.Truncate(time.Second),
				Field:     [3]float64{f[0], f[1], f[2]},
				//TODO: bug?
				Driver: scale*f[4] - 273,
				Sensor: scale*f[3] - 273,
			})
		}

		for _, f := range flux.Split(full, truncate) {
			filename, err := f.Filename(base, path)
			if err != nil {
				log.Fatalf("unable to build file name %s: %v", path, err)
			}
			if err := f.WriteFile(string(filename)); err != nil {
				log.Fatalf("unable to store file %s: %v", string(filename), err)
			}
		}

		if !(interval > 0) {
			break
		}
	}
}
