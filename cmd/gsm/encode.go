package main

import (
	"fmt"
	"strings"
)

func (g Gsm) Encode() ([]byte, error) {

	var lines []string
	for _, r := range g.Readings {
		lines = append(lines, strings.Join([]string{
			g.Prefix,
			r.Timestamp.Format(gsmTimeFormat),
			fmt.Sprintf("%.1f", r.Field),
			fmt.Sprintf("%.1f", float64(r.Quality)),
		}, ", "))
	}

	return []byte(strings.Join(lines, "\n") + "\n"), nil

}
