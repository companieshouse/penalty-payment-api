//coverage:ignore file

package testutils

import (
	"context"
	"fmt"
	"io"

	"github.com/companieshouse/chs.go/log"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
)

type StoppableContainer interface {
	Start()
	Execute(cmd []string) (int, io.Reader, error)
	GetHost() string
	GetPort() string
	GetZkHost() string
	GetZkPort() string
	Stop()
	GetBrokerAddress() string
}

type standardContainer struct {
	req       testcontainers.ContainerRequest
	container testcontainers.Container
	port      string
	host      string
}

func (c *standardContainer) Start() {
	containerInstance, err := testcontainers.GenericContainer(context.TODO(), testcontainers.GenericContainerRequest{
		ContainerRequest: c.req,
		Started:          true,
	})
	if err != nil {
		panic(err)
	}
	host, err := containerInstance.Host(context.TODO())
	if err != nil {
		panic(err)
	}
	port, err := containerInstance.MappedPort(context.TODO(), nat.Port(c.req.ExposedPorts[0]))
	if err != nil {
		panic(err)
	}
	c.container = containerInstance
	c.port = port.Port()
	c.host = host
}

func (c *standardContainer) GetPort() string {
	return c.port
}

func (c *standardContainer) GetHost() string {
	return c.host
}

func (c *standardContainer) GetZkPort() string {
	return c.port
}

func (c *standardContainer) GetZkHost() string {
	return c.host
}

func (c *standardContainer) Stop() {
	if err := c.container.Terminate(context.TODO()); err != nil {
		log.Error(err)
	}
}

func (c *standardContainer) GetBrokerAddress() string {
	return fmt.Sprintf("%s:%s", c.host, c.port)
}

func (c *standardContainer) Execute(cmd []string) (int, io.Reader, error) {
	return c.container.Exec(context.TODO(), cmd)
}

type containerAggregator struct {
	containers []StoppableContainer
}

func AggregateContainers(containers ...StoppableContainer) *containerAggregator {
	return &containerAggregator{containers: containers}
}

func (d *containerAggregator) Start() {
	for _, c := range d.containers {
		c.Start()
	}
}

func (d *containerAggregator) Stop() {
	for _, c := range d.containers {
		c.Stop()
	}
}

func createNetwork(name string) testcontainers.Network {
	network, err := testcontainers.GenericNetwork(context.TODO(), testcontainers.GenericNetworkRequest{
		NetworkRequest: testcontainers.NetworkRequest{
			Name: name,
		},
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to create network: %v", err))
	} else {
		return network
	}
}
