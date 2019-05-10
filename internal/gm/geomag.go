package gm

import (
	"os"
	"path/filepath"
)

type Decoder interface {
	Decode([]byte) error
}

type Encoder interface {
	Encode() ([]byte, error)
	Merge([]byte) error
}

// ReadFile will read and decode a fluxgate file.
func ReadFile(path string, dc Decoder) error {

	data, err := readFile(path)
	if err != nil {
		return err
	}

	if err := dc.Decode(data); err != nil {
		return err
	}

	return nil
}

// WriteFile will encode and write a fluxgate file, existing values will be merged.
func WriteFile(path string, en Encoder) error {

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	if _, err := os.Stat(path); err == nil {
		data, err := readFile(path)
		if err != nil {
			return err
		}

		if err := en.Merge(data); err != nil {
			return err
		}
	}

	data, err := en.Encode()
	if err != nil {
		return err
	}

	if err := writeFile(path, data); err != nil {
		return err
	}

	return nil
}
