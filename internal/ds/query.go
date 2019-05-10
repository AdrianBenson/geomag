package ds

import (
	"time"
)

type Dataselect struct {
	Service string
	Timeout time.Duration
}

func NewDataselect(service string, timeout time.Duration) Dataselect {
	return Dataselect{
		Service: service,
		Timeout: timeout,
	}
}

func (d Dataselect) Query(srcnames []string, at time.Time, length time.Duration) (map[time.Time][]int32, error) {

	client := NewClient(d.Timeout)

	obs := make(map[time.Time][]int32)
	for _, s := range srcnames {
		r, err := NewSource(s).Request(d.Service, at, length)
		if err != nil {
			return nil, err
		}
		if err := client.Sample(r, func(s string, t time.Time, v int32) error {
			obs[t] = append(obs[t], v)
			return nil
		}); err != nil {
			return nil, err
		}
	}

	return obs, nil
}
