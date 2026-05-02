package readlimit

import (
	"io"
	"os"
)

func ReadAll(r io.Reader, maxBytes int64) ([]byte, bool, error) {
	limited := io.LimitReader(r, maxBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, false, err
	}
	if int64(len(data)) > maxBytes {
		return nil, true, nil
	}
	return data, false, nil
}

func ReadFile(path string, maxBytes int64) ([]byte, bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, false, err
	}
	defer file.Close()

	return ReadAll(file, maxBytes)
}
