/*
Copyright 2023 Codenotary Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package source

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/nxadm/tail"
	log "github.com/sirupsen/logrus"
)

type fileMonitor struct {
	Prefix       []byte `json:"prefix"`
	PrefixLength int    `json:"prefix_length"`
	Offset       int64  `json:"offset"`
}

type fileWatch struct {
	fi *os.FileInfo
	t  *tail.Tail
	Fm *fileMonitor `json:"file_monitor"`
}

type fileTail struct {
	pattern        string
	follow         bool
	registryDB     bool
	registryDBFile string
	registryMutex  sync.RWMutex
	registry       map[string]fileWatch
	lC             chan string
	wg             sync.WaitGroup
	ctx            context.Context
}

func NewFileTail(ctx context.Context, pattern string, follow bool, registryDB bool, registryFolder string) (*fileTail, error) {

	registry := map[string]fileWatch{}

	registryDBFile := "registry-file.txt"

	if registryDB {
		if registryFolder != "" {
			fi, err := os.Stat(registryFolder)
			if err != nil {
				return nil, fmt.Errorf("could not stat registry directory: %w", err)
			}

			if !fi.IsDir() {
				return nil, fmt.Errorf("registry folder is not a directory")
			}

			registryDBFile = path.Join(registryFolder, registryDBFile)
		}

		frBytes, err := os.ReadFile(registryDBFile)
		if err != nil {
			log.WithError(err).WithField("path", registryDBFile).Info("registry file cannot be read")
		} else {
			err = json.Unmarshal(frBytes, &registry)
			if err != nil {
				log.WithError(err).WithField("path", registryDBFile).Info("registry file cannot be unmarshaled, ignoring")
			} else {
				log.WithField("path", registry).Debug("Using registry file")
			}
		}
	}

	ft := fileTail{
		pattern:        pattern,
		follow:         follow,
		registryDB:     registryDB,
		registryDBFile: registryDBFile,
		registryMutex:  sync.RWMutex{},
		registry:       registry,
		lC:             make(chan string),
		ctx:            ctx,
	}

	ft.watchFiles()
	return &ft, nil
}

func (ft *fileTail) ReadLine() chan string {
	return ft.lC
}

func (ft *fileTail) SaveState() {
	if ft.registryDB {
		ft.registryMutex.Lock()
		defer ft.registryMutex.Unlock()

		frBytes, err := json.Marshal(ft.registry)
		if err != nil {
			log.WithError(err).WithField("path", ft.registry).Error("Could not marshal file registry")
		} else {
			err = os.WriteFile(ft.registryDBFile, frBytes, 0666)
			if err != nil {
				log.WithError(err).WithField("path", ft.registryDBFile).Error("Could not write file registry")
			}

			log.WithField("path", ft.registryDBFile).Info("Saved file tail state")
		}
	}
}

func (ft *fileTail) watchFiles() {
	go func() {
	watchLoop:
		for {
			newFiles, err := ft.listFiles()
			if err != nil {
				log.WithError(err).Error("could not list new files")
			} else {
				for _, fw := range newFiles {
					ft.wg.Add(1)
					fw := fw
					go func() {
						log.WithField("file", fw.t.Filename).Debug("Starting new file tailer")
						defer ft.wg.Done()
						for {
							select {
							case <-ft.ctx.Done():
								log.WithField("file", fw.t.Filename).Info("Closing")
								return
							case l, ok := <-fw.t.Lines:
								if l != nil && l.Err != nil {
									log.WithError(err).WithField("file", fw.t.Filename).Error("could not read lines, closing")
									return
								}

								if !ok {
									log.WithField("file", fw.t.Filename).Info("File closed")
									return
								}

								ft.lC <- l.Text
								ft.registryMutex.RLock()
								if fw.Fm.Offset > l.SeekInfo.Offset {
									log.WithField("file", fw.t.Filename).WithField("fm_offset", fw.Fm.Offset).WithField("offset", l.SeekInfo.Offset).Debug("Detected truncation")
									fw.Fm.Prefix = []byte{}
									fw.Fm.PrefixLength = 0
									fw.Fm.Offset = 0
								}

								fw.Fm.Offset += int64(len(l.Text)) + 1
								if fw.Fm.PrefixLength < 1000 {
									fw.Fm.Prefix = append(fw.Fm.Prefix, []byte(l.Text)...)
									fw.Fm.Prefix = append(fw.Fm.Prefix, []byte("\n")...)
									fw.Fm.PrefixLength += len(l.Text) + 1
								}
								ft.registryMutex.RUnlock()
							}
						}
					}()
				}
			}

			if !ft.follow {
				break
			}

			select {
			case <-ft.ctx.Done():
				break watchLoop
			case <-time.NewTicker(30 * time.Second).C:
			}
		}

		ft.wg.Wait()
		close(ft.lC)
	}()
}

func (ft *fileTail) listFiles() ([]fileWatch, error) {
	matches, err := filepath.Glob(ft.pattern)
	if err != nil {
		return nil, fmt.Errorf("could not glob for pattern, %w", err)
	}

	newFileRegistry := map[string]fileWatch{}
	newFiles := []fileWatch{}

	log.WithField("files", matches).Debug("Matching files")

	ft.registryMutex.Lock()
	defer ft.registryMutex.Unlock()

nextFile:
	for _, m := range matches {
		fi, err := os.Stat(m)
		if err != nil {
			log.WithError(err).WithField("file", m).Warn("Could not get file info, skipping")
			continue
		}

		// file already monitored, move to new registry
		for k, v := range ft.registry {
			// check file as already monitored in current app run
			if v.fi != nil && os.SameFile(*v.fi, fi) {
				log.WithField("file", k).Debug("Same file detected")
				newFileRegistry[m] = fileWatch{fi: &fi, t: v.t, Fm: v.Fm}
				delete(ft.registry, k)
				continue nextFile
			}

			// check if file was monitored in prev app run
			if k == m && v.t == nil && v.Fm != nil && v.Fm.Offset != 0 && v.Fm.PrefixLength > 0 {
				f, err := os.Open(m)
				if err != nil {
					log.WithError(err).WithField("file", m).Warn("Could not open file, skipping")
					continue nextFile
				}
				prefixBytes := make([]byte, v.Fm.PrefixLength)
				n, err := f.Read(prefixBytes)
				if n != v.Fm.PrefixLength || err != nil {
					log.WithError(err).WithField("file", m).Debug("Not a matching file, skipping")
				} else if bytes.Equal(prefixBytes, v.Fm.Prefix) {
					log.WithField("file", k).WithField("offset", v.Fm.Offset).Debug("Previous file detected")
					t, err := tail.TailFile(m, tail.Config{Follow: ft.follow, Logger: log.WithField("TAIL", m), Location: &tail.SeekInfo{Offset: v.Fm.Offset, Whence: io.SeekStart}})
					if err != nil {
						log.WithError(err).WithField("file", m).Warn("Could not start new tail, skipping")
						continue
					}

					newFileRegistry[m] = fileWatch{fi: &fi, t: t, Fm: v.Fm}
					delete(ft.registry, k)
					newFiles = append(newFiles, newFileRegistry[m])
					continue nextFile
				} else {
					log.WithError(err).WithField("file", m).WithField("prefixStored", string(v.Fm.Prefix)).WithField("prefixRead", string(prefixBytes)).Debug("Prefix mismatch, skipping")
				}
			}
		}

		// it is new file to monitor
		t, err := tail.TailFile(m, tail.Config{Follow: ft.follow, Logger: log.WithField("TAIL", m)})
		if err != nil {
			log.WithError(err).WithField("file", m).Warn("Could not start new tail, skipping")
			continue
		}

		newFileRegistry[m] = fileWatch{fi: &fi, t: t, Fm: &fileMonitor{}}
		newFiles = append(newFiles, newFileRegistry[m])
		log.WithField("file", m).Debug("Monitoring new file")
	}

	for _, v := range ft.registry {
		if v.t != nil {
			log.WithField("file", v.t.Filename).Info("Closing tail for removed files")
			v.t.StopAtEOF()
		}
	}

	ft.registry = newFileRegistry
	return newFiles, nil
}
