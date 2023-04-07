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
