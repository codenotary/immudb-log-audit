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
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type pgAuditStderrEntry struct {
	pgAuditEntry
	UID             string    `json:"uid"`
	Timestamp       time.Time `json:"timestamp"`        // timestamp from log line
	ServerTimestamp time.Time `json:"server_timestamp"` // server timestamp
}

type pgAuditLineParser struct {
}

func NewPGAuditLineParser() *pgAuditLineParser {
	return &pgAuditLineParser{}
}

func (p *pgAuditLineParser) Parse(line string) ([]byte, error) {
	// assumed default log_line_prefix '%m [%p] '
	if len(line) < 26 { // min length of timestamp with timezone
		return nil, fmt.Errorf("invalid log line prefix, too short")
	}
	cur := 26
	pos := strings.Index(line[cur:], " ") // find end of timezone abbreviation
	if pos < 0 {
		return nil, fmt.Errorf("invalid log line prefix")
	}

	cur += pos
	ts, err := time.Parse("2006-01-02 15:04:05.000 MST", line[:cur])
	if err != nil {
		return nil, fmt.Errorf("could not parse timestamp '%s': %w", line[:cur], err)
	}

	pos = strings.Index(line[cur:], "AUDIT: ")
	if pos < 0 {
		return nil, fmt.Errorf("not a pgaudit line")
	}

	cur += pos + len("AUDIT: ")
	csvReader := csv.NewReader(strings.NewReader(line[cur:]))
	csvFields, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("invalid csv line, %w", err)
	}

	if len(csvFields) < 9 {
		return nil, fmt.Errorf("invalid csv fields length: %d", len(csvFields))
	}

	statementID, err := strconv.Atoi(csvFields[1])
	if err != nil {
		return nil, fmt.Errorf("could not parse statementID, %w", err)
	}

	substatementID, err := strconv.Atoi(csvFields[2])
	if err != nil {
		return nil, fmt.Errorf("could not parse substatementID, %w", err)
	}

	pgae := &pgAuditStderrEntry{
		UID:             uuid.New().String(),
		ServerTimestamp: time.Now().UTC(),
		Timestamp:       ts,
		pgAuditEntry: pgAuditEntry{
			AuditType:      csvFields[0],
			StatementID:    statementID,
			SubstatementID: substatementID,
			Class:          csvFields[3],
			Command:        csvFields[4],
			ObjectType:     csvFields[5],
			ObjectName:     csvFields[6],
			Statement:      csvFields[7],
			Parameter:      csvFields[8],
		},
	}

	bytes, err := json.Marshal(pgae)
	if err != nil {
		return nil, fmt.Errorf("could not marshal pg audit entry, %w", err)
	}

	return bytes, nil
}
