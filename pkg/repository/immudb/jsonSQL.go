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

type column struct {
	name  string
	cType string
}

type JsonSQLRepository struct {
	client     immudb.ImmuClient
	collection string
	columns    []column
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

	cfg, err := NewConfigs(cli).Read(collection)
	if err != nil {
		return nil, fmt.Errorf("collection is missing definition, %w", err)
	}

	columns := []column{}
	for _, c := range cfg.Indexes {
		splitted := strings.Split(c, "=")
		if len(splitted) != 2 {
			return nil, fmt.Errorf("invalid index definition, %s", c)
		}

		columns = append(columns, column{name: splitted[0], cType: splitted[1]})
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

	return jr.WriteBytes(objectBytes)
}

func (jr *JsonSQLRepository) WriteBytes(jBytes []byte) (uint64, error) {
	// parse with gjson
	gjsonObject := gjson.ParseBytes(jBytes)

	params := map[string]interface{}{"__value__": jBytes}
	cSlice := []string{}
	for _, c := range jr.columns {
		if c.name == "__value__" {
			continue
		}
		cSlice = append(cSlice, c.name)
		gjr := gjsonObject.Get(c.name)
		if !gjr.Exists() {
			return 0, fmt.Errorf("missing field %s in object", c)
		}

		if c.cType == "INTEGER" {
			params[c.name] = gjr.Int()
		} else if strings.HasPrefix(c.cType, "VARCHAR") {
			params[c.name] = gjr.String()
		} else if c.cType == "TIMESTAMP" {
			params[c.name] = gjr.Time()
		} else {
			return 0, fmt.Errorf("unsupported field type %s", c.cType)
		}
	}

	sb := strings.Builder{}
	sb.WriteString("UPSERT INTO ")
	sb.WriteString(jr.collection)
	sb.WriteString(" (\"")
	sb.WriteString(strings.Join(cSlice, "\",\""))
	sb.WriteString("\", \"__value__\") VALUES (@")
	sb.WriteString(strings.Join(cSlice, ",@"))
	sb.WriteString(",@__value__);")
	log.WithField("sql", sb.String()).WithField("collection", jr.collection).Trace("inserting row")
	res, err := jr.client.SQLExec(context.TODO(), sb.String(), params)
	if err != nil {
		return 0, fmt.Errorf("could not insert into collection, %w", err)
	}

	return res.Txs[0].Header.Id, nil
}

func (jr *JsonSQLRepository) Read(query string) ([][]byte, error) {
	// intentionally accepting query as is for now.
	sb := strings.Builder{}
	sb.WriteString("SELECT \"")
	sb.WriteString(jr.columns[0].name)
	sb.WriteString("\",__value__ FROM ")
	sb.WriteString(jr.collection)
	if query != "" {
		sb.WriteString(" WHERE ")
		sb.WriteString(query)
	}

	page := fmt.Sprintf(" ORDER BY \"%s\" DESC LIMIT 999", jr.columns[0].name)
	ret := [][]byte{}
	for {
		log.WithField("sql", sb.String()+page).WithField("collection", jr.collection).Info("reading")
		res, err := jr.client.SQLQuery(context.TODO(), sb.String()+page, nil, true)
		if err != nil {
			return nil, err
		}

		if len(res.Rows) < 999 {
			break
		}

		for _, r := range res.Rows {
			ret = append(ret, r.Values[0].GetBs())
		}

		if jr.columns[0].cType == "INTEGER" {
			page = fmt.Sprintf(" \"%s\" < %d ORDER BY \"%s\" DESC LIMIT 999;", jr.columns[0].name, res.Rows[len(res.Rows)-1].Values[0].GetN(), jr.columns[0].name)
		} else if strings.HasPrefix(jr.columns[0].cType, "VARCHAR") {
			page = fmt.Sprintf(" \"%s\" < '%s' ORDER BY \"%s\" DESC LIMIT 999;", jr.columns[0].name, res.Rows[len(res.Rows)-1].Values[0].GetS(), jr.columns[0].name)
		} else if jr.columns[0].cType == "TIMESTAMP" {
			page = fmt.Sprintf(" \"%s\" < %d ORDER BY \"%s\" DESC LIMIT 999;", jr.columns[0].name, res.Rows[len(res.Rows)-1].Values[0].GetTs(), jr.columns[0].name)
		} else {
			return nil, fmt.Errorf("unsupported field type %s", jr.columns[0].cType)
		}

		if !strings.Contains(strings.ToLower(sb.String()), "where") {
			page = " WHERE " + page
		}
	}

	return ret, nil
}

func (jr *JsonSQLRepository) History(query string) ([][]byte, error) {
	// intentionally accepting query as is for now.
	sb := strings.Builder{}
	sb.WriteString("SELECT \"")
	sb.WriteString(jr.columns[0].name)
	sb.WriteString("\",__value__ FROM ")
	sb.WriteString(jr.collection)
	sb.WriteString(" ")
	if query != "" {
		sb.WriteString(query)
	} else {
		sb.WriteString("SINCE TX 1 UNTIL NOW() ")
	}

	h := [][]byte{}

	page := fmt.Sprintf(" ORDER BY \"%s\" DESC LIMIT 999", jr.columns[0].name)
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

		if jr.columns[0].cType == "INTEGER" {
			page = fmt.Sprintf(" \"%s\" < %d ORDER BY \"%s\" DESC LIMIT 999;", jr.columns[0].name, res.Rows[len(res.Rows)-1].Values[0].GetN(), jr.columns[0].name)
		} else if strings.HasPrefix(jr.columns[0].cType, "VARCHAR") {
			page = fmt.Sprintf(" \"%s\" < '%s' ORDER BY \"%s\" DESC LIMIT 999;", jr.columns[0].name, res.Rows[len(res.Rows)-1].Values[0].GetS(), jr.columns[0].name)
		} else if jr.columns[0].cType == "TIMESTAMP" {
			page = fmt.Sprintf(" \"%s\" < %d ORDER BY \"%s\" DESC LIMIT 999;", jr.columns[0].name, res.Rows[len(res.Rows)-1].Values[0].GetTs(), jr.columns[0].name)
		} else {
			return nil, fmt.Errorf("unsupported field type %s", jr.columns[0].cType)
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
	for _, column := range columns {
		splitted := strings.Split(column, "=")
		if len(splitted) != 2 {
			return fmt.Errorf("invalid index definition, %s", column)
		}
		sb.WriteString("\"")
		sb.WriteString(splitted[0])
		sb.WriteString("\"")
		sb.WriteString(" ")
		sb.WriteString(splitted[1])
		sb.WriteString(",")
		indexes = append(indexes, splitted[0])
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

	return nil
}
