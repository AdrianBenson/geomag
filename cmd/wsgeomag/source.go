package main

import (
	"strings"
)

type Source struct {
	Network  string
	Station  string
	Location string
	Channel  string
}

func NewSource(srcname string) Source {

	srcname = strings.TrimSpace(srcname)
	srcname = strings.Replace(srcname, "_", "_ ", -1)

	parts := strings.Split(srcname, "_")

	s := Source{"*", "*", "*", "*"}

	switch {
	case len(parts) > 3:
		if v := strings.TrimSpace(parts[0]); v != "" {
			s.Network = v
		}
		if v := strings.TrimSpace(parts[1]); v != "" {
			s.Station = v
		}
		if v := strings.TrimSpace(parts[2]); v != "" {
			s.Location = v
		}
		if v := strings.TrimSpace(parts[3]); v != "" {
			s.Channel = v
		}
	case len(parts) > 2:
		if v := strings.TrimSpace(parts[0]); v != "" {
			s.Network = v
		}
		if v := strings.TrimSpace(parts[1]); v != "" {
			s.Station = v
		}
		if v := strings.TrimSpace(parts[2]); v != "" {
			s.Channel = v
		}
	case len(parts) > 1:
		if v := strings.TrimSpace(parts[0]); v != "" {
			s.Network = v
		}
		if v := strings.TrimSpace(parts[1]); v != "" {
			s.Station = v
		}
	case len(parts) > 0:
		if v := strings.TrimSpace(parts[0]); v != "" {
			s.Network = v
		}
	}

	return s
}

func (s Source) String() string {
	return strings.Join([]string{
		s.Network,
		s.Station,
		s.Location,
		s.Channel,
	}, "_")
}
