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

package lineparser

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type wrap struct {
	UID     string    `json:"uid"`
	Ts      time.Time `json:"log_timestamp"`
	Message string    `json:"message"`
}

type wrapLineParser struct {
}

func NewWrapLineParser() *wrapLineParser {
	return &wrapLineParser{}
}

func (*wrapLineParser) Parse(line string) ([]byte, error) {
	w := wrap{
		UID:     uuid.New().String(),
		Ts:      time.Now(),
		Message: line,
	}

	return json.Marshal(w)
}
