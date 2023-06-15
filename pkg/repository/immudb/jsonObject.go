package immudb

import (
	"context"
	"encoding/json"
	"fmt"

	immuCliHttp "github.com/codenotary/immudb-log-audit/pkg/client/immudb"
)

type JsonObjectRepository struct {
	client     *immuCliHttp.HTTPClient
	collection string
}

func NewJsonObjectRepository(client *immuCliHttp.HTTPClient, collection string) (*JsonObjectRepository, error) {
	return &JsonObjectRepository{
		client:     client,
		collection: collection,
	}, nil
}

func (jr *JsonObjectRepository) Write(jObject interface{}) (uint64, error) {
	b, err := json.Marshal(jObject)
	if err != nil {
		return 0, fmt.Errorf("could not marshal jObject, %w", err)
	}

	return jr.WriteBytes(b)
}

func (jr *JsonObjectRepository) WriteBytes(jBytes []byte) (uint64, error) {
	var mi map[string]interface{}
	err := json.Unmarshal(jBytes, &mi)
	if err != nil {
		return 0, fmt.Errorf("could not unmarshal jObject, %w", err)
	}

	err = jr.client.InsertDocument(context.Background(), jr.collection, mi)
	if err != nil {
		return 0, fmt.Errorf("could not insert document: %w", err)
	}

	return 0, nil
}

func SetupJsonObjectRepository(client *immuCliHttp.HTTPClient, collection string, fields []immuCliHttp.CollectionField, indexes []immuCliHttp.CollectionIndex) error {
	ctx := context.Background()
	err := client.CreateCollection(ctx, collection, fields, indexes)
	if err != nil {
		return fmt.Errorf("could not setup json object repository, %w", err)
	}

	return nil
}
