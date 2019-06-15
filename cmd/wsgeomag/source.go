package main

import (
	//	"net/url"
	"strings"
	//	"time"
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

/**
const fdsnFormat = "2006-01-02T15:04:05.000000"
const fdsnQuery = "/fdsnws/dataselect/1/query?"

func (s Source) Request(fdsn string, endtime time.Time, length time.Duration) (string, error) {

	values := url.Values{}

	values.Add("starttime", endtime.Add(-length).Format(fdsnFormat))
	values.Add("endtime", endtime.Format(fdsnFormat))
	values.Add("network", s.Network)
	values.Add("station", s.Station)
	values.Add("location", s.Location)
	values.Add("channel", s.Channel)

	req, err := url.Parse(strings.TrimRight(fdsn, "/") + fdsnQuery + values.Encode())
	if err != nil {
		return "", err
	}

	return req.String(), nil
}
**/
