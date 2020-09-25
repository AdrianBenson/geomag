package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/nightlyone/lockfile"

	"github.com/ozym/geomag/internal/mseed"
	"github.com/ozym/geomag/internal/raw"
	"github.com/ozym/geomag/internal/slink"
)

const timeFormat = "2006,01,02,15,04,05"

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Build geomag data files using seedlink\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "  %s [options] <command> [options] <server>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "General Options:\n")
		fmt.Fprintf(os.Stderr, "\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n")
	}

	var verbose bool
	flag.BoolVar(&verbose, "verbose", false, "make noise")

	var lock string
	flag.StringVar(&lock, "lockfile", "", "provide a process lock file")

	// seedlink options
	var netdly time.Duration
	flag.DurationVar(&netdly, "netdly", 0, "provide network delay")

	var netto time.Duration
	flag.DurationVar(&netto, "netto", 300*time.Second, "provide network timeout")

	var keepalive time.Duration
	flag.DurationVar(&keepalive, "keepalive", 0, "provide keep-alive")

	var streams string
	flag.StringVar(&streams, "streams", "*_*", "provide default seedlink streams query")

	var selectors string
	flag.StringVar(&selectors, "selectors", "???", "provide default seedlink selection query")

	var startup time.Duration
	flag.DurationVar(&startup, "startup", time.Hour, "startup offset if no statefile")

	var statefile string
	flag.StringVar(&statefile, "statefile", "", "provide a running state file")

	var state time.Duration
	flag.DurationVar(&state, "state", 5*time.Minute, "how often to save state")

	var truncate time.Duration
	flag.DurationVar(&truncate, "truncate", time.Hour, "interval to store files")

	var base string
	flag.StringVar(&base, "base", ".", "base output directory")

	var path string
	flag.StringVar(&path, "path", "{{year}}/{{year}}.{{yearday}}/{{year}}.{{yearday}}.{{hour}}{{minute}}.{{second}}.{{toupper .Label}}.csv", "file name template")

	var dp int
	flag.IntVar(&dp, "dp", 0, "number of decimal places for raw data")

	var gain float64
	flag.Float64Var(&gain, "gain", 1.0, "apply a gain to the raw data")

	flag.Parse()

	args := flag.Args()
	if !(len(args) > 0) {
		flag.Usage()
		fmt.Fprintf(os.Stderr, "Missing command or seedlink server\n")
		os.Exit(1)
	}

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

	server := args[len(args)-1]

	fi, err := os.Stat(base)
	switch {
	case err != nil:
		log.Fatalf("cannot write to base directory: %v", err)
	case !fi.IsDir():
		log.Fatalf("cannot write to base directory: %s: not a directory", base)
	}

	handler := make(chan []byte, 20000)
	go func() {
		msr := mseed.NewMSRecord()
		defer mseed.FreeMSRecord(msr)

		for b := range handler {
			if err := msr.Unpack(b, 512, 1, 0); err != nil {
				log.Printf("skipping block, unable to unpack block: %v", err)
				continue
			}

			srcname := msr.SrcName(0)

			sps := float64(msr.Samprate())
			if !(sps > 0) {
				log.Printf("skipping block, invalid sample rate %s: %g", srcname, sps)
				continue
			}

			st, dt := msr.Starttime(), time.Duration(float64(time.Second)/sps)

			samples, err := msr.DataSamples()
			if err != nil {
				log.Printf("skipping block, unable to decode samples %s: %v", srcname, err)
				continue
			}

			geomag := raw.NewRaw(srcname, dp)
			for i, s := range samples {
				geomag.Add(raw.NewReading(st.Add(time.Duration(i)*dt), srcname, gain*float64(s)))
			}

			if verbose {
				log.Printf("handling packet %s: %s (%d)", srcname, st, len(samples))
			}
			if err := geomag.Store(base, path, truncate); err != nil {
				log.Fatalf("unable to store observations: %v", err)
			}
		}
	}()

	var slconn *slink.SLCD
	slconn = slink.NewSLCD()
	defer slink.FreeSLCD(slconn)

	// seedlink settings
	slconn.SetNetDly(int(netdly / time.Second))
	slconn.SetNetTo(int(netto / time.Second))
	slconn.SetKeepAlive(int(keepalive / time.Second))
	slconn.ParseStreamList(streams, selectors)

	// conection
	slconn.SetSLAddr(server)
	defer slconn.Disconnect()

	switch {
	case statefile != "":
		if err := os.MkdirAll(filepath.Dir(statefile), 0775); err != nil {
			log.Fatalf("unable to create statefile parent directory %s: %v", statefile, err)
		}
		switch _, err := os.Stat(statefile); err {
		case nil:
			if n := slconn.RecoverState(statefile); n != 0 {
				slconn.SetBeginTime(time.Now().UTC().Add(-startup).Format(timeFormat))
			}
		default:
			slconn.SetBeginTime(time.Now().UTC().Add(-startup).Format(timeFormat))
		}
	default:
		slconn.SetBeginTime(time.Now().UTC().Add(-startup).Format(timeFormat))
	}

	var last time.Time
	for {
		p, rc := slconn.Collect()
		if rc == slink.SLTERMINATE {
			log.Printf("SLTERMINATE signal received.")
			break
		} else if rc != slink.SLPACKET {
			log.Printf("Collect return value not SLPACKET or SLTERMINATE: %d", rc)
			break
		}
		if p.PacketType() != slink.SLDATA {
			continue
		}

		handler <- p.GetMSRecord()

		if statefile != "" {
			if verbose {
				log.Printf("saving state: %s", statefile)
			}
			if t := time.Now(); t.Sub(last) > state {
				slconn.SaveState(statefile)
				last = t
			}
		}
	}
}
