package main

import (
	"fmt"
	"strings"
)

func (b Benmore) Encode() ([]byte, error) {

	var lines []string
	for _, r := range b.Readings {
		lines = append(lines, strings.Join([]string{
			b.Prefix,
			r.Timestamp.Format(benmoreFormat),
			fmt.Sprintf("%.8f", r.Field),
		}, ", "))
	}

	return []byte(strings.Join(lines, "\n") + "\n"), nil
}
