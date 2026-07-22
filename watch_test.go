package viper

import (
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

func TestWatchConfigDebounce(t *testing.T) {
	// Create a temporary config file
	f, err := os.CreateTemp("", "viper_test_*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())

	_, err = f.Write([]byte("foo: bar\n"))
	if err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	f.Close()

	v := New()
	v.SetConfigFile(f.Name())
	err = v.ReadInConfig()
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	var changeCount int32
	v.OnConfigChange(func(e fsnotify.Event) {
		atomic.AddInt32(&changeCount, 1)
	})

	v.WatchConfig()
	// Wait a bit for watcher to start
	time.Sleep(100 * time.Millisecond)

	// Simulate in-place write (truncate + write) multiple times rapidly
	for i := 0; i < 3; i++ {
		f2, err := os.OpenFile(f.Name(), os.O_RDWR|os.O_TRUNC, 0666)
		if err != nil {
		t.Fatalf("failed to open file: %v", err)
		}
		_, err = f2.Write([]byte("foo: baz\n"))
		if err != nil {
			t.Fatalf("failed to write to file: %v", err)
		}
		f2.Close()
		time.Sleep(10 * time.Millisecond)
	}

	// Wait for events to settle (debounce is 100ms, so 500ms is plenty)
	time.Sleep(500 * time.Millisecond)

	count := atomic.LoadInt32(&changeCount)
	if count != 1 {
		t.Errorf("expected OnConfigChange to be called exactly once, got %d", count)
	}
}
