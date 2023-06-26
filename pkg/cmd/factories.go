package cmd

import (
	"fmt"

	"github.com/codenotary/immudb-log-audit/pkg/lineparser"
	"github.com/codenotary/immudb-log-audit/pkg/service"
)

func NewLineParser(name string) (service.LineParser, error) {
	var lp service.LineParser
	switch name {
	case "":
		lp = lineparser.NewDefaultLineParser()
	case "pgaudit":
		lp = lineparser.NewPGAuditLineParser()
	case "pgauditjsonlog":
		lp = lineparser.NewPGAuditJSONLogLineParser()
	case "wrap":
		lp = lineparser.NewWrapLineParser()
	default:
		return nil, fmt.Errorf("not supported parser: %s", name)
	}

	return lp, nil
}
