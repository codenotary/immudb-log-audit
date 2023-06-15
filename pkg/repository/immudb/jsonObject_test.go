package immudb

import (
	"testing"

	"github.com/codenotary/immudb-log-audit/pkg/client/immudb"
	"github.com/codenotary/immudb-log-audit/test/utils"
	"github.com/stretchr/testify/assert"
)

func TestObjects(t *testing.T) {
	_, cli, containerID := utils.RunImmudbContainer()
	defer utils.StopImmudbContainer(containerID)

	err := SetupJsonObjectRepository(cli, "test", []immudb.CollectionField{
		{
			Name: "index1",
			Type: "STRING",
		},
	}, []immudb.CollectionIndex{
		{
			Fields:   []string{"index1"},
			IsUnique: true,
		},
	})
	assert.NoError(t, err)

	jr, err := NewJsonObjectRepository(cli, "test")
	assert.NoError(t, err)

	type testJSON struct {
		Index1   string            `json:"index1,omitempty"`
		Index2   bool              `json:"index2"`
		Index3   int               `json:"index3,omitempty"`
		Index4   float64           `json:"index4,omitempty"`
		Payload1 map[string]string `json:"payload1,omitempty"`
	}

	_, err = jr.Write(testJSON{
		Index1: "A",
	})
	assert.NoError(t, err)

	_, err = jr.Write(testJSON{
		Index1: "A",
	})
	assert.Error(t, err)
}

// func TestObjectsManual(t *testing.T) {

// 	_, _, containerID := utils.RunImmudbContainer()
// 	defer utils.StopImmudbContainer(containerID)

// 	cli, err := immuCliHttp.NewClientWithResponses(fmt.Sprintf("http://127.0.0.1:22222/api/v2"))
// 	assert.NoError(t, err)

// 	database := "defaultdb"
// 	pw := "immudb"
// 	osRes, err := cli.OpenSessionWithResponse(context.TODO(), immuCliHttp.ImmudbmodelOpenSessionRequest{
// 		Database: &database,
// 		Password: &pw,
// 		Username: &pw,
// 	})
// 	assert.NoError(t, err)
// 	fmt.Printf("%s\n", string(osRes.Body))

// 	sessionID := *osRes.JSON200.SessionID
// 	idName := "_id"
// 	isUnique := false
// 	res, err := cli.CreateCollection(context.TODO(), "mycollection", immuCliHttp.ModelCreateCollectionRequest{
// 		DocumentIdFieldName: &idName,
// 		Fields: &[]immuCliHttp.ModelField{
// 			{
// 				Name: "field1",
// 				Type: immuCliHttp.INTEGER,
// 			},
// 		},
// 		Indexes: &[]immuCliHttp.ModelIndex{
// 			immuCliHttp.ModelIndex{
// 				Fields:   &[]string{"field1"},
// 				IsUnique: &isUnique,
// 			},
// 		},
// 	}, func(ctx context.Context, req *http.Request) error {
// 		req.Header.Set("sessionid", sessionID)
// 		return nil
// 	})
// 	assert.NoError(t, err)
// 	b, err := io.ReadAll(res.Body)
// 	assert.NoError(t, err)
// 	fmt.Printf("%s\n", string(b))

// 	cRes, err := cli.GetCollectionsWithResponse(context.TODO(), func(ctx context.Context, req *http.Request) error {
// 		req.Header.Set("sessionid", sessionID)
// 		return nil
// 	})
// 	assert.NoError(t, err)
// 	b, _ = json.Marshal(*cRes.JSON200.Collections)
// 	fmt.Printf("COLLECTIONS: %s\n", string(b))
// }
