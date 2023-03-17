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

package main

import (
	"flag"

	"github.com/codenotary/immudb-log-audit/cmd"
)

var flagQueryOnly bool
var flagAuditTrailJson bool
var flagPgauditTrail bool
var flagFollow bool

func init() {
	flag.BoolVar(&flagQueryOnly, "query-only", false, "if True, do not save into immudb")
	flag.BoolVar(&flagAuditTrailJson, "audit-trail-json", false, "if True, run AuditTrailJson")
	flag.BoolVar(&flagPgauditTrail, "pgaudit", false, "if True, run pgaudit")
	flag.BoolVar(&flagFollow, "follow", false, "if True, run pgaudit")
	flag.Parse()
}

func main() {
	cmd.Execute()
	return
}
