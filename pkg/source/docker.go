package source

import (
	"bufio"
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type DockerTail struct {
	reader  io.ReadCloser
	scanner *bufio.Scanner
}

func NewDockerTail(container string, follow bool, since string, showStdout bool, showStderr bool) (*DockerTail, error) {
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

	return &DockerTail{
		reader:  reader,
		scanner: scanner,
	}, nil
}

func (dt *DockerTail) ReadLine() (string, error) {
	if dt.scanner.Scan() {
		b := dt.scanner.Bytes()
		if b[0] == 1 || b[0] == 2 {
			return string(b[8:]), nil
		}

		return string(b), nil
	}

	return "", io.EOF
}
