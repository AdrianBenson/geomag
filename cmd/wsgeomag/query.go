package main

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const fdsnFormat = "2006-01-02T15:04:05.000000"
const fdsnQuery = "/fdsnws/dataselect/1/query?"

type Dataselect struct {
	Service string
	Timeout time.Duration
}

func NewDataselect(service string, timeout time.Duration) *Dataselect {
	return &Dataselect{
		Service: service,
		Timeout: timeout,
	}
}

func (d *Dataselect) Query(srcnames []string, at time.Time, length time.Duration) ([]byte, error) {

	client := &http.Client{
		Timeout: d.Timeout,
	}

	var data []byte
	for _, s := range srcnames {
		query, err := d.Request(s, at, length)
		if err != nil {
			return nil, err
		}

		request, err := http.NewRequest("GET", query, nil)
		if err != nil {
			return nil, err
		}

		resp, err := client.Do(request)
		if resp == nil || err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		data = append(data, body...)
	}

	return data, nil
}

func (d *Dataselect) Request(srcname string, endtime time.Time, length time.Duration) (string, error) {

	s := NewSource(srcname)

	values := url.Values{}
	values.Add("starttime", endtime.Add(-length).Format(fdsnFormat))
	values.Add("endtime", endtime.Format(fdsnFormat))
	values.Add("network", s.Network)
	values.Add("station", s.Station)
	values.Add("location", s.Location)
	values.Add("channel", s.Channel)

	req, err := url.Parse(strings.TrimRight(d.Service, "/") + fdsnQuery + values.Encode())
	if err != nil {
		return "", err
	}

	return req.String(), nil
}
