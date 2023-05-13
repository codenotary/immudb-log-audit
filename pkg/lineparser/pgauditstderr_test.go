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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPgauditParse(t *testing.T) {
	type testData struct {
		line      string
		expected  *pgAuditStderrEntry
		expectErr bool
	}

	tdd := []testData{
		{
			line: `2023-02-03 21:15:01.759 GMT [294] LOG:  AUDIT: SESSION,1,1,WRITE,INSERT,,,"insert into audit_trail(id, ts, usr, action, sourceip, context) VALUES ('c06984ff-ea4b-44e1-a7ff-d08376180614', NOW(), 'user0', 1, '127.0.0.1', 'some context')",<not logged>`,
			expected: &pgAuditStderrEntry{
				Timestamp:       time.Date(2023, 02, 03, 21, 15, 1, 759*1000000, func() *time.Location { l, _ := time.LoadLocation("GMT"); return l }()),
				ServerTimestamp: time.Now(),
				pgAuditEntry: pgAuditEntry{
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

		var entry pgAuditStderrEntry
		assert.NoError(t, json.Unmarshal(b, &entry))
		assert.NotEmpty(t, entry.UID)
		assert.Equal(t, td.expected.AuditType, entry.AuditType)
		assert.Equal(t, td.expected.Class, entry.Class)
		assert.Equal(t, td.expected.Command, entry.Command)
		assert.WithinDuration(t, td.expected.ServerTimestamp, entry.ServerTimestamp, 1*time.Millisecond)
		assert.WithinDuration(t, td.expected.Timestamp, entry.Timestamp, 100*time.Millisecond)
		assert.Equal(t, td.expected.ObjectName, td.expected.ObjectName)
		assert.Equal(t, td.expected.ObjectType, td.expected.ObjectType)
		assert.Equal(t, td.expected.Parameter, td.expected.Parameter)
		assert.Equal(t, td.expected.Statement, td.expected.Statement)
		assert.Equal(t, td.expected.StatementID, td.expected.StatementID)
		assert.Equal(t, td.expected.SubstatementID, td.expected.SubstatementID)
		assert.NotEmpty(t, entry.UID)
	}
}
