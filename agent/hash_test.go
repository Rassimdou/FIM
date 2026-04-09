package agent

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestCalculateHash(t *testing.T) {
	// Setup a temporary directory for our test files
	// The testing package will automatically clean this up when the test finishes!
	tempDir := t.TempDir()

	//  Create a temporary file with known content
	content := []byte("FIM test content")
	testFile := filepath.Join(tempDir, "test.txt")

	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	// Manually calculate the expected hash so we know what the answer SHOULD be
	hasher := sha256.New()
	hasher.Write(content)
	expectedHash := hex.EncodeToString(hasher.Sum(nil))

	//  TEST CASE 1: Successful Hash
	t.Run("Valid Hash Calculation", func(t *testing.T) {
		// Call our function with a 1MB limit (plenty big enough for "FIM test content")
		actualHash, err := CalculateHash(testFile, 1*1024*1024)

		if err != nil {
			t.Errorf("CalculateHash returned an unexpected error: %v", err)
		}

		if actualHash != expectedHash {
			t.Errorf("Hash mismatch! Expected %s, got %s", expectedHash, actualHash)
		}
	})

	//TEST CASE 2: File Exceeds Size Limit
	t.Run("File Exceeds Max Size", func(t *testing.T) {
		// Call our function but lie and say the max size is only 5 bytes.
		// Since "FIM test content" is 16 bytes, it should fail!
		_, err := CalculateHash(testFile, 5)

		if err == nil {
			t.Error("Expected CalculateHash to return an error for exceeding max size, but it didn't")
		}
	})

	// TEST CASE 3: File Doesn't Exist
	t.Run("File Not Found", func(t *testing.T) {
		_, err := CalculateHash(filepath.Join(tempDir, "ghost.txt"), 1024)

		if err == nil {
			t.Error("Expected CalculateHash to return an error for a missing file, but it didn't")
		}
	})
}
