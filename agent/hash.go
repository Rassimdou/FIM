package agent

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"time"
)

// CalculateHash calculates the SHA256 hash of a file, with retries for locked files and skips files larger than maxSizeBytes.
func CalculateHash(filePath string, maxSizeBytes int64) (string, error) {
	var file *os.File
	var err error

	// Retry mechanism: try up to 3 times with a 500ms delay
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		// Check file size first
		info, statErr := os.Stat(filePath)
		if statErr != nil {
			err = statErr
			time.Sleep(500 * time.Millisecond)
			continue
		}

		if info.Size() > maxSizeBytes {
			return "", fmt.Errorf("file too large: %d bytes (max %d)", info.Size(), maxSizeBytes)
		}

		file, err = os.Open(filePath)
		if err == nil {
			break // Successfully opened
		}

		// If there's an error opening (e.g., locked), wait and retry
		time.Sleep(500 * time.Millisecond)
	}

	if err != nil {
		return "", fmt.Errorf("failed to open file after %d retries: %w", maxRetries, err)
	}
	defer file.Close()

	//create a new sha256 hash
	hasher := sha256.New()

	_, err = io.Copy(hasher, file)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
