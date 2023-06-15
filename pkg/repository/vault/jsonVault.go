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

package vault

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	vaultclient "github.com/codenotary/immudb-log-audit/pkg/client/vault"
	log "github.com/sirupsen/logrus"
)

type JsonVaultRepository struct {
	client        vaultclient.ClientWithResponsesInterface
	ledger        string
	collection    string
	bulkMode      bool
	docBuf        []byte
	docBufCounter int
}

func NewJsonVaultRepository(client vaultclient.ClientWithResponsesInterface, ledger string, collection string, bulkMode bool) (*JsonVaultRepository, error) {
	return &JsonVaultRepository{
		client:     client,
		ledger:     ledger,
		collection: collection,
		bulkMode:   bulkMode,
		docBuf:     []byte(`{"documents": [`),
	}, nil
}

func (jv *JsonVaultRepository) WriteBytes(jBytes []byte) (uint64, error) {
	ctx := context.Background()
	var txID uint64
	if jv.bulkMode {
		//TODO: add ticker
		if jv.docBufCounter > 0 {
			jv.docBuf = append(jv.docBuf, ',')
		}

		jv.docBuf = append(jv.docBuf, jBytes...)
		jv.docBufCounter++
		if jv.docBufCounter < 100 {
			return 0, nil
		}

		jv.docBuf = append(jv.docBuf, []byte("]}")...)
		res, err := jv.client.DocumentCreateManyWithBodyWithResponse(ctx, jv.ledger, jv.collection, "application/json", bytes.NewBuffer(jv.docBuf))
		jv.docBuf = []byte(`{"documents": [`)
		jv.docBufCounter = 0
		if err != nil {
			return 0, fmt.Errorf("error writing document to vault, %w", err)
		}

		if res.JSON200 == nil {
			return 0, fmt.Errorf("error writing document to vault, %d, %s", res.StatusCode(), string(res.Body))
		}

		log.WithField("documentID", res.JSON200.DocumentIds[0]).WithField("txID", func() string {
			if res.JSON200.TransactionId != nil {
				return *res.JSON200.TransactionId
			}
			return ""
		}()).Debug("Created document")

		if res.JSON200.TransactionId != nil {
			txID, err = strconv.ParseUint(*res.JSON200.TransactionId, 10, 64)
			if err != nil {
				log.WithField("txID", *res.JSON200.TransactionId).WithError(err).Error("could not convert transactionID")
			}
		}
	} else {
		res, err := jv.client.DocumentCreateWithBodyWithResponse(ctx, jv.ledger, jv.collection, "application/json", bytes.NewReader(jBytes))
		if err != nil {
			return 0, fmt.Errorf("error writing document to vault, %w", err)
		}

		if res.JSON200 == nil {
			return 0, fmt.Errorf("error writing document to vault, %d, %s", res.StatusCode(), string(res.Body))
		}

		log.WithField("documentID", res.JSON200.DocumentId).WithField("txID", func() string {
			if res.JSON200.TransactionId != nil {
				return *res.JSON200.TransactionId
			}
			return ""
		}()).Debug("Created document")

		if res.JSON200.TransactionId != nil {
			txID, err = strconv.ParseUint(*res.JSON200.TransactionId, 10, 64)
			if err != nil {
				log.WithField("txID", *res.JSON200.TransactionId).WithError(err).Error("could not convert transactionID")
			}
		}
	}

	return txID, nil
}

func (jv *JsonVaultRepository) Write(jObject interface{}) (uint64, error) {
	ctx := context.Background()

	res, err := jv.client.DocumentCreateWithResponse(ctx, jv.ledger, jv.collection, jObject)
	if err != nil {
		return 0, fmt.Errorf("error writing document to vault, %w", err)
	}

	if res.JSON200 == nil {
		return 0, fmt.Errorf("error writing document to vault, %d, %s", res.StatusCode(), string(res.Body))
	}

	log.WithField("documentID", res.JSON200.DocumentId).WithField("txID", func() string {
		if res.JSON200.TransactionId != nil {
			return *res.JSON200.TransactionId
		}
		return ""
	}()).Debug("Created document")

	var txID uint64
	if res.JSON200.TransactionId != nil {
		txID, err = strconv.ParseUint(*res.JSON200.TransactionId, 10, 64)
		if err != nil {
			log.WithField("txID", *res.JSON200.TransactionId).WithError(err).Error("could not convert transactionID")
		}
	}

	return txID, nil
}

