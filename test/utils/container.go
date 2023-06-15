package utils

import (
	"context"
	"fmt"
	"log"
	"time"

	immuHttp "github.com/codenotary/immudb-log-audit/pkg/client/immudb"
	immuCliHttp "github.com/codenotary/immudb/pkg/api/httpclient"
	immudb "github.com/codenotary/immudb/pkg/client"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"

	"github.com/phayes/freeport"
)

func RunImmudbContainer() (immudb.ImmuClient, *immuHttp.HTTPClient, string) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	ports, err := freeport.GetFreePorts(2)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "codenotary/immudb:dev",
		Tty:   false,
		ExposedPorts: nat.PortSet{
			nat.Port("3322/tcp"): {},
			nat.Port("8080/tcp"): {},
		},
	}, &container.HostConfig{
		PortBindings: nat.PortMap{
			nat.Port("3322/tcp"): []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: fmt.Sprint(ports[0])}},
			nat.Port("8080/tcp"): []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: fmt.Sprint(ports[1])}},
		},
	}, nil, nil, "")
	if err != nil {
		log.Fatal(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		log.Fatal(err)
	}

	opts := immudb.DefaultOptions().WithPort(ports[0])
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

	ihc, err := immuCliHttp.NewClientWithResponses(fmt.Sprintf("http://%s:%d/api/v2", ic.GetOptions().Address, ports[1]))
	if err != nil {
		log.Fatal(err)
	}

	ih, err := immuHttp.NewHTTPClient(ctx, ihc, "defaultdb", "immudb", "immudb")
	if err != nil {
		log.Fatal(err)
	}

	return ic, ih, resp.ID
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
