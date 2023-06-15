package immudb

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/codenotary/immudb-log-audit/test/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQL(t *testing.T) {

	immuCli, _, containerID := utils.RunImmudbContainer()
	defer utils.StopImmudbContainer(containerID)

	//"id=INTEGER AUTO_INCREMENT", "user=VARCHAR[256]", "dbname=VARCHAR[256]", "session_id=VARCHAR[256]", "statement_id=INTEGER", "substatement_id=INTEGER", "server_timestamp=TIMESTAMP", "timestamp=TIMESTAMP", "audit_type=VARCHAR[256]", "class=VARCHAR[256]", "command=VARCHAR[256]"}
	err := SetupJsonSQLRepository(immuCli, "testsql", "index1", []string{"index1=VARCHAR[257]", "index2=BOOLEAN", "index3=TIMESTAMP", "index4=FLOAT", "index5=INTEGER"})
	require.NoError(t, err)

	jr, err := NewJsonSQLRepository(immuCli, "testsql")
	require.NoError(t, err)
	assert.NotNil(t, jr)

	type testJSON struct {
		Index1   string            `json:"index1,omitempty"`
		Index2   bool              `json:"index2"`
		Index3   time.Time         `json:"index3,omitempty"`
		Index4   float64           `json:"index4,omitempty"`
		Index5   int               `json:"index5,omitempty"`
		Payload1 map[string]string `json:"payload1,omitempty"`
	}

	testInvalidJSONs := []testJSON{
		{},
		{
			Index2: true,
		},
	}

	for i, tw := range testInvalidJSONs {
		tw := tw
		t.Run(fmt.Sprintf("Test invalid write %d", i), func(t *testing.T) {
			_, err := jr.Write(tw)
			assert.Error(t, err)
		})
	}

	testJSONs := []testJSON{
		{
			Index1: "1",
			Index2: false,
			Index3: time.Date(2023, time.May, 1, 1, 1, 0, 0, time.UTC),
			Index4: 33.5,
		},
		{
			Index1: "1",
			Index2: true,
			Index3: time.Date(2023, time.May, 1, 1, 1, 0, 0, time.UTC),
			Index4: 3.5,
		},
		{
			Index1: "2",
			Index2: false,
			Index3: time.Date(2023, time.May, 1, 1, 2, 0, 0, time.UTC),
			Index4: 2.5,
		},
		{
			Index1: "3",
			Index2: true,
			Index3: time.Date(2023, time.May, 1, 1, 3, 0, 0, time.UTC),
			Index4: -3.5,
		},
		{
			Index1: "4",
		},
		{
			Index1: "5",
			Index2: false,
		},
		{
			Index1: "6",
			Index2: false,
			Index3: time.Date(2023, time.May, 1, 1, 6, 0, 0, time.UTC),
			Payload1: map[string]string{
				"a": "b",
				"c": "d",
			},
		},
	}

	for i, tw := range testJSONs {
		tw := tw
		t.Run(fmt.Sprintf("Test write %d", i), func(t *testing.T) {
			_, err := jr.Write(tw)
			assert.NoError(t, err)
		})
	}

	type testReadData struct {
		testName    string
		shouldError bool
		query       string
		toRead      []*testJSON
	}

	trd := []testReadData{
		{
			testName: "Read all stored",
			query:    "",
			toRead:   []*testJSON{&testJSONs[1], &testJSONs[2], &testJSONs[3], &testJSONs[4], &testJSONs[5], &testJSONs[6]},
		},
		{
			testName: "Read one stored by index1",
			query:    "index1='6'",
			toRead:   []*testJSON{&testJSONs[6]},
		},
		{
			testName: "Read one stored by index1",
			query:    "index2=False",
			toRead:   []*testJSON{&testJSONs[2], &testJSONs[4], &testJSONs[5], &testJSONs[6]},
		},
	}

	for _, tr := range trd {
		tr := tr
		t.Run(tr.testName, func(t *testing.T) {
			bb, err := jr.Read(tr.query)
			if tr.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, bb, len(tr.toRead))
				for i := range tr.toRead {
					var received testJSON
					err := json.Unmarshal(bb[i], &received)
					assert.NoError(t, err)
					for j, trd := range tr.toRead {
						if received.Index1 == trd.Index1 {
							assert.EqualValues(t, *trd, received)
							break
						}
						assert.Less(t, j, len(tr.toRead))
					}
				}
			}
		})
	}
}
