package agent

import (
	"context"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/Rassimdou/FIM/proto"
	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	fsw      *fsnotify.Watcher
	cfg      *Config
	debounce map[string]time.Time
	hostname string
	agentOS  string
	onEvent  func(*proto.FileEvent)
}

func NewWatcher(cfg *Config, onEvent func(*proto.FileEvent)) (*Watcher, error) {

	// create fsnotify watcher
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	//get hostname and OS
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	//build Watcher struct
	w := &Watcher{
		fsw:      fsw,
		cfg:      cfg,
		debounce: make(map[string]time.Time),
		hostname: hostname,
		agentOS:  runtime.GOOS,
		onEvent:  onEvent,
	}

	//add watch paths
	for _, root := range cfg.Watch.Paths {
		if cfg.Watch.Recursive {

			//walk every entry under root
			err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return nil
				}

				if d.IsDir() {
					if addErr := fsw.Add(path); addErr != nil {
						return addErr
					}
				}
				return nil
			})
			if err != nil {
				return nil, err
			}

		} else {
			err = fsw.Add(root)
			if err != nil {
				return nil, err
			}
		}

	}

	return w, nil
}

func (w *Watcher) Start(ctx context.Context) error {
	defer w.fsw.Close()

	for {
		select {
		case event, ok := <-w.fsw.Events:
			if !ok {
				return nil
			}

			if time.Since(w.debounce[event.Name]) < 500*time.Millisecond {
				continue
			}

			w.debounce[event.Name] = time.Now()

			// Keep recursive watching alive for directories created after startup.
			if w.cfg.Watch.Recursive && event.Op.Has(fsnotify.Create) {
				if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
					if err := w.fsw.Add(event.Name); err != nil {
						log.Printf("watcher add dir error: %v", err)
					}
				}
			}

			eventType := opToEventType(event.Op)
			if eventType == proto.EventType_UNKNOWN {
				continue
			}

			fileEvent := &proto.FileEvent{
				Hostname:  w.hostname,
				Os:        w.agentOS,
				FilePath:  event.Name,
				EventType: eventType,
				Timestamp: time.Now().UnixNano(),
			}
			w.onEvent(fileEvent)

		case err, ok := <-w.fsw.Errors:
			if !ok {
				return nil
			}
			log.Printf("watcher error: %v", err)

		case <-ctx.Done():
			return nil
		}
	}
}

func opToEventType(op fsnotify.Op) proto.EventType {
	switch {
	case op.Has(fsnotify.Create):
		return proto.EventType_FILE_CREATE
	case op.Has(fsnotify.Write):
		return proto.EventType_FILE_MODIFY
	case op.Has(fsnotify.Remove):
		return proto.EventType_FILE_DELETE
	case op.Has(fsnotify.Chmod):
		return proto.EventType_CHANGE_PERMISSION
	default:
		return proto.EventType_UNKNOWN
	}
}
