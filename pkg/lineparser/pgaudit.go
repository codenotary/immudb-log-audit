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
	"fmt"
	"strconv"
	"strings"
)

type pgAuditEntry struct {
	AuditType      string `json:"audit_type"`
	StatementID    int    `json:"statement_id"`
	SubstatementID int    `json:"substatement_id"`
	Class          string `json:"class,omitempty"`
	Command        string `json:"command,omitempty"`
	ObjectType     string `json:"object_type,omitempty"`
	ObjectName     string `json:"object_name,omitempty"`
	Statement      string `json:"statement,omitempty"`
	Parameter      string `json:"parameter,omitempty"`
}

// converts pgaudit log line after AUDIT:
func toPgauditEntry(s string) (*pgAuditEntry, error) {
	csvReader := csv.NewReader(strings.NewReader(s))
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

	pgae := &pgAuditEntry{
		AuditType:      csvFields[0],
		StatementID:    statementID,
		SubstatementID: substatementID,
		Class:          csvFields[3],
		Command:        csvFields[4],
		ObjectType:     csvFields[5],
		ObjectName:     csvFields[6],
		Statement:      csvFields[7],
		Parameter:      csvFields[8],
	}

	return pgae, nil
}
