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

package immudb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/codenotary/immudb/pkg/api/schema"
	immudb "github.com/codenotary/immudb/pkg/client"
	"github.com/tidwall/gjson"
)

type JsonKVRepository struct {
	client      immudb.ImmuClient
	collection  string
	indexedKeys []string // first key is considered primary key
}

func NewJsonKVRepository(cli immudb.ImmuClient, collection string) (*JsonKVRepository, error) {
	if collection == "" {
		return nil, errors.New("collection cannot be empty")
	}

	// read collection definition
	b, err := NewConfigs(cli).ReadConfig(collection)
	if err != nil {
		return nil, fmt.Errorf("collection is missing definition, %w", err)
	}

	var indexes []string
	err = json.Unmarshal(b, &indexes)
	if err != nil {
		return nil, fmt.Errorf("invalid collection configuration: %w", err)
	}

	log.WithField("indexes", indexes).Info("Indexes from immudb")

	return &JsonKVRepository{
		client:      cli,
		collection:  collection,
		indexedKeys: indexes,
	}, nil
}

func SetupJsonKVRepository(cli immudb.ImmuClient, collection string, indexedKeys []string) error {
	b, err := json.Marshal(indexedKeys)
	if err != nil {
		return fmt.Errorf("could not marshal indexes definition, %w", err)
	}

	_, err = cli.Set(context.TODO(), []byte(fmt.Sprintf("%s.collection", collection)), b)
	if err != nil {
		return fmt.Errorf("could not store indexes definition, %w", err)
	}

	cfgs := NewConfigs(cli)
	err = cfgs.WriteConfig(collection, b)
	if err != nil {
		return fmt.Errorf("could not store collection config, %w", err)
	}

	log.WithField("collection", collection).WithField("indexes", indexedKeys).Info("Created")

	return nil
}

func (jr *JsonKVRepository) Write(jObject interface{}) (uint64, error) {
	objectBytes, err := json.Marshal(jObject)
	if err != nil {
		return 0, fmt.Errorf("could not marshal object: %w", err)
	}

	return jr.WriteBytes([][]byte{objectBytes})
}

// Writes json bytes as key-values in immudb, with help of gjson to extract
// json fields as indexes.
//
// Underlying key structure is as follow:
//
// Primary index: <collection>.<primary field name>.{<primary field value as text>}
// Additional indexes: <collection>.<indexed field name>.{<indexed field value as text>}.{<primary field value as text>}
// Original json as bytes: <collection>.payload.<primary field name>.{<primary field value as text>}
//
// Indexes values contain the name of payload key

func (jr *JsonKVRepository) WriteBytes(jBytesArr [][]byte) (uint64, error) {
	if len(jr.indexedKeys) == 0 {
		return 0, errors.New("primary key is mandataory")
	}

	var txID uint64
	for _, jBytes := range jBytesArr {
		// parse with gjson
		gjsonObject := gjson.ParseBytes(jBytes)

		// resolve primary key, format "key1+key2+..."
		var pks []string
		for _, pkPart := range strings.Split(jr.indexedKeys[0], "+") {
			gjPK := gjsonObject.Get(pkPart)
			if !gjPK.Exists() {
				return 0, fmt.Errorf("missing primary key in json, %s", pkPart)
			}
			pks = append(pks, gjPK.String())
		}

		pk := strings.Join(pks, "_")

		immudbObjectRequest := &schema.SetRequest{
			KVs: []*schema.KeyValue{
				{ // crete primary key index
					Key:   []byte(fmt.Sprintf("%s.%s.{%s}", jr.collection, jr.indexedKeys[0], pk)),
					Value: []byte(fmt.Sprintf("%s.payload.%s.{%s}", jr.collection, jr.indexedKeys[0], pk)), //value is link to payload
				},
				{ // create payload entry
					Key:   []byte(fmt.Sprintf("%s.payload.%s.{%s}", jr.collection, jr.indexedKeys[0], pk)),
					Value: jBytes,
				},
			},
		}

		for i := 1; i < len(jr.indexedKeys); i++ {
			gjSK := gjsonObject.Get(jr.indexedKeys[i])
			if !gjSK.Exists() {
				//	return 0, errors.New("missing secondary key in json")
				continue
			}

			immudbObjectRequest.KVs = append(immudbObjectRequest.KVs,
				&schema.KeyValue{ // crete secondary key index <collection>.<SKName>.<SKVALUE>.<PKVALUE>
					Key:   []byte(fmt.Sprintf("%s.%s.{%s}.{%s}", jr.collection, jr.indexedKeys[i], gjSK.String(), pk)),
					Value: []byte([]byte(fmt.Sprintf("%s.payload.%s.{%s}", jr.collection, jr.indexedKeys[0], pk))), //value is link to payload
				},
			)
		}

		txh, err := jr.client.SetAll(context.TODO(), immudbObjectRequest)
		if err != nil {
			return 0, fmt.Errorf("could not store object: %w", err)
		}

		log.WithField("txID", txh.Id).Trace("Wrote entry")
		txID = txh.Id
	}

	return txID, nil
}

// for now just based on SK
func (jr *JsonKVRepository) Read(key string, prefix string) ([][]byte, error) {
	if key == "" {
		key = jr.indexedKeys[0]
	}

	validKey := false
	for _, s := range jr.indexedKeys {
		if s == key {
			validKey = true
		}
	}
	if !validKey {
		return nil, fmt.Errorf("not indexed key %s", key)
	}

	seekKey := []byte("")
	var objects [][]byte
	for {
		entries, err := jr.client.Scan(context.TODO(), &schema.ScanRequest{
			Prefix:  []byte(fmt.Sprintf("%s.%s.{%s", jr.collection, key, prefix)),
			SeekKey: seekKey,
			Limit:   999,
		})
		if err != nil {
			return nil, fmt.Errorf("could not scan for objects, %w", err)
		}

		if len(entries.Entries) == 0 {
			log.WithField("key", key).WithField("prefix", prefix).Debug("No more entries matching condition")
			break
		}

		for _, e := range entries.Entries {
			// retrieve an object
			objectEntry, err := jr.client.Get(context.Background(), e.Value)
			if err != nil {
				return nil, fmt.Errorf("could not scan for object, %w", err)
			}

			seekKey = e.Key
			// filter out possible old entries by secondary index
			if e.Tx == objectEntry.Tx {
				objects = append(objects, objectEntry.Value)
			}
		}
	}

	return objects, nil
}

type History struct {
	Entry    []byte
	TxID     uint64
	Revision uint64
}

func (imo *JsonKVRepository) History(primaryKeyValue string) ([]History, error) {
	offset := uint64(0)
	objects := []History{}
	for {
		entries, err := imo.client.History(context.TODO(), &schema.HistoryRequest{
			Key:    []byte(fmt.Sprintf("%s.payload.%s.{%s}", imo.collection, imo.indexedKeys[0], primaryKeyValue)),
			Offset: offset,
			Limit:  999,
		})

		if err != nil {
			return nil, err
		}

		for _, e := range entries.Entries {
			objects = append(objects, History{
				Entry:    e.Value,
				Revision: e.Revision,
				TxID:     e.Tx,
			})
			offset++
		}

		if len(entries.Entries) < 999 {
			log.WithField("key", primaryKeyValue).Debug("No more history entries")
			break
		}
	}

	return objects, nil
}
