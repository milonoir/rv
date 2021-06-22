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
		return nil, fmt.Errorf("read file %q: %w", file, err)
	}

	if len(b) == 0 {
		return nil, fmt.Errorf("empty file: %q", file)
	}

	return b, nil
}
