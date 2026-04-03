package agent 

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"io"

)


func CalculateHash(filePath string) (string, error) {
	file ,err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	//create a new sha256 hash
	hasher := sha256.New()

	_, err = io.Copy(hasher, file)
	if  err != nil {
		return "", err
	}
	
	return hex.EncodeToString(hasher.Sum(nil)), nil
	


}