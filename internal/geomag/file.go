package geomag

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
)

const tmpPrefix = ".xxxx"

func writeFile(path string, data []byte) error {

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	tmp, err := ioutil.TempFile(filepath.Dir(path), ".xxxx")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(data); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	var ondisk []byte
	if _, err := os.Stat(path); err == nil {
		ondisk, err = ioutil.ReadFile(path)
		if err != nil {
			return err
		}
	}

	if !bytes.Equal(data, ondisk) {
		if err := os.Rename(tmp.Name(), path); err != nil {
			return err
		}
		if err := os.Chmod(path, 0644); err != nil {
			return err
		}
	}

	return nil
}
