package gm

import (
//	"sync"
//	"time"
)

/**
type Cacher interface {
	Lock()
	Unlock()

	Add(string, Valuer)
	Store(
	//
}
**/

/**
type Cache struct {
	mu sync.RWMutex

	base      string
	path      string
	truncate  time.Duration
	precision int

	raw map[string]Formatter
}

func (c *Cache) Add(srcname string, r Reading) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.raw[srcname]; !ok {
		c.raw[srcname] = gm.NewRaw(srcname, c.precision)
	}

	if x, ok := c.raw[srcname]; ok {
		x.Add(r)
	}
}

func (c *Cache) Store() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for k, v := range c.raw {
		if err := v.Store(c.base, c.path, c.truncate); err != nil {
			return err
		}
		delete(c.raw, k)
	}

	return nil
}
**/
