package main

import (
	"github.com/ozym/geomag/internal/gm"
)

// ReadFile will read and decode a benmore file.
func (b *Benmore) ReadFile(path string) error {
	return gm.ReadFile(path, b)
}

// WriteFile will encode and write a benmore file, existing values will be merged.
func (b *Benmore) WriteFile(path string) error {
	return gm.WriteFile(path, b)
}
