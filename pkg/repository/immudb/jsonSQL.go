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
	"github.com/tidwall/gjson"

	immudb "github.com/codenotary/immudb/pkg/client"
)

type sqlcolumn struct {
	Name    string
	CType   string
	Primary bool
}

type JsonSQLRepository struct {
	client     immudb.ImmuClient
	collection string
	columns    []sqlcolumn
}

func NewJsonSQLRepository(cli immudb.ImmuClient, collection string) (*JsonSQLRepository, error) {
	// retrieve collection table and columns
	tx, err := cli.NewTx(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("could not create transaction for sql repository, %w", err)
	}

	res, err := tx.SQLQuery(context.TODO(), fmt.Sprintf("select name from Tables() where name like '%s';", collection), nil)
	if err != nil {
		return nil, fmt.Errorf("could not query tables, %w", err)
	}

	if len(res.Rows) != 1 {
		return nil, errors.New("collection does not exist")
	}

	b, err := NewConfigs(cli).ReadConfig(collection)
	if err != nil {
		return nil, fmt.Errorf("collection is missing definition, %w", err)
	}

	var columns []sqlcolumn
	err = json.Unmarshal(b, &columns)
	if err != nil {
		return nil, fmt.Errorf("could not read collection config: %w", err)
	}

	_, err = tx.Commit(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("could not initialize sql repository, %w", err)
	}

	log.WithField("columns", columns).Info("Columns from immudb")
	return &JsonSQLRepository{
		client:     cli,
		collection: collection,
		columns:    columns,
	}, nil
}

func (jr *JsonSQLRepository) Write(jObject interface{}) (uint64, error) {
	objectBytes, err := json.Marshal(jObject)
	if err != nil {
		return 0, fmt.Errorf("could not marshal object: %w", err)
	}

	return jr.WriteBytes([][]byte{objectBytes})
}

func (jr *JsonSQLRepository) WriteBytes(jBytesArr [][]byte) (uint64, error) {
	var txID uint64
	for _, jBytes := range jBytesArr {
		// parse with gjson
		gjsonObject := gjson.ParseBytes(jBytes)

		params := map[string]interface{}{"__value__": jBytes}
		cSlice := []string{}
		command := "UPSERT"
		for _, c := range jr.columns {
			if c.Name == "__value__" {
				continue
			}

			if c.CType == "INTEGER AUTO_INCREMENT" {
				command = "INSERT"
				continue
			}

			cSlice = append(cSlice, c.Name)
			gjr := gjsonObject.Get(c.Name)
			if c.Primary && !gjr.Exists() {
				return 0, fmt.Errorf("missing field %s in object", c.Name)
			}

			if c.CType == "INTEGER" {
				params[c.Name] = gjr.Int()
			} else if strings.HasPrefix(c.CType, "VARCHAR") {
				params[c.Name] = gjr.String()
			} else if c.CType == "TIMESTAMP" {
				params[c.Name] = gjr.Time()
			} else if c.CType == "BOOLEAN" {
				params[c.Name] = gjr.Bool()
			} else if c.CType == "FLOAT" {
				params[c.Name] = gjr.Float()
			} else {
				return 0, fmt.Errorf("unsupported field type %s", c.CType)
			}
		}

		sb := strings.Builder{}
		sb.WriteString(command)
		sb.WriteString(" INTO ")
		sb.WriteString(jr.collection)
		sb.WriteString(" (\"")
		sb.WriteString(strings.Join(cSlice, "\",\""))
		sb.WriteString("\", \"__value__\") VALUES (@")
		sb.WriteString(strings.Join(cSlice, ",@"))
		sb.WriteString(",@__value__);")
		log.WithField("sql", sb.String()).WithField("collection", jr.collection).Trace("Inserting row")
		res, err := jr.client.SQLExec(context.TODO(), sb.String(), params)
		if err != nil {
			return 0, fmt.Errorf("could not insert into collection, %w", err)
		}

		txID = res.Txs[0].Header.Id
	}

	return txID, nil
}

