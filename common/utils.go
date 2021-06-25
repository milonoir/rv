package common

import (
	"fmt"
	"io/ioutil"
)

// LoadFile returns a byte slice containing the contents of the given file.
//
// Will return an error if the file contents have a length of 0.
func LoadFile(file string) ([]byte, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	if len(b) == 0 {
		return nil, fmt.Errorf("read %s: empty file", file)
	}

	return b, nil
}
