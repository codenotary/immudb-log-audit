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

package source

import (
	"bufio"
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type dockerTail struct {
	reader  io.ReadCloser
	scanner *bufio.Scanner
}

func NewDockerTail(container string, follow bool, since string, showStdout bool, showStderr bool) (*dockerTail, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("could not create docker client, %w", err)
	}

	cli.NegotiateAPIVersion(context.TODO())

	reader, err := cli.ContainerLogs(context.TODO(), container, types.ContainerLogsOptions{Follow: follow, Since: since, ShowStdout: showStdout, ShowStderr: showStderr})
	if err != nil {
		return nil, fmt.Errorf("could not create docker logs reader: %w", err)
	}
	scanner := bufio.NewScanner(reader)

	return &dockerTail{
		reader:  reader,
		scanner: scanner,
	}, nil
}

func (dt *dockerTail) ReadLine() (string, error) {
	if dt.scanner.Scan() {
		b := dt.scanner.Bytes()

		if len(b) == 0 {
			return "", nil
		}

		if len(b) > 8 && (b[0] == 1 || b[0] == 2) {
			return string(b[8:]), nil
		}

		return string(b), nil
	}

	return "", io.EOF
}
