package main

import (
	"bytes"
	"fmt"
	"strings"
)

// Encode builds a byte slice containing the standard file header and reading lines.
func (f Fluxgate) Encode() ([]byte, error) {

	var lines []string

	var buf bytes.Buffer
	if err := f.Format(&buf); err != nil {
		return nil, err
	}

	for _, r := range f.Readings {
		var parts []string

		parts = append(parts, r.Timestamp.Format("15 04 05"))

		parts = append(parts, fmt.Sprintf("%7.2f", r.Driver))
		parts = append(parts, fmt.Sprintf("%6.2f", r.Sensor))

		parts = append(parts, fmt.Sprintf("%7.4f", r.Field[0]))
		parts = append(parts, fmt.Sprintf("%9.4f", r.Field[1]))
		parts = append(parts, fmt.Sprintf("%8.4f", r.Field[2]))

		calc := f.toCalc(r.Field)

		parts = append(parts, fmt.Sprintf("%10.3f", calc[0]))
		parts = append(parts, fmt.Sprintf("%9.3f", calc[1]))
		parts = append(parts, fmt.Sprintf("%10.3f", calc[2]))

		parts = append(parts, fmt.Sprintf("%10.3f", f.toF(calc)))
		parts = append(parts, fmt.Sprintf("%7.2f", f.toI(calc)))
		parts = append(parts, fmt.Sprintf("%8.4f", r.Benmore))

		lines = append(lines, strings.Join(parts, " "))
	}

	return append(buf.Bytes(), []byte(strings.Join(lines, "\n")+"\n")...), nil
}
