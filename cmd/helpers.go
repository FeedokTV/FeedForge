package cmd

import (
	"fmt"
	"io"
	"os"
)

func openInput(path string) (io.ReadCloser, error) {
	if path == "" {
		return os.Stdin, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open input %q: %w", path, err)
	}
	return file, nil
}

func openOutput(path string) (io.WriteCloser, error) {
	if path == "" {
		return os.Stdout, nil
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("open output %q: %w", path, err)
	}
	return f, nil
}
