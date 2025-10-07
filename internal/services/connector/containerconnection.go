package connector

import (
	"context"
	"net"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
	"github.com/pkg/errors"
)

type ContainerConnection struct {
	cli         *client.Client
	containerID string
	attachID    string
	conn        net.Conn
}

func (c *ContainerConnection) ResizeTTY(ctx context.Context, width, height uint) error {
	if err := c.cli.ContainerResize(ctx, c.containerID, container.ResizeOptions{
		Height: height,
		Width:  width,
	}); err != nil {
		return errors.Wrap(err, "connector.ContainerConnection#ResizeTTY: failed to resize")
	}
	return nil
}

func (c *ContainerConnection) ContainerID() string {
	return c.containerID
}

func (c *ContainerConnection) Conn() net.Conn {
	return c.conn
}

func (c *ContainerConnection) Close() error {
	err := c.conn.Close()
	if err != nil {
		return errors.Wrap(err, "connector.ContainerConnection#Close: failed to close connection")
	}
	return nil
}
