package utils

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
)

func ReadFileTail(filePath string, maxLen int) (data []byte, err error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		err = f.Close()
		if err != nil {
			err = fmt.Errorf("failed to close file: %w", err)
			data = nil
		}
	}()

	var dataReader io.Reader = f

	if maxLen > 0 {
		stat, err := f.Stat()
		if err != nil {
			return nil, fmt.Errorf("failed to stat file: %w", err)
		}

		if stat.Size() > int64(maxLen) {
			_, err = f.Seek(-int64(maxLen), io.SeekEnd)
			if err != nil {
				return nil, fmt.Errorf("failed to seek file: %w", err)
			}
		}

		dataReader = io.LimitReader(f, int64(maxLen))
	}

	data, err = io.ReadAll(dataReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}

// filters values from a slice
func Filter[T any](s []T, f func(T) bool) []T {
	var r []T
	for _, v := range s {
		if f(v) {
			r = append(r, v)
		}
	}
	return r
}

func ParseCommandLine(s string) ([]string, error) {
	args := make([]string, 0)
	var currentArg strings.Builder
	inQuotes := false
	escaped := false

	for _, r := range s {
		switch {
		case escaped:
			currentArg.WriteRune(r)
			escaped = false
		case r == '\\':
			escaped = true
		case r == '"':
			inQuotes = !inQuotes
		case unicode.IsSpace(r) && !inQuotes:
			if currentArg.Len() > 0 {
				args = append(args, currentArg.String())
				currentArg.Reset()
			}
		default:
			currentArg.WriteRune(r)
		}
	}

	if escaped || inQuotes {
		return nil, fmt.Errorf("unexpected end of string")
	}

	if currentArg.Len() > 0 {
		args = append(args, currentArg.String())
	}

	return args, nil
}
