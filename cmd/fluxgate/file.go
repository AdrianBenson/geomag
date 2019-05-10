package main

import (
	"github.com/ozym/geomag/internal/gm"
)

// ReadFile will read and decode a fluxgate file.
func (f *Fluxgate) ReadFile(path string) error {
	return gm.ReadFile(path, f)
}

// WriteFile will encode and write a fluxgate file, existing values will be merged.
func (f *Fluxgate) WriteFile(path string) error {
	return gm.WriteFile(path, f)
}
