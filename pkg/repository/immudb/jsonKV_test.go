package immudb

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/codenotary/immudb-log-audit/test/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKV(t *testing.T) {
	immuCli, _, containerID := utils.RunImmudbContainer()
	defer utils.StopImmudbContainer(containerID)

	err := SetupJsonKVRepository(immuCli, "testkv", []string{"index1", "index2", "index3", "index4"})
	require.NoError(t, err)

	jr, err := NewJsonKVRepository(immuCli, "testkv")
	require.NoError(t, err)

	type testJSON struct {
		Index1   string            `json:"index1,omitempty"`
		Index2   bool              `json:"index2"`
		Index3   int               `json:"index3,omitempty"`
		Index4   float64           `json:"index4,omitempty"`
		Payload1 map[string]string `json:"payload1,omitempty"`
	}

	testInvalidJSONs := []testJSON{
		{
			Index2: true,
			Index3: 12,
			Index4: 3.5,
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
			Index3: 11,
			Index4: 33.5,
		},
		{
			Index1: "1",
			Index2: true,
			Index3: 1,
			Index4: 3.5,
		},
		{
			Index1: "2",
			Index2: false,
			Index3: 2,
			Index4: 2.5,
		},
		{
			Index1: "3",
			Index2: true,
			Index3: -3,
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
			Index3: 4,
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
		key         string
		prefix      string
		toRead      []*testJSON
	}

	trd := []testReadData{
		{
			testName: "Read all stored by index1",
			key:      "index1",
			prefix:   "",
			toRead:   []*testJSON{&testJSONs[1], &testJSONs[2], &testJSONs[3], &testJSONs[4], &testJSONs[5], &testJSONs[6]},
		},
		{
			testName: "Read one stored by index1",
			key:      "index1",
			prefix:   "5",
			toRead:   []*testJSON{&testJSONs[5]},
		},
		{
			testName: "Read all that are true by index2",
			key:      "index2",
			prefix:   "true",
			toRead:   []*testJSON{&testJSONs[1], &testJSONs[3]},
		},
		{
			testName: "Read all that are false by index2",
			key:      "index2",
			prefix:   "false",
			toRead:   []*testJSON{&testJSONs[2], &testJSONs[4], &testJSONs[5], &testJSONs[6]},
		},
		{
			testName: "Read by index3",
			key:      "index3",
			prefix:   "",
			toRead:   []*testJSON{&testJSONs[1], &testJSONs[2], &testJSONs[3], &testJSONs[6]},
		},
	}

	for _, tr := range trd {
		tr := tr
		t.Run(tr.testName, func(t *testing.T) {
			bb, err := jr.Read(tr.key, tr.prefix)
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

	type testReadHistory struct {
		testName     string
		shouldError  bool
		primaryValue string
		toRead       []*testJSON
	}

	trh := []testReadHistory{
		{
			testName:     "Two history entries for 1",
			shouldError:  false,
			primaryValue: "1",
			toRead:       []*testJSON{&testJSONs[0], &testJSONs[1]},
		},
		{
			testName:     "No history entries fo 111",
			shouldError:  true,
			primaryValue: "111",
		},
		{
			testName:     "One history entry for 6",
			shouldError:  false,
			primaryValue: "6",
			toRead:       []*testJSON{&testJSONs[6]},
		},
	}

	for _, tr := range trh {
		tr := tr
		t.Run(tr.testName, func(t *testing.T) {
			h, err := jr.History(tr.primaryValue)
			if tr.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, h, len(tr.toRead))
				for i := 0; i < len(tr.toRead); i++ {
					assert.Equal(t, uint64(i+1), h[i].Revision)
					var received testJSON
					err = json.Unmarshal(h[i].Entry, &received)
					assert.NoError(t, err)
					assert.EqualValues(t, *tr.toRead[i], received)
				}
			}
		})
	}
}
