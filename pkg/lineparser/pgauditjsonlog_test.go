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

// {"timestamp":"2023-05-13 21:09:08.502 GMT","user":"postgres","dbname":"postgres","pid":138,"remote_host":"172.22.0.1","remote_port":58300,"session_id":"645ffc74.8a","line_num":1,"ps":"CREATE TABLE","session_start":"2023-05-13 21:09:08 GMT","vxid":"3/44","txid":736,"error_severity":"LOG","message":"AUDIT: SESSION,1,1,DDL,CREATE TABLE,,,\"create table if not exists audit_trail (id VARCHAR, ts TIMESTAMP, usr VARCHAR, action INTEGER, sourceip VARCHAR, context VARCHAR, PRIMARY KEY(id));\",<not logged>","backend_type":"client backend","query_id":0}
// {"timestamp":"2023-05-13 21:09:08.505 GMT","user":"postgres","dbname":"postgres","pid":138,"remote_host":"172.22.0.1","remote_port":58300,"session_id":"645ffc74.8a","line_num":2,"ps":"INSERT","session_start":"2023-05-13 21:09:08 GMT","vxid":"3/45","txid":0,"error_severity":"LOG","message":"AUDIT: SESSION,2,1,WRITE,INSERT,,,\"insert into audit_trail(id, ts, usr, action, sourceip, context) VALUES ('3a5719ef-258e-4416-a352-82c1e674f2ef', NOW(), 'user0', 1, '127.0.0.1', 'some context')\",<not logged>","backend_type":"client backend","query_id":0}
// {"timestamp":"2023-05-13 21:09:13.511 GMT","user":"postgres","dbname":"postgres","pid":138,"remote_host":"172.22.0.1","remote_port":58300,"session_id":"645ffc74.8a","line_num":3,"ps":"INSERT","session_start":"2023-05-13 21:09:08 GMT","vxid":"3/46","txid":0,"error_severity":"LOG","message":"AUDIT: SESSION,3,1,WRITE,INSERT,,,\"insert into audit_trail(id, ts, usr, action, sourceip, context) VALUES ('d48c0c4b-b77b-403e-97c0-b711f630cc03', NOW(), 'user1', 1, '127.0.0.1', 'some context')\",<not logged>","backend_type":"client backend","query_id":0}
// {"timestamp":"2023-05-13 21:09:18.518 GMT","user":"postgres","dbname":"postgres","pid":138,"remote_host":"172.22.0.1","remote_port":58300,"session_id":"645ffc74.8a","line_num":4,"ps":"INSERT","session_start":"2023-05-13 21:09:08 GMT","vxid":"3/47","txid":0,"error_severity":"LOG","message":"AUDIT: SESSION,4,1,WRITE,INSERT,,,\"insert into audit_trail(id, ts, usr, action, sourceip, context) VALUES ('348c6067-6e3a-482f-8b81-86ca12d6c5ed', NOW(), 'user2', 1, '127.0.0.1', 'some context')\",<not logged>","backend_type":"client backend","query_id":0}
// {"timestamp":"2023-05-13 21:09:23.521 GMT","user":"postgres","dbname":"postgres","pid":138,"remote_host":"172.22.0.1","remote_port":58300,"session_id":"645ffc74.8a","line_num":5,"ps":"INSERT","session_start":"2023-05-13 21:09:08 GMT","vxid":"3/48","txid":0,"error_severity":"LOG","message":"AUDIT: SESSION,5,1,WRITE,INSERT,,,\"insert into audit_trail(id, ts, usr, action, sourceip, context) VALUES ('be730d06-ff8f-42f1-8679-f535b0680446', NOW(), 'user3', 1, '127.0.0.1', 'some context')\",<not logged>","backend_type":"client backend","query_id":0}
// {"timestamp":"2023-05-13 21:09:28.527 GMT","user":"postgres","dbname":"postgres","pid":138,"remote_host":"172.22.0.1","remote_port":58300,"session_id":"645ffc74.8a","line_num":6,"ps":"INSERT","session_start":"2023-05-13 21:09:08 GMT","vxid":"3/49","txid":0,"error_severity":"LOG","message":"AUDIT: SESSION,6,1,WRITE,INSERT,,,\"insert into audit_trail(id, ts, usr, action, sourceip, context) VALUES ('34b8cab9-0796-4ae7-ae83-3f49ccd23391', NOW(), 'user4', 1, '127.0.0.1', 'some context')\",<not logged>","backend_type":"client backend","query_id":0}
// {"timestamp":"2023-05-13 21:09:33.535 GMT","user":"postgres","dbname":"postgres","pid":138,"remote_host":"172.22.0.1","remote_port":58300,"session_id":"645ffc74.8a","line_num":7,"ps":"INSERT","session_start":"2023-05-13 21:09:08 GMT","vxid":"3/50","txid":0,"error_severity":"LOG","message":"AUDIT: SESSION,7,1,WRITE,INSERT,,,\"insert into audit_trail(id, ts, usr, action, sourceip, context) VALUES ('8503fa00-1a1e-4329-8740-83b8c28b4c81', NOW(), 'user5', 1, '127.0.0.1', 'some context')\",<not logged>","backend_type":"client backend","query_id":0}
// {"timestamp":"2023-05-13 21:09:38.541 GMT","user":"postgres","dbname":"postgres","pid":138,"remote_host":"172.22.0.1","remote_port":58300,"session_id":"645ffc74.8a","line_num":8,"ps":"INSERT","session_start":"2023-05-13 21:09:08 GMT","vxid":"3/51","txid":0,"error_severity":"LOG","message":"AUDIT: SESSION,8,1,WRITE,INSERT,,,\"insert into audit_trail(id, ts, usr, action, sourceip, context) VALUES ('c8e4dac3-6d94-4f6e-acb6-ed0161b75407', NOW(), 'user6', 1, '127.0.0.1', 'some context')\",<not logged>","backend_type":"client backend","query_id":0}
// {"timestamp":"2023-05-13 21:09:43.548 GMT","user":"postgres","dbname":"postgres","pid":138,"remote_host":"172.22.0.1","remote_port":58300,"session_id":"645ffc74.8a","line_num":9,"ps":"INSERT","session_start":"2023-05-13 21:09:08 GMT","vxid":"3/52","txid":0,"error_severity":"LOG","message":"AUDIT: SESSION,9,1,WRITE,INSERT,,,\"insert into audit_trail(id, ts, usr, action, sourceip, context) VALUES ('fc432a5d-5823-4b04-b544-ef47c64f405e', NOW(), 'user7', 1, '127.0.0.1', 'some context')\",<not logged>","backend_type":"client backend","query_id":0}
// {"timestamp":"2023-05-13 21:09:48.551 GMT","user":"postgres","dbname":"postgres","pid":138,"remote_host":"172.22.0.1","remote_port":58300,"session_id":"645ffc74.8a","line_num":10,"ps":"INSERT","session_start":"2023-05-13 21:09:08 GMT","vxid":"3/53","txid":0,"error_severity":"LOG","message":"AUDIT: SESSION,10,1,WRITE,INSERT,,,\"insert into audit_trail(id, ts, usr, action, sourceip, context) VALUES ('4212c416-95cc-459e-b950-a86fee9b3a69', NOW(), 'user8', 1, '127.0.0.1', 'some context')\",<not logged>","backend_type":"client backend","query_id":0}
// {"timestamp":"2023-05-13 21:09:53.558 GMT","user":"postgres","dbname":"postgres","pid":138,"remote_host":"172.22.0.1","remote_port":58300,"session_id":"645ffc74.8a","line_num":11,"ps":"INSERT","session_start":"2023-05-13 21:09:08 GMT","vxid":"3/54","txid":0,"error_severity":"LOG","message":"AUDIT: SESSION,11,1,WRITE,INSERT,,,\"insert into audit_trail(id, ts, usr, action, sourceip, context) VALUES ('ec0960bb-1fed-458c-82c5-98a3fe19d1da', NOW(), 'user9', 1, '127.0.0.1', 'some context')\",<not logged>","backend_type":"client backend","query_id":0}
// {"timestamp":"2023-05-13 21:09:58.564 GMT","user":"postgres","dbname":"postgres","pid":138,"remote_host":"172.22.0.1","remote_port":58300,"session_id":"645ffc74.8a","line_num":12,"ps":"INSERT","session_start":"2023-05-13 21:09:08 GMT","vxid":"3/55","txid":0,"error_severity":"LOG","message":"AUDIT: SESSION,12,1,WRITE,INSERT,,,\"insert into audit_trail(id, ts, usr, action, sourceip, context) VALUES ('7ca3bea2-e8cd-4817-9336-c3d94cc0405b', NOW(), 'user10', 1, '127.0.0.1', 'some context')\",<not logged>","backend_type":"client backend","query_id":0}
func TestPgauditParseJSONLog(t *testing.T) {
	type testData struct {
		line      string
		expected  *pgAuditJSONLogEntry
		expectErr bool
	}

	tdd := []testData{
		{
			line: `{"timestamp":"2023-05-13 21:09:08.502 GMT","user":"postgres","dbname":"postgres","pid":138,"remote_host":"172.22.0.1","remote_port":58300,"session_id":"645ffc74.8a","line_num":1,"ps":"CREATE TABLE","session_start":"2023-05-13 21:09:08 GMT","vxid":"3/44","txid":736,"error_severity":"LOG","message":"AUDIT: SESSION,1,1,DDL,CREATE TABLE,,,\"create table if not exists audit_trail (id VARCHAR, ts TIMESTAMP, usr VARCHAR, action INTEGER, sourceip VARCHAR, context VARCHAR, PRIMARY KEY(id));\",<not logged>","backend_type":"client backend","query_id":0}`,
			expected: &pgAuditJSONLogEntry{
				Timestamp:       pgauditTimestamp{time.Date(2023, 05, 13, 21, 9, 8, 502*1000000, func() *time.Location { l, _ := time.LoadLocation("GMT"); return l }())},
				ServerTimestamp: time.Now(),
				User:            "postgres",
				DBName:          "postgres",
				RemoteHost:      "172.22.0.1",
				RemotePort:      58300,
				SessionID:       "645ffc74.8a",
				LineNumber:      1,
				PS:              "CREATE TABLE",
				SessionStart:    pgauditTimestamp{time.Date(2023, 05, 13, 21, 9, 8, 0, func() *time.Location { l, _ := time.LoadLocation("GMT"); return l }())},
				pgAuditEntry: pgAuditEntry{
					AuditType:      "SESSION",
					StatementID:    1,
					SubstatementID: 1,
					Class:          "DDL",
					Command:        "CREATE TABLE",
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

	pga := NewPGAuditJSONLogLineParser()

	for _, td := range tdd {
		b, err := pga.Parse(td.line)
		if td.expectErr {
			assert.Error(t, err)
			assert.Nil(t, b)
			continue
		}

		var entry pgAuditJSONLogEntry
		assert.NoError(t, json.Unmarshal(b, &entry))
		assert.NotEmpty(t, entry.UID)
		assert.Equal(t, td.expected.AuditType, entry.AuditType)
		assert.Equal(t, td.expected.Class, entry.Class)
		assert.Equal(t, td.expected.Command, entry.Command)
		assert.WithinDuration(t, td.expected.ServerTimestamp, entry.ServerTimestamp, 1*time.Millisecond)
		assert.WithinDuration(t, td.expected.Timestamp.Time, entry.Timestamp.Time, 100*time.Millisecond)
		assert.Equal(t, td.expected.ObjectName, td.expected.ObjectName)
		assert.Equal(t, td.expected.ObjectType, td.expected.ObjectType)
		assert.Equal(t, td.expected.Parameter, td.expected.Parameter)
		assert.Equal(t, td.expected.Statement, td.expected.Statement)
		assert.Equal(t, td.expected.StatementID, td.expected.StatementID)
		assert.Equal(t, td.expected.SubstatementID, td.expected.SubstatementID)
		assert.NotEmpty(t, entry.UID)
	}
}