func (jv *JsonVaultRepository) Read(queryString string) ([][]byte, error) {
	ctx := context.Background()

	keepOpen := true
	req := vaultclient.SearchDocumentJSONRequestBody{
		Page:     1,
		PerPage:  100,
		KeepOpen: &keepOpen,
	}

	if queryString != "" {
		var query vaultclient.Query
		err := json.Unmarshal([]byte(queryString), &query)
		if err != nil {
			return nil, fmt.Errorf("invalid query, %w", err)
		}
		req.Query = &query
	}

	var documents [][]byte

	for {
		res, err := jv.client.SearchDocumentWithResponse(ctx, jv.ledger, jv.collection, req)
		if err != nil {
			return nil, fmt.Errorf("error querying vault, %w", err)
		}

		if res.JSON200 == nil {
			return nil, fmt.Errorf("error querying vault, %d, %s", res.StatusCode(), string(res.Body))
		}

		for _, d := range res.JSON200.Revisions {
			document, err := json.Marshal(d.Document)
			if err != nil {
				log.WithError(err).WithField("document", d.Document).Error("Could not marshal document")
				continue
			}

			documents = append(documents, document)
		}

		if res.JSON200.SearchId == "" || len(res.JSON200.Revisions) == 0 {
			break
		}

		req.Page++
		req.SearchId = &res.JSON200.SearchId
	}

	return documents, nil
}

func (jv *JsonVaultRepository) Audit(documentID string) ([][]byte, error) {
	ctx := context.Background()

	req := vaultclient.DocumentAuditRequest{
		Desc:    true,
		Page:    1,
		PerPage: 100,
	}

	var documents [][]byte

	for {
		res, err := jv.client.AuditDocumentWithResponse(ctx, jv.ledger, jv.collection, documentID, req)
		if err != nil {
			return nil, fmt.Errorf("error querying vault, %w", err)
		}

		if res.JSON200 == nil {
			return nil, fmt.Errorf("error querying vault, %d, %s", res.StatusCode(), string(res.Body))
		}

		for _, d := range res.JSON200.Revisions {
			document, err := json.Marshal(d)
			if err != nil {
				log.WithError(err).WithField("document", d).Error("Could not marshal document")
				continue
			}

			documents = append(documents, document)
		}

		if len(res.JSON200.Revisions) == 0 || len(res.JSON200.Revisions) < 100 {
			break
		}

		req.Page++
	}

	return documents, nil
}

func SetupJsonObjectRepository(client vaultclient.ClientWithResponsesInterface, ledger string, collection string, createRequest *vaultclient.CollectionCreateRequest) error {
	ctx := context.Background()
	resGet, err := client.CollectionGetWithResponse(ctx, ledger, collection)
	if err != nil {
		return fmt.Errorf("could not get collection,error %w", err)
	} else if resGet.JSON400 != nil {
		return fmt.Errorf("could not get collection, 400 %s, error %s", resGet.JSON400.Status, resGet.JSON400.Error)
	} else if resGet.JSON403 != nil {
		return fmt.Errorf("could not get collection, forbidden, error %s", resGet.JSON403.Error)
	} else if resGet.JSON500 != nil {
		return fmt.Errorf("could not read collection, 500 %s, error %s", resGet.JSON500.Status, resGet.JSON500.Error)
	} else if resGet.JSON404 != nil {
		log.Info("Collection does not exist, creating ...")
	} else if resGet.JSON200 != nil {
		log.WithField("collection", *resGet.JSON200).Info("Using existing collection")
		return nil
	} else {
		return fmt.Errorf("could not read default collection, %s, error %s", resGet.Status(), string(resGet.Body))
	}

	if createRequest == nil {
		log.Info("Creating empty collection")
		createRequest = &vaultclient.CollectionCreateRequest{}
	}

	resCreate, err := client.CollectionCreateWithResponse(ctx, ledger, collection, *createRequest)
	if err != nil {
		return fmt.Errorf("could not create collection, %w", err)
	} else if resCreate.JSON400 != nil {
		return fmt.Errorf("could not create collection, %s, error %s", resCreate.JSON400.Status, resCreate.JSON400.Error)
	} else if resCreate.JSON402 != nil {
		return fmt.Errorf("could not create collection, %s, error %s", resCreate.JSON402.Status, resCreate.JSON402.Error)
	} else if resCreate.JSON403 != nil {
		return fmt.Errorf("could not create collection, %s, error %s", resCreate.JSON403.Status, resCreate.JSON403.Error)
	} else if resCreate.JSON409 != nil {
		return fmt.Errorf("could not create collection, %s, error %s", resCreate.JSON409.Status, resCreate.JSON409.Error)
	} else if resCreate.JSON500 != nil {
		return fmt.Errorf("could not create collection, %s, error %s", resCreate.JSON500.Status, resCreate.JSON500.Error)
	}

	log.Info("Collection created")

	return nil
}
