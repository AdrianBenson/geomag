package geomag

import (
	"fmt"
	"io"
	"text/template"
	"time"
)

const txtHeader = `e2={{printf "%.3f" (index .E 2)}}
zpolarity={{index .Polarity z}}
xcoil={{index .Coil x}}
date={{stamp "2006/01/02" .Timestamp}} ({{yearday .Timestamp}}) {{stamp "15:04:05" .Timestamp}} GMT
model={{.Model}}
sensor={{.Sensor}}
e0={{printf "%.3f" (index .E 0)}}
xbias={{printf "0x%02x" (index .Bias x)}}
zbias={{printf "0x%02x" (index .Bias z)}}
zcoil={{index .Coil z}}
step={{.Step}}
e4={{printf "%.3f" (index .E 4)}}
e3={{printf "%.3f" (index .E 3)}}
code={{.Label}}
e1={{printf "%.3f" (index .E 1)}}
epoch={{.Timestamp.Unix}}
ycoil={{index .Coil y}}
ybias={{printf "0x%02x" (index .Bias y)}}
zoffset={{index .Offset z}}
yres={{index .Res y}}
zres={{index .Res z}}
xres={{index .Res x}}
scale={{.Gain}}
driver={{.Driver}}
HH MM SS  S Temp D Temp   X Raw     Y Raw    Z Raw     X Calc    Y Calc     Z Calc     F Calc  I Calc  Benmore
`

func (f Fluxgate) Format(wr io.Writer) error {
	tmpl, err := template.New("txt").Funcs(
		template.FuncMap{
			"stamp": func(f string, t time.Time) string {
				return t.Format(f)
			},
			"yearday": func(t time.Time) string {
				return fmt.Sprintf("%04d.%03d", t.Year(), t.YearDay())
			},
			"x": func() int {
				return 0
			},
			"y": func() int {
				return 1
			},
			"z": func() int {
				return 2
			},
		}).Parse(txtHeader)
	if err != nil {
		return err
	}

	if err := tmpl.Execute(wr, f); err != nil {
		return err
	}

	return nil

}