func (jr *JsonSQLRepository) Read(query string) ([][]byte, error) {
	// intentionally accepting query as is for now.
	sb := strings.Builder{}
	sb.WriteString("SELECT \"")
	sb.WriteString(jr.columns[0].Name)
	sb.WriteString("\",__value__ FROM ")
	sb.WriteString(jr.collection)
	if query != "" {
		sb.WriteString(" WHERE ")
		sb.WriteString(query)
	}

	page := fmt.Sprintf(" ORDER BY \"%s\" DESC LIMIT 999", jr.columns[0].Name)
	ret := [][]byte{}
	for {
		log.WithField("sql", sb.String()+page).WithField("collection", jr.collection).Info("Reading")
		res, err := jr.client.SQLQuery(context.TODO(), sb.String()+page, nil, true)
		if err != nil {
			return nil, err
		}

		for _, r := range res.Rows {
			ret = append(ret, r.Values[1].GetBs())
		}

		if len(res.Rows) < 999 {
			log.WithField("rows_count", len(res.Rows)).WithField("rows_total", len(ret)).Trace("No more pages")
			break
		}

		if jr.columns[0].CType == "INTEGER" || jr.columns[0].CType == "INTEGER AUTO_INCREMENT" {
			page = fmt.Sprintf(" \"%s\" < %d ORDER BY \"%s\" DESC LIMIT 999;", jr.columns[0].Name, res.Rows[len(res.Rows)-1].Values[0].GetN(), jr.columns[0].Name)
		} else if strings.HasPrefix(jr.columns[0].CType, "VARCHAR") {
			page = fmt.Sprintf(" \"%s\" < '%s' ORDER BY \"%s\" DESC LIMIT 999;", jr.columns[0].Name, res.Rows[len(res.Rows)-1].Values[0].GetS(), jr.columns[0].Name)
		} else if jr.columns[0].CType == "TIMESTAMP" {
			page = fmt.Sprintf(" \"%s\" < %d ORDER BY \"%s\" DESC LIMIT 999;", jr.columns[0].Name, res.Rows[len(res.Rows)-1].Values[0].GetTs(), jr.columns[0].Name)
		} else if jr.columns[0].CType == "BOOLEAN" {
			page = fmt.Sprintf(" \"%s\" < %t ORDER BY \"%s\" DESC LIMIT 999;", jr.columns[0].Name, res.Rows[len(res.Rows)-1].Values[0].GetB(), jr.columns[0].Name)
		} else if jr.columns[0].CType == "FLOAT" {
			page = fmt.Sprintf(" \"%s\" < %f ORDER BY \"%s\" DESC LIMIT 999;", jr.columns[0].Name, res.Rows[len(res.Rows)-1].Values[0].GetF(), jr.columns[0].Name)
		} else {
			return nil, fmt.Errorf("unsupported field type %s", jr.columns[0].CType)
		}

		if !strings.Contains(strings.ToLower(sb.String()), "where") {
			page = " WHERE " + page
		} else {
			page = " AND " + page
		}
	}

	return ret, nil
}

