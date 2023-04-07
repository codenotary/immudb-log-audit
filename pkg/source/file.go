package source

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/nxadm/tail"
	log "github.com/sirupsen/logrus"
)

type fileWatch struct {
	fi os.FileInfo
	t  *tail.Tail
}

type fileTail struct {
	pattern      string
	follow       bool
	fileRegistry map[string]fileWatch
	lC           chan string
	wg           sync.WaitGroup
}

func NewFileTail(pattern string, follow bool) (*fileTail, error) {
	ft := fileTail{
		pattern:      pattern,
		follow:       follow,
		fileRegistry: map[string]fileWatch{},
		lC:           make(chan string),
	}

	ft.watchFiles()
	return &ft, nil
}

func (ft *fileTail) ReadLine() (string, error) {
	l, ok := <-ft.lC
	if !ok {
		return "", io.EOF
	}

	return l, nil
}

func (ft *fileTail) watchFiles() {
	go func() {
		for {
			newFiles, err := ft.listFiles()
			if err != nil {
				log.WithError(err).Error("could not list new files")
			} else {
				for _, t := range newFiles {
					ft.wg.Add(1)
					t := t
					go func() {
						log.WithField("file", t.Filename).Debug("Starting new file tailer")
						defer ft.wg.Done()
						for {
							l, ok := <-t.Lines
							if l != nil && l.Err != nil {
								log.WithError(err).WithField("file", t.Filename).Error("could not read lines, closing")
								return
							}

							if !ok {
								log.WithField("file", t.Filename).Info("File closed")
								return
							}

							ft.lC <- l.Text
						}
					}()
				}
			}

			if !ft.follow {
				ft.wg.Wait()
				close(ft.lC)
				break
			}

			time.Sleep(30 * time.Second)
		}
	}()
}

func (t *fileTail) listFiles() ([]*tail.Tail, error) {
	matches, err := filepath.Glob(t.pattern)
	if err != nil {
		return nil, fmt.Errorf("could not glob for pattern, %w", err)
	}

	newFileRegistry := map[string]fileWatch{}
	newFiles := []*tail.Tail{}

	log.WithField("files", matches).Debug("Matching files")
nextFile:
	for _, m := range matches {
		fi, err := os.Stat(m)
		if err != nil {
			log.WithError(err).WithField("file", m).Warn("Could not get file info, skipping")
			continue
		}

		// file already monitored, move to new registry
		for k, v := range t.fileRegistry {
			if os.SameFile(v.fi, fi) {
				log.WithField("file", k).Debug("Same file detected")
				newFileRegistry[m] = fileWatch{fi: fi, t: v.t}
				delete(t.fileRegistry, k)
				continue nextFile
			}
		}

		// it is new file to monitor
		t, err := tail.TailFile(m, tail.Config{Follow: t.follow, Logger: log.WithField("TAIL", m)})
		if err != nil {
			log.WithError(err).WithField("file", m).Warn("Could not start new tail, skipping")
			continue
		}

		newFileRegistry[m] = fileWatch{fi: fi, t: t}
		newFiles = append(newFiles, t)
		log.WithField("file", m).Debug("Monitoring new file")
	}

	for _, v := range t.fileRegistry {
		log.WithField("file", v.t.Filename).Info("Closing tail for removed files")
		v.t.Stop()
	}

	t.fileRegistry = newFileRegistry
	return newFiles, nil
}
