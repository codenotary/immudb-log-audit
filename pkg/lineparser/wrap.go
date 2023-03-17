package lineparser

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type wrap struct {
	Uid     string    `json:"uid"`
	Ts      time.Time `json:"log_timestamp"`
	Message string    `json:"message"`
}

type WrapLineParser struct {
}

func NewWrapLineParser() *WrapLineParser {
	return &WrapLineParser{}
}

func (*WrapLineParser) Parse(line string) ([]byte, error) {
	w := wrap{
		Uid:     uuid.New().String(),
		Ts:      time.Now(),
		Message: line,
	}

	return json.Marshal(w)
}
