package source

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
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
	pattern      string
	follow       bool
	fileRegistry map[string]fileWatch
	lC           chan string
	wg           sync.WaitGroup
	ctx          context.Context
}

func NewFileTail(ctx context.Context, pattern string, follow bool) (*fileTail, error) {

	fileRegistry := map[string]fileWatch{}

	frBytes, err := os.ReadFile("fileregistry.json")
	if err != nil {
		log.WithError(err).Info("fileregistry.json cannot be read")
	} else {
		err = json.Unmarshal(frBytes, &fileRegistry)
		if err != nil {
			log.WithError(err).Info("fileregistry.json cannot be unmarshaled, ignoring")
		} else {
			log.WithField("fileregistry", fileRegistry).Debug("Using fileregistry.json")
		}
	}

	ft := fileTail{
		pattern:      pattern,
		follow:       follow,
		fileRegistry: fileRegistry,
		lC:           make(chan string),
		ctx:          ctx,
	}

	ft.watchFiles()
	return &ft, nil
}

func (ft *fileTail) ReadLine() (string, error) {
	l, ok := <-ft.lC
	if !ok {
		return l, io.EOF
	}

	return l, nil
}

func (ft *fileTail) saveRegistry() {
	frBytes, err := json.Marshal(ft.fileRegistry)
	if err != nil {
		log.WithError(err).Error("Could not marshal file registry")
	} else {
		err = os.WriteFile("fileregistry.json", frBytes, 0666)
		if err != nil {
			log.WithError(err).Error("Could not write file registry")
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
								if fw.Fm.Offset > l.SeekInfo.Offset {
									log.WithField("file", fw.t.Filename).WithField("fm_offset", fw.Fm.Offset).WithField("offset", l.SeekInfo.Offset).Debug("Detected truncation")
									fw.Fm.Prefix = []byte{}
									fw.Fm.PrefixLength = 0
									fw.Fm.Offset = 0
								}

								fw.Fm.Offset += int64(len(l.Text))
								if fw.Fm.PrefixLength < 1000 {
									fw.Fm.Prefix = append(fw.Fm.Prefix, []byte(l.Text)...)
									fw.Fm.Prefix = append(fw.Fm.Prefix, []byte("\n")...)
									fw.Fm.PrefixLength += len(l.Text) + 1
								}
							}
						}
					}()
				}
			}

			ft.saveRegistry()

			select {
			case <-ft.ctx.Done():
				break watchLoop
			case <-time.NewTicker(30 * time.Second).C:
			}
		}

		ft.wg.Wait()
		close(ft.lC)
		ft.saveRegistry()
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
nextFile:
	for _, m := range matches {
		fi, err := os.Stat(m)
		if err != nil {
			log.WithError(err).WithField("file", m).Warn("Could not get file info, skipping")
			continue
		}

		// file already monitored, move to new registry
		for k, v := range ft.fileRegistry {
			// check file as already monitored in current app run
			if v.fi != nil && os.SameFile(*v.fi, fi) {
				log.WithField("file", k).Debug("Same file detected")
				newFileRegistry[m] = fileWatch{fi: &fi, t: v.t, Fm: v.Fm}
				delete(ft.fileRegistry, k)
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
					log.WithField("file", k).Debug("Previous file detected")
					t, err := tail.TailFile(m, tail.Config{Follow: ft.follow, Logger: log.WithField("TAIL", m), Location: &tail.SeekInfo{Offset: v.Fm.Offset, Whence: 0}})
					if err != nil {
						log.WithError(err).WithField("file", m).Warn("Could not start new tail, skipping")
						continue
					}

					newFileRegistry[m] = fileWatch{fi: &fi, t: t, Fm: v.Fm}
					delete(ft.fileRegistry, k)
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

	for _, v := range ft.fileRegistry {
		if v.t != nil {
			log.WithField("file", v.t.Filename).Info("Closing tail for removed files")
			v.t.StopAtEOF()
		}
	}

	ft.fileRegistry = newFileRegistry
	return newFiles, nil
}
