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
	ctx     context.Context
	lC      chan string
}

func NewDockerTail(ctx context.Context, container string, follow bool, since string, showStdout bool, showStderr bool) (*dockerTail, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("could not create docker client, %w", err)
	}

	cli.NegotiateAPIVersion(ctx)

	reader, err := cli.ContainerLogs(ctx, container, types.ContainerLogsOptions{Follow: follow, Since: since, ShowStdout: showStdout, ShowStderr: showStderr})
	if err != nil {
		return nil, fmt.Errorf("could not create docker logs reader: %w", err)
	}
	scanner := bufio.NewScanner(reader)

	dt := &dockerTail{
		reader:  reader,
		scanner: scanner,
		ctx:     ctx,
		lC:      make(chan string),
	}

	go dt.read()
	return dt, nil
}

func (dt *dockerTail) read() {
	for stop := false; !stop && dt.scanner.Scan(); {
		b := dt.scanner.Bytes()
		if len(b) == 0 {
			continue
		}

		var s string
		if len(b) > 8 && (b[0] == 1 || b[0] == 2) {
			s = string(b[8:])
		} else {
			s = string(b)
		}

		select {
		case dt.lC <- s:
		case <-dt.ctx.Done():
			stop = true
		}
	}

	close(dt.lC)
}

func (*dockerTail) SaveState() {
	// noop
}

func (dt *dockerTail) ReadLine() chan string {
	return dt.lC
}
