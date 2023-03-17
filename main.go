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
