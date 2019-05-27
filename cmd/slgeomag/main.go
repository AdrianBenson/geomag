package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/GeoNet/kit/mseed"
	"github.com/GeoNet/kit/slink"

	"github.com/ozym/geomag/internal/gm"
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
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "  benmore    -- process benmore observations\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Use: \"%s <command> --help\" for more information about a specific command\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\n")
	}

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

	var offset time.Duration
	flag.DurationVar(&offset, "offset", 0, "offset from length")

	var publish time.Duration
	flag.DurationVar(&publish, "publish", time.Minute, "interval to publish")

	var truncate time.Duration
	flag.DurationVar(&truncate, "truncate", time.Hour, "interval to store files")

	var base string
	flag.StringVar(&base, "base", ".", "base output directory")

	var path string
	flag.StringVar(&path, "path", "{{year}}/{{year}}.{{yearday}}/{{year}}.{{yearday}}.{{hour}}{{minute}}.{{second}}.{{toupper .Label}}.csv", "file name template")

	var label string
	flag.StringVar(&label, "label", "", "provide a file name label")

	var dp int
	flag.IntVar(&dp, "dp", 0, "number of decimal places for raw data")

	flag.Parse()

	args := flag.Args()
	if !(len(args) > 0) {
		flag.Usage()
		fmt.Fprintf(os.Stderr, "Missing command or seedlink server\n")
		os.Exit(1)
	}

	server := args[len(args)-1]

	if err := os.MkdirAll(filepath.Dir(base), 0755); err != nil {
		log.Fatalf("unable to create base parent directory %s: %v", base, err)
	}

	// general miniseed amplitude block handler
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
				log.Printf("skipping block, invalid sample rate: %g", sps)
				continue
			}

			st, dt := msr.Starttime(), time.Duration(float64(time.Second)/sps)

			samples, err := msr.DataSamples()
			if err != nil {
				log.Printf("skipping block, unable to decode samples: %v", err)
				continue
			}

			raw := gm.NewRaw(srcname, dp)
			for i, s := range samples {
				raw.Add(gm.NewReading(st.Add(time.Duration(i)*dt), srcname, float64(s)))
			}

			log.Println("packet", srcname, st, len(samples))
			if err := raw.Store(base, path, truncate); err != nil {
				log.Fatalf("unable to store observations: %v", err)
			}
		}
	}()

	slconn := slink.NewSLCD()
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
		if err := os.MkdirAll(filepath.Dir(statefile), 0755); err != nil {
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
		if rc != slink.SLPACKET {
			break
		}
		if p.PacketType() != slink.SLDATA {
			continue
		}

		handler <- p.GetMSRecord()

		if statefile != "" {
			if t := time.Now(); t.Sub(last) > state {
				slconn.SaveState(statefile)
				last = t
			}
		}
	}
}