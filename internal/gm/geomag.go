package gm

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

/**
type Reading struct {
	Timestamp time.Time
	Label     string
	Field     float64
}

func NewReading(t time.Time, l string, v float64) Reading {
	return Reading{
		Timestamp: t,
		Label:     l,
		Field:     v,
	}
}
**/

type Valuer interface {
	At() time.Time
	Tag() string
	Values() []float64
}

type Formatter interface {
	At() time.Time
	Tag() string
	Encode() ([]byte, error)
	Decode(data []byte) error
	Merge(data []byte) error
	Split(time.Duration) []Formatter
}

func Store(base, path string, truncate time.Duration, format Formatter) error {
	for _, f := range format.Split(truncate) {
		filename, err := Filename(base, path, f)
		if err != nil {
			return err
		}
		if err := WriteFile(string(filename), f); err != nil {
			return err
		}
	}

	return nil
}

func ReadFile(path string, format Formatter) error {

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	if err := format.Decode(data); err != nil {
		return err
	}

	return nil
}

func WriteFile(path string, format Formatter) error {

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	if _, err := os.Stat(path); err == nil {
		data, err := readFile(path)
		if err != nil {
			return err
		}
		if err := format.Merge(data); err != nil {
			return err
		}
	}

	data, err := format.Encode()
	if err != nil {
		return err
	}

	if err := writeFile(path, data); err != nil {
		return err
	}

	return nil
}
