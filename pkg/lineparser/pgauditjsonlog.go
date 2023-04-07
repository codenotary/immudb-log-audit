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

package lineparser

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

type pgauditTimestamp struct {
	time.Time
}

func (pgt *pgauditTimestamp) UnmarshalJSON(bytes []byte) error {
	s := strings.Trim(string(bytes), "\"")
	ts, err := time.Parse("2006-01-02 15:04:05.000 MST", s)
	if err != nil {
		ts, err = time.Parse("2006-01-02 15:04:05 MST", s)
		if err != nil {
			return fmt.Errorf("could not unmarshal timestamp '%s': %w", string(bytes), err)
		}
	}

	pgt.Time = ts
	return nil
}

type PGAuditJSONLogEntry struct {
	PGAuditEntry
	UID             string           `json:"uid"`
	ServerTimestamp time.Time        `json:"server_timestamp"`
	Timestamp       pgauditTimestamp `json:"timestamp"`
	User            string           `json:"user"`
	DBName          string           `json:"dbname"`
	RemoteHost      string           `json:"remote_host"`
	RemotePort      int              `json:"remote_port"`
	SessionID       string           `json:"session_id"`
	LineNumber      int              `json:"line_num"`
	PS              string           `json:"ps,omitempty"`
	SessionStart    pgauditTimestamp `json:"session_start"`
}

type PGAuditJSONLogLineParser struct {
}

func NewPGAuditJSONLogLineParser() *PGAuditJSONLogLineParser {
	return &PGAuditJSONLogLineParser{}
}

func (p *PGAuditJSONLogLineParser) Parse(line string) ([]byte, error) {
	r := gjson.Get(line, "message")
	if !r.Exists() {
		return nil, errors.New("not a pgaudit line, missing 'messagae' field")
	}

	pgae, err := toPgauditEntry(r.String())
	if err != nil {
		return nil, fmt.Errorf("not a pgaudit line, %w", err)
	}

	var pgaje PGAuditJSONLogEntry
	err = json.Unmarshal([]byte(line), &pgaje)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal json log, %w", err)
	}

	pgaje.PGAuditEntry = *pgae
	pgaje.UID = uuid.New().String()
	pgaje.ServerTimestamp = time.Now().UTC()

	bytes, err := json.Marshal(pgaje)
	if err != nil {
		return nil, fmt.Errorf("could not marshal pg audit entry, %w", err)
	}

	return bytes, nil
}
