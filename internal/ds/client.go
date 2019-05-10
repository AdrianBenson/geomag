package ds

import (
	"io/ioutil"
	"net/http"
	"time"

	"github.com/GeoNet/kit/mseed"
)

type Client struct {
	*http.Client
}

func NewClient(timeout time.Duration) *Client {
	return &Client{
		Client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) Raw(query string, fn func([]byte) error) error {

	request, err := http.NewRequest("GET", query, nil)
	if err != nil {
		return err
	}

	resp, err := c.Do(request)
	if resp == nil || err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := fn(body); err != nil {
		return err
	}

	return nil
}

func (c *Client) Blocks(query string, fn func([]byte) error) error {
	return c.Raw(query, func(data []byte) error {
		for i := 0; i < len(data); i += 512 {
			if err := fn(data[i : i+512]); err != nil {
				return err
			}
		}
		return nil
	})
}

func (c *Client) Samples(query string, fn func(string, time.Time, time.Duration, []int32) error) error {
	return c.Blocks(query, func(blk []byte) error {

		msr := mseed.NewMSRecord()
		defer mseed.FreeMSRecord(msr)

		if err := msr.Unpack(blk, 512, 1, 0); err != nil {
			return err
		}

		sps := msr.Samprate()
		if !(sps > 0.0) {
			return nil
		}

		dt := time.Duration(float64(time.Second) / float64(sps))

		samples, err := msr.DataSamples()
		if err != nil {
			return err
		}

		if err := fn(msr.SrcName(0), msr.Starttime(), dt, samples); err != nil {
			return err
		}

		return nil
	})
}

func (c *Client) Sample(query string, fn func(string, time.Time, int32) error) error {
	return c.Samples(query, func(s string, t time.Time, d time.Duration, v []int32) error {
		for i, j := range v {
			if err := fn(s, t.Add(time.Duration(i)*d), j); err != nil {
				return err
			}
		}
		return nil
	})
}
