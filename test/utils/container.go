package utils

import (
	"context"
	"fmt"
	"log"
	"time"

	immudb "github.com/codenotary/immudb/pkg/client"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"

	"github.com/phayes/freeport"
)

func RunImmudbContainer() (immudb.ImmuClient, string) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	port, err := freeport.GetFreePort()
	if err != nil {
		log.Fatal(err)
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "codenotary/immudb-dev1",
		Tty:   false,
		ExposedPorts: nat.PortSet{
			nat.Port("3322/tcp"): {},
		},
	}, &container.HostConfig{
		PortBindings: nat.PortMap{
			nat.Port("3322/tcp"): []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: fmt.Sprint(port)}},
		},
	}, nil, nil, "")
	if err != nil {
		log.Fatal(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		log.Fatal(err)
	}

	opts := immudb.DefaultOptions().WithPort(port)
	ic := immudb.NewClient().WithOptions(opts)
	i := 0
	for {
		err = ic.OpenSession(context.TODO(), []byte("immudb"), []byte("immudb"), "defaultdb")
		if err == nil {
			break
		}

		if i > 100 {
			cli.ContainerStop(ctx, resp.ID, container.StopOptions{})
			cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{
				RemoveVolumes: true,
				Force:         true,
			})

			log.Fatal(err)
		}
		time.Sleep(20 * time.Millisecond)
		i++
	}

	return ic, resp.ID
}

func StopImmudbContainer(containerID string) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()
	ctx := context.Background()

	cli.ContainerStop(ctx, containerID, container.StopOptions{})
	cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	})
}
