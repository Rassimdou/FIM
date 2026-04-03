package agent

import (
	"context"
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/Rassimdou/FIM/proto"
)

func ScanExistingFiles(ctx context.Context, cfg *Config, onEvent func(*proto.FileEvent)) error {

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	agentOS := runtime.GOOS

	for _, root := range cfg.Watch.Paths {
		log.Printf("Starting baseline scan on: %s", root)

		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if ctx.Err() != nil {
				return errors.New("scan cancelled")

			}

			if err != nil {
				log.Printf("scan error accessing path %q: %v\n", path, err)
				return nil
			}
			//if its a folder check if its excluded
			if d.IsDir() {
				if isExcluded(path, cfg.Watch.Exclude) {
					return filepath.SkipDir
				}
				//skip we only care about files here
				return nil
			}

			//check if file is excluded
			if isExcluded(path, cfg.Watch.Exclude) {
				return nil
			}

			//calculate hash of the file
			hashValue, err := CalculateHash(path)
			if err != nil {
				log.Printf("failed to calculate hash for %s: %v", path, err)
				return nil //skip this file but keep scanning
			}

			//create the scan event
			fileEvent := &proto.FileEvent{
				Hostname:  hostname,
				Os:        agentOS,
				FilePath:  path,
				EventType: proto.EventType_FILE_SCAN,
				NewHash:   hashValue,
				Timestamp: time.Now().UnixNano(),
			}

			//send event to the queue
			onEvent(fileEvent)

			return nil
		})

		if err != nil {
			log.Printf("error during baseline scan of %s: %v", root, err)
		}

	}
	log.Println("Baseline scan completed.")
	return nil

}
