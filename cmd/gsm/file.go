package main

import (
	"github.com/ozym/geomag/internal/gm"
)

// ReadFile will read and decode a benmore file.
func (g *Gsm) ReadFile(path string) error {
	return gm.ReadFile(path, g)
}

// WriteFile will encode and write a benmore file, existing values will be merged.
func (g *Gsm) WriteFile(path string) error {
	return gm.WriteFile(path, g)
}
