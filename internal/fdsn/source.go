package fdsn

import (
	"net/url"
	"strings"
	"time"
)

type Source struct {
	Network  string
	Station  string
	Location string
	Channel  string
}

func (s Source) String() string {
	return strings.Join([]string{
		s.Network,
		s.Station,
		s.Location,
		s.Channel,
	}, "_")
}

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
