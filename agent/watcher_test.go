package agent

import (
	"testing"

	"github.com/Rassimdou/FIM/proto"
	"github.com/fsnotify/fsnotify"
)

func TestIsExcluded(t *testing.T) {
	excludePatterns := []string{".git", "*.tmp", "*.log"}

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"Normal File", "home/user/document.txt", false},
		{"Wildcard Log", "var/log/syslog.log", true},
		{"Wildcard Tmp", "tmp/cache.tmp", true},
		{"Exact Match Dir", "/opt/project/.git", true},
		{"Prefix Mismatch", "git_repo/file.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isExcluded(tt.path, excludePatterns)
			if result != tt.expected {
				t.Errorf("isExcluded(%q) = %v; want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestOpToEventType(t *testing.T) {
	tests := []struct {
		name     string
		op       fsnotify.Op
		expected proto.EventType
	}{
		{"Create", fsnotify.Create, proto.EventType_FILE_CREATE},
		{"Write", fsnotify.Write, proto.EventType_FILE_MODIFY},
		{"Remove", fsnotify.Remove, proto.EventType_FILE_DELETE},
		{"Rename", fsnotify.Rename, proto.EventType_UNKNOWN}, // We don't map Rename currently
		{"Chmod", fsnotify.Chmod, proto.EventType_CHANGE_PERMISSION},
		{"Combo (Create+Write)", fsnotify.Create | fsnotify.Write, proto.EventType_FILE_CREATE}, // Create is checked first
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := opToEventType(tt.op)
			if result != tt.expected {
				t.Errorf("opToEventType(%v) = %v; want %v", tt.op, result, tt.expected)
			}
		})
	}
}
