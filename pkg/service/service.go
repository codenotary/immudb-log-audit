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

package service

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

type lineProvider interface {
	ReadLine() chan string
	SaveState()
}

type LineParser interface {
	Parse(line string) ([]byte, error)
}

type JsonRepository interface {
	WriteBytes(b [][]byte) (uint64, error)
}

type AuditHistoryEntry struct {
	Entry    []byte
	Revision uint64
	TXID     uint64
}

type AuditService struct {
	lineProvider   lineProvider
	jsonRepository JsonRepository
	lineParser     LineParser
}

func NewAuditService(lineProvider lineProvider, lineParser LineParser, jsonRepository JsonRepository) *AuditService {
	return &AuditService{
		lineProvider:   lineProvider,
		lineParser:     lineParser,
		jsonRepository: jsonRepository,
	}
}

func (as *AuditService) Run() error {
	bufferSize := 200
	saveStateTicker := time.NewTicker(5 * time.Second)
	buf := [][]byte{}

	for stop := false; !stop; {
		select {
		case l, ok := <-as.lineProvider.ReadLine():
			if !ok {
				stop = true
				if l == "" {
					break
				}
			}
			b, err := as.lineParser.Parse(l)
			if err != nil {
				log.WithError(err).WithField("line", l).Debug("Invalid line format, skipping")
				if !stop {
					continue
				}
			}

			buf = append(buf, b)
			if len(buf) == bufferSize || (len(buf) > 0 && stop) {
				id, err := as.jsonRepository.WriteBytes(buf)
				if err != nil {
					return fmt.Errorf("could not store audit entry, %w", err)
				}

				log.WithField("TXID", id).WithField("line", string(buf[len(buf)-1])).Trace("Stored line")
				buf = [][]byte{}

				if stop {
					as.lineProvider.SaveState()
				}
			}
		case <-saveStateTicker.C:
			if len(buf) > 0 {
				id, err := as.jsonRepository.WriteBytes(buf)
				if err != nil {
					return fmt.Errorf("could not store audit entry, %w", err)
				}

				log.WithField("TXID", id).WithField("line", string(buf[len(buf)-1])).Trace("Stored line")
				buf = [][]byte{}
				as.lineProvider.SaveState()
			}

			saveStateTicker = time.NewTicker(5 * time.Second)
		}
	}

	return nil
}
