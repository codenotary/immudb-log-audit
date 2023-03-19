package lineparser

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPgauditParse(t *testing.T) {
	type testData struct {
		line      string
		expected  *PGAuditEntry
		expectErr bool
	}

	tdd := []testData{
		{
			line: `2023-02-03 21:15:01.759 GMT [294] LOG:  AUDIT: SESSION,1,1,WRITE,INSERT,,,"insert into audit_trail(id, ts, usr, action, sourceip, context) VALUES ('c06984ff-ea4b-44e1-a7ff-d08376180614', NOW(), 'user0', 1, '127.0.0.1', 'some context')",<not logged>`,
			expected: &PGAuditEntry{
				Timestamp:      time.Now(),
				LogTimestamp:   time.Date(2023, 02, 03, 21, 15, 1, 759*1000000, func() *time.Location { l, _ := time.LoadLocation("GMT"); return l }()),
				AuditType:      "SESSION",
				StatementID:    1,
				SubstatementID: 1,
				Class:          "WRITE",
				Command:        "INSERT",
				ObjectType:     "",
				ObjectName:     "",
				Statement:      "insert into audit_trail(id, ts, usr, action, sourceip, context) VALUES ('c06984ff-ea4b-44e1-a7ff-d08376180614', NOW(), 'user0', 1, '127.0.0.1', 'some context')",
				Parameter:      "<not logged>",
			},
			expectErr: false,
		},
		{
			line:      `some invalid line that cannot be parsed`,
			expected:  nil,
			expectErr: true,
		},
	}

	pga := NewPGAuditLineParser()

	for _, td := range tdd {
		b, err := pga.Parse(td.line)
		if td.expectErr {
			assert.Error(t, err)
			assert.Nil(t, b)
			continue
		}

		var entry PGAuditEntry
		assert.NoError(t, json.Unmarshal(b, &entry))
		assert.Equal(t, td.expected.AuditType, entry.AuditType)
		assert.Equal(t, td.expected.Class, entry.Class)
		assert.Equal(t, td.expected.Command, entry.Command)
		assert.WithinDuration(t, td.expected.LogTimestamp, entry.LogTimestamp, 0)
		assert.WithinDuration(t, td.expected.Timestamp, entry.Timestamp, 100*time.Millisecond)
		assert.Equal(t, td.expected.ObjectName, td.expected.ObjectName)
		assert.Equal(t, td.expected.ObjectType, td.expected.ObjectType)
		assert.Equal(t, td.expected.Parameter, td.expected.Parameter)
		assert.Equal(t, td.expected.Statement, td.expected.Statement)
		assert.Equal(t, td.expected.StatementID, td.expected.StatementID)
		assert.Equal(t, td.expected.SubstatementID, td.expected.SubstatementID)
	}
}
