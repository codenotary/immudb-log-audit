package immudb

import (
	"context"
	"encoding/json"
	"fmt"

	immudb "github.com/codenotary/immudb/pkg/client"
)

type Config struct {
	Parser  string
	Type    string
	Indexes []string
}

type configs struct {
	cli immudb.ImmuClient
}

func NewConfigs(cli immudb.ImmuClient) *configs {
	return &configs{
		cli: cli,
	}
}

func (c *configs) Read(collection string) (*Config, error) {
	entry, err := c.cli.Get(context.TODO(), []byte(fmt.Sprintf("%s.config", collection)))
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = json.Unmarshal(entry.Value, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *configs) Write(collection string, cfg Config) error {
	b, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	_, err = c.cli.Set(context.TODO(), []byte(fmt.Sprintf("%s.config", collection)), b)
	if err != nil {
		return err
	}

	return nil
}
