package agent

import (
	"io/fs"
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
					return err
				}

				// only add directories, fsnotify watches dirs not files
				if d.IsDir() {
					if addErr := fsw.Add(path); addErr != nil {
						return addErr
					}
				}
				return err
			})
		} else {
			err = fsw.Add(root)
			if err != nil {
				return nil, err
			}
		}

	}

	return w, nil
}