func (jr *JsonSQLRepository) History(query string) ([][]byte, error) {
	// intentionally accepting query as is for now.
	sb := strings.Builder{}
	sb.WriteString("SELECT \"")
	sb.WriteString(jr.columns[0].Name)
	sb.WriteString("\",__value__ FROM ")
	sb.WriteString(jr.collection)
	sb.WriteString(" ")
	if query != "" {
		sb.WriteString(query)
	} else {
		sb.WriteString("SINCE TX 1 UNTIL NOW() ")
	}

	h := [][]byte{}

	page := fmt.Sprintf(" ORDER BY \"%s\" DESC LIMIT 999", jr.columns[0].Name)
	for {
		log.WithField("sql", sb.String()+page).WithField("collection", jr.collection).Info("history")
		res, err := jr.client.SQLQuery(context.TODO(), sb.String()+page, nil, true)
		if err != nil {
			return nil, err
		}

		if len(res.Rows) < 999 {
			break
		}

		for _, r := range res.Rows {
			if err != nil {
				return nil, fmt.Errorf("error querying for row TX")
			}

			h = append(h, r.Values[1].GetBs())
		}

		if jr.columns[0].CType == "INTEGER" {
			page = fmt.Sprintf(" \"%s\" < %d ORDER BY \"%s\" DESC LIMIT 999;", jr.columns[0].Name, res.Rows[len(res.Rows)-1].Values[0].GetN(), jr.columns[0].Name)
		} else if strings.HasPrefix(jr.columns[0].CType, "VARCHAR") {
			page = fmt.Sprintf(" \"%s\" < '%s' ORDER BY \"%s\" DESC LIMIT 999;", jr.columns[0].Name, res.Rows[len(res.Rows)-1].Values[0].GetS(), jr.columns[0].Name)
		} else if jr.columns[0].CType == "TIMESTAMP" {
			page = fmt.Sprintf(" \"%s\" < %d ORDER BY \"%s\" DESC LIMIT 999;", jr.columns[0].Name, res.Rows[len(res.Rows)-1].Values[0].GetTs(), jr.columns[0].Name)
		} else if jr.columns[0].CType == "BOOLEAN" {
			page = fmt.Sprintf(" \"%s\" < %t ORDER BY \"%s\" DESC LIMIT 999;", jr.columns[0].Name, res.Rows[len(res.Rows)-1].Values[0].GetB(), jr.columns[0].Name)
		} else if jr.columns[0].CType == "FLOAT" {
			page = fmt.Sprintf(" \"%s\" < %f ORDER BY \"%s\" DESC LIMIT 999;", jr.columns[0].Name, res.Rows[len(res.Rows)-1].Values[0].GetF(), jr.columns[0].Name)
		} else {
			return nil, fmt.Errorf("unsupported field type %s", jr.columns[0].CType)
		}

		if !strings.Contains(strings.ToLower(sb.String()), "where") {
			page = " WHERE " + page
		}

	}
	return h, nil
}

func SetupJsonSQLRepository(cli immudb.ImmuClient, collection string, primaryKey string, columns []string) error {
	if collection == "" {
		return errors.New("collection cannot be empty")
	}

	columnsCfg := []sqlcolumn{}
	for _, columnStr := range columns {
		splitted := strings.Split(columnStr, "=")
		if len(splitted) != 2 {
			return fmt.Errorf("invalid index definition, %s", columnStr)
		}
		columnsCfg = append(columnsCfg, sqlcolumn{Name: splitted[0], CType: splitted[1],
			Primary: func() bool {
				for _, pk := range strings.Split(primaryKey, ",") {
					if pk == splitted[0] {
						return true
					}
				}
				return false
			}()})
	}

	// create table representing audit log
	tx, err := cli.NewTx(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	sb := strings.Builder{}
	sb.WriteString("CREATE TABLE IF NOT EXISTS ")
	sb.WriteString(collection)
	sb.WriteString(" ( ")
	indexes := []string{}
	for _, columnCfg := range columnsCfg {
		sb.WriteString("\"")
		sb.WriteString(columnCfg.Name)
		sb.WriteString("\"")
		sb.WriteString(" ")
		sb.WriteString(columnCfg.CType)
		sb.WriteString(",")
		indexes = append(indexes, columnCfg.Name)
	}
	sb.WriteString(" __value__ BLOB, PRIMARY KEY (")
	sb.WriteString(primaryKey)
	sb.WriteString("));")

	log.WithField("sql", sb.String()).Info("Creating collection table")
	err = tx.SQLExec(context.TODO(), sb.String(), nil)

	if err != nil {
		log.Fatal(err)
	}

	sb = strings.Builder{}
	sb.WriteString("CREATE INDEX IF NOT EXISTS ON ")
	sb.WriteString(collection)
	sb.WriteString("(\"")
	sb.WriteString(strings.Join(indexes, "\",\""))
	sb.WriteString("\");")

	log.WithField("sql", sb.String()).Info("Creating indexes")
	err = tx.SQLExec(context.TODO(), sb.String(), nil)
	if err != nil {
		log.Fatal(err)
	}

	_, err = tx.Commit(context.TODO())
	if err != nil {
		return err
	}

	b, err := json.Marshal(columnsCfg)
	if err != nil {
		return fmt.Errorf("could not store collection config: %w", err)
	}

	cfgs := NewConfigs(cli)
	err = cfgs.WriteConfig(collection, b)
	if err != nil {
		return fmt.Errorf("could not store collection config, %w", err)
	}

	return nil
}
