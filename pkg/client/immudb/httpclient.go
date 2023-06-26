package immudb

import (
	"context"
	"fmt"
	"net/http"
	"time"

	immuCliHttp "github.com/codenotary/immudb/pkg/api/httpclient"
)

type HTTPClient struct {
	sessionID       string
	client          immuCliHttp.ClientWithResponsesInterface
	keepAliveCancel context.CancelFunc
}

func NewHTTPClient(ctx context.Context, client immuCliHttp.ClientWithResponsesInterface, database string, user string, password string) (*HTTPClient, error) {
	res, err := client.OpenSessionWithResponse(ctx, immuCliHttp.ImmudbmodelOpenSessionRequest{
		Database: &database,
		Username: &user,
		Password: &password,
	})

	if err != nil {
		return nil, fmt.Errorf("could not open session for immudb, %w", err)
	}

	if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("could not open session for immudb, status not OK, %s", res.Status())
	}

	keepAliveCtx, keepAliveCancel := context.WithCancel(ctx)

	c := &HTTPClient{
		sessionID:       *res.JSON200.SessionID,
		client:          client,
		keepAliveCancel: keepAliveCancel,
	}

	go c.keepAlive(keepAliveCtx)

	return c, nil
}

type CollectionField struct {
	Name string
	Type string
}

type CollectionIndex struct {
	Fields   []string
	IsUnique bool
}

func (c *HTTPClient) CreateCollection(ctx context.Context, name string, fields []CollectionField, indexes []CollectionIndex) error {

	modelFields := []immuCliHttp.ModelField{}
	for _, f := range fields {
		modelFields = append(modelFields, immuCliHttp.ModelField{
			Name: f.Name,
			Type: immuCliHttp.ModelFieldType(f.Type),
		})
	}

	modelIndexes := []immuCliHttp.ModelIndex{}
	for _, i := range indexes {
		modelIndexes = append(modelIndexes, immuCliHttp.ModelIndex{
			Fields:   &i.Fields,
			IsUnique: &i.IsUnique,
		})
	}

	res, err := c.client.CreateCollectionWithResponse(ctx, name, immuCliHttp.ModelCreateCollectionRequest{
		Fields:  &modelFields,
		Indexes: &modelIndexes,
	}, func(ctx context.Context, req *http.Request) error {
		req.Header.Set("sessionid", c.sessionID)
		return nil
	})

	if err != nil {
		return fmt.Errorf("error creating collection: %w", err)
	}

	if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("could not create collection, status not OK, %s", res.Status())
	}

	return nil
}

func (c *HTTPClient) InsertDocument(ctx context.Context, collection string, document map[string]interface{}) error {
	res, err := c.client.InsertDocumentsWithResponse(ctx, collection, immuCliHttp.ModelInsertDocumentsRequest{
		Documents: &[]map[string]interface{}{
			document,
		},
	}, func(ctx context.Context, req *http.Request) error {
		req.Header.Set("sessionid", c.sessionID)
		return nil
	})

	if err != nil {
		return fmt.Errorf("error inserting document: %w", err)
	}

	if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("could not insert document, status not OK, %s", res.Status())
	}

	sres, err := c.client.SearchDocumentsWithResponse(ctx, collection, immuCliHttp.ModelSearchDocumentsRequest{
		Page:     1,
		PageSize: 10,
	}, func(ctx context.Context, req *http.Request) error {
		req.Header.Set("sessionid", c.sessionID)
		return nil
	})

	if err != nil {
		return fmt.Errorf("error searching document: %w", err)
	}

	for _, r := range *sres.JSON200.Revisions {
		fmt.Printf("SEARCH: %+v\n", *r.Document)
	}

	return nil
}

func (c *HTTPClient) Close() {
	c.keepAliveCancel()
	c.client.CloseSessionWithResponse(context.Background(), make(map[string]interface{}), func(ctx context.Context, req *http.Request) error {
		req.Header.Set("sessionid", c.sessionID)
		return nil
	})
}

func (c *HTTPClient) keepAlive(ctx context.Context) {
	for {
		c.client.KeepAliveWithResponse(context.TODO(), make(map[string]interface{}), func(ctx context.Context, req *http.Request) error {
			req.Header.Set("sessionid", c.sessionID)
			return nil
		})

		t := time.NewTimer(5 * time.Second)
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}
