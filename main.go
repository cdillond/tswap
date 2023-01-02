package tswap

// Package tswap automatically updates an html/template.Template when files in the directory where the template definitions are stored are updated.
// This can be useful if you want to work on changes to a website's UI without recompiling your application every time minor updates to a template definition file are made.
// tswap is dependent on github.com/fsnotify/fsnotify, which works for most, but not all, commonly used OS's.

import (
	"errors"
	"fmt"
	"html/template"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// AutoUpdate takes a pointer to the template.Template that should be updated, the path to the directory where the template definition files are stored, and a pointer to a read-write mutex. The mutex should be used when accessing the template.Template from your web app to avoid race conditions resulting from updates made by AutoUpdate.
// Make sure errors can be received from the (buffered) chan error returned by AutoUpdate. Otherwise, AutoUpdate will pause if the channel is full.
// Calls to AutoUpdate must be followed by a blocking operation.
func AutoUpdate(t *template.Template, dir string, rwm *sync.RWMutex) chan error {
	errChan := make(chan error, 5)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		watcher.Close()
		errChan <- fmt.Errorf("tswap AutoUpdate error: %w", err)
		return errChan
	}

	if err = watcher.Add(dir); err != nil {
		watcher.Close()
		errChan <- fmt.Errorf("tswap AutoUpdate error: %w", err)
		return errChan
	}

	go func() {
		defer watcher.Close()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					errChan <- errors.New("AutoUpdater shutting down")
					return
				}
				fmt.Println(event)
				rwm.Lock()
				temp, err := template.ParseGlob(dir + `*`)
				if err != nil {
					errChan <- fmt.Errorf("tswap AutoUpdate error: %w", err)
					rwm.Unlock()
					continue
				}
				*t = *temp
				rwm.Unlock()
			case err, ok := <-watcher.Errors:
				if !ok {
					errChan <- errors.New("AutoUpdater shutting down")
					return
				}
				errChan <- fmt.Errorf("tswap AutoUpdate error: %w", err)
			}
		}
	}()
	return errChan
}
