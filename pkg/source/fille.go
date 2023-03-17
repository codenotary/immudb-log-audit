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
