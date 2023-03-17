/*
Copyright 2022 Codenotary Inc. All rights reserved.

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
	"fmt"
	"io"

	"github.com/hpcloud/tail"
)

type FileTail struct {
	t *tail.Tail
}

func NewFileTail(path string, follow bool) (*FileTail, error) {
	t, err := tail.TailFile(path, tail.Config{Follow: follow})
	if err != nil {
		return nil, fmt.Errorf("could not create file tail: %w", err)
	}

	return &FileTail{t: t}, nil
}

func (ft *FileTail) ReadLine() (string, error) {
	l, ok := <-ft.t.Lines
	if l != nil && l.Err != nil {
		return "", fmt.Errorf("error reading line from tail: %w", l.Err)
	}

	if !ok {
		return "", io.EOF
	}

	return l.Text, nil
}
