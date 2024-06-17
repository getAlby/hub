package utils

import (
	"fmt"
	"io"
	"os"
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
