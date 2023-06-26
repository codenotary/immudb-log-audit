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
	"net/http"
	"strconv"

	vaultclient "github.com/codenotary/immudb-log-audit/pkg/client/vault"
	log "github.com/sirupsen/logrus"
)

type JsonVaultRepository struct {
	client     vaultclient.ClientWithResponsesInterface
	ledger     string
	collection string
	batchMode  bool
}

func NewJsonVaultRepository(client vaultclient.ClientWithResponsesInterface, ledger string, collection string, batchMode bool) (*JsonVaultRepository, error) {
	return &JsonVaultRepository{
		client:     client,
		ledger:     ledger,
		collection: collection,
		batchMode:  batchMode,
	}, nil
}

func (jv *JsonVaultRepository) WriteBytes(jBytes [][]byte) (uint64, error) {
	ctx := context.Background()
	var txID uint64
	if jv.batchMode {
		//TODO: add ticker
		docBuf := []byte(`{"documents": [`)
		docBufCounter := 0
		for i := 0; i < len(jBytes); i++ {
			if docBufCounter > 0 {
				docBuf = append(docBuf, ',')
			}

			docBuf = append(docBuf, jBytes[i]...)
			docBufCounter++
			if docBufCounter < 100 && i != len(jBytes)-1 {
				continue
			}

			docBuf = append(docBuf, []byte("]}")...)
			res, err := jv.client.DocumentCreateManyWithBodyWithResponse(ctx, jv.ledger, jv.collection, "application/json", bytes.NewBuffer(docBuf))
			docBuf = []byte(`{"documents": [`)
			docBufCounter = 0
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
		}
	} else {
		for i := 0; i < len(jBytes); i++ {
			log.WithField("line", string(jBytes[i])).Debug("Writing line")
			res, err := jv.client.DocumentCreateWithBodyWithResponse(ctx, jv.ledger, jv.collection, "application/json", bytes.NewReader(jBytes[i]))
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
		query := vaultclient.Query{}
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
	}

	if resGet.JSON200 != nil {
		log.WithField("collection", *resGet.JSON200).Info("Using existing collection")
		return nil
	} else if resGet.JSON404 != nil {
		log.WithField("collection", collection).Info("Collection does not exist, creating ...")
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
	}

	if resCreate.StatusCode() != http.StatusOK {
		return fmt.Errorf("could not create collection, %s, error %s", resGet.Status(), string(resGet.Body))
	}

	log.WithField("collection", collection).Info("Collection created")

	return nil
}
