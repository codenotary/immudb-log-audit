package lineparser

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type PGAuditEntry struct {
	Timestamp      time.Time `json:"timestamp"`
	LogTimestamp   time.Time `json:"log_timestamp"`
	AuditType      string    `json:"audit_type"`
	StatementID    int       `json:"statement_id"`
	SubstatementID int       `json:"substatement_id,omitempty"`
	Class          string    `json:"class,omitempty"`
	Command        string    `json:"command,omitempty"`
	ObjectType     string    `json:"object_type,omitempty"`
	ObjectName     string    `json:"object_name,omitempty"`
	Statement      string    `json:"statement,omitempty"`
	Parameter      string    `json:"parameter,omitempty"`
}

type PGAuditLineParser struct {
}

func NewPGAuditLineParser() *PGAuditLineParser {
	return &PGAuditLineParser{}
}

func (p *PGAuditLineParser) Parse(line string) ([]byte, error) {
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

	pgae := &PGAuditEntry{
		Timestamp:      time.Now().UTC(),
		LogTimestamp:   ts,
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

	bytes, err := json.Marshal(pgae)
	if err != nil {
		return nil, fmt.Errorf("could not marshal pg audit entry, %w", err)
	}

	return bytes, nil
}
