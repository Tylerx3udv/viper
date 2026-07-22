package viper

import (
	"log"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

// WatchConfig starts watching the config file for changes.
func WatchConfig() { supportedViper.WatchConfig() }

// WatchConfig starts watching the config file for changes.
func (v *Viper) WatchConfig() {
	initWatcher()

	configFile := filepath.Clean(v.configFile)
	configDir, _ := filepath.Split(configFile)
	realConfigFile, _ := filepath.EvalSymlinks(configFile)

	watcher.Add(configDir)

	go func() {
		var (
			delay        = 100 * time.Millisecond
			timer        *time.Timer
			timerChan    <-chan time.Time
			pendingEvent *fsnotify.Event
		)
		defer func() {
			if timer != nil {
				timer.Stop()
			}
		}()

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				// we only care about the config file with the absolute path
				currentConfigFile, _ := filepath.EvalSymlinks(v.configFile)
				if filepath.Clean(event.Name) == configFile || filepath.Clean(event.Name) == realConfigFile || filepath.Clean(event.Name) == currentConfigFile {
					if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
						pendingEvent = &event
						if timer != nil {
							timer.Stop()
						}
						timer = time.NewTimer(delay)
						timerChan = timer.C
					}
				}

			case <-timerChan:
				timerChan = nil
				timer = nil
				if pendingEvent != nil {
					event := *pendingEvent
					pendingEvent = nil
					err := v.ReadInConfig()
					if err != nil {
						log.Println("error:", err)
						continue
					}
					v.OnConfigChange(event)
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()
}
