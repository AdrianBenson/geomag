package raw

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
)

const tmpPrefix = ".xxxx"

func readFile(path string) ([]byte, error) {

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func writeFile(path string, data []byte) error {

	dirmode := getDirMode(filepath.Dir(path))
	if err := os.MkdirAll(filepath.Dir(path), dirmode); err != nil {
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

	var disk []byte
	if _, err := os.Stat(path); err == nil {
		if disk, err = ioutil.ReadFile(path); err != nil {
			return err
		}
	}

	if bytes.Equal(data, disk) {
		return nil
	}

	if err := os.Rename(tmp.Name(), path); err != nil {
		return err
	}

	fmode := getFileMode(path)
	if err := os.Chmod(path, fmode); err != nil {
		return err
	}

	return nil
}

// True if setgid sticky bit is set
func isModeSetgid(fi os.FileInfo) bool {
	return fi != nil && fi.IsDir() && fi.Mode()&os.ModeSetgid != 0
}

// True if is dir and setgid has been set
func getFileMode(filename string) os.FileMode {
	fi, _ := os.Stat(filepath.Dir(filename))
	fmode := os.FileMode(0644)
	if isModeSetgid(fi) {
		fmode = os.FileMode(0664)
	}

	return fmode
}

// Step back through tree until find dir that exists, test setgid is set
// and set dir creation mode accordingly.
func getDirMode(path string) os.FileMode {
	dirmode := os.FileMode(0755)
	for {
		fi, err := os.Stat(path)
		if os.IsExist(err) {
			if isModeSetgid(fi) {
				dirmode = os.FileMode(0775)
			}
			break
		}
		if path == "." || path == string(os.PathSeparator) {
			break
		} else {
			path = filepath.Dir(path)
		}
	}

	return dirmode
}
