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
	"fmt"

	immudb "github.com/codenotary/immudb/pkg/client"
)

type configs struct {
	cli immudb.ImmuClient
}

func NewConfigs(cli immudb.ImmuClient) *configs {
	return &configs{
		cli: cli,
	}
}

func (c *configs) ReadConfig(collection string) ([]byte, error) {
	entry, err := c.cli.Get(context.TODO(), []byte(fmt.Sprintf("%s.config", collection)))
	if err != nil {
		return nil, err
	}

	return entry.Value, nil
}

func (c *configs) ReadTypeParser(collection string) (string, string, error) {
	entry, err := c.cli.Get(context.TODO(), []byte(fmt.Sprintf("%s.config.parser", collection)))
	if err != nil {
		return "", "", err
	}

	parser := string(entry.Value)

	entry, err = c.cli.Get(context.TODO(), []byte(fmt.Sprintf("%s.config.type", collection)))
	if err != nil {
		return "", "", err
	}

	typ := string(entry.Value)
	return typ, parser, nil
}

func (c *configs) WriteConfig(collection string, b []byte) error {
	_, err := c.cli.Set(context.TODO(), []byte(fmt.Sprintf("%s.config", collection)), b)
	if err != nil {
		return err
	}

	return nil
}

func (c *configs) WriteTypeParser(collection string, t string, p string) error {
	_, err := c.cli.Set(context.TODO(), []byte(fmt.Sprintf("%s.config.parser", collection)), []byte(p))
	if err != nil {
		return err
	}

	_, err = c.cli.Set(context.TODO(), []byte(fmt.Sprintf("%s.config.type", collection)), []byte(t))
	if err != nil {
		return err
	}

	return nil
}
