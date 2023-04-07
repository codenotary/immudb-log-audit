package lineparser

import (
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
)

type PGAuditEntry struct {
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
func toPgauditEntry(s string) (*PGAuditEntry, error) {
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

	pgae := &PGAuditEntry{
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
